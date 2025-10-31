/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package settings

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
	"k8s.io/ingress-nginx/test/e2e/framework/httpexpect"
)

var _ = framework.IngressNginxDescribe("[Flag] enable-ssl-passthrough", func() {
	f := framework.NewDefaultFramework("ssl-passthrough", framework.WithHTTPBunEnabled())

	ginkgo.BeforeEach(func() {
		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--enable-ssl-passthrough")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		f.WaitForNginxServer("_",
			func(server string) bool {
				return strings.Contains(server, "listen 442")
			})
	})

	ginkgo.Describe("With enable-ssl-passthrough enabled", func() {
		ginkgo.It("should enable ssl-passthrough-proxy-port on a different port", func() {
			err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, "--ssl-passthrough-proxy-port=1442")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
				return err
			})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

			f.WaitForNginxServer("_",
				func(server string) bool {
					return strings.Contains(server, "listen 1442")
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", "something").
				Expect().
				Status(http.StatusNotFound)
		})

		ginkgo.Context("when handling traffic", func() {
			var tlsConfig *tls.Config
			host := "testpassthrough.com"
			url := "https://" + net.JoinHostPort(host, "443")
			echoName := "echopass"
			secretName := host

			ginkgo.BeforeEach(func() {
				/* Even with enable-ssl-passthrough enabled, only annotated ingresses may receive the traffic */
				annotations := map[string]string{
					"nginx.ingress.kubernetes.io/ssl-passthrough": "true",
				}

				ingressDef := framework.NewSingleIngress(host,
					"/",
					host,
					f.Namespace,
					echoName,
					80,
					annotations)
				var err error
				tlsConfig, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
					[]string{host},
					secretName,
					ingressDef.Namespace)

				volumeMount := []corev1.VolumeMount{
					{
						Name:      "certs",
						ReadOnly:  true,
						MountPath: "/certs",
					},
				}
				volume := []corev1.Volume{
					{
						Name: "certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secretName,
							},
						},
					},
				}
				envs := []corev1.EnvVar{
					{
						Name:  "HTTPBUN_SSL_CERT",
						Value: "/certs/tls.crt",
					},
					{
						Name:  "HTTPBUN_SSL_KEY",
						Value: "/certs/tls.key",
					},
				}

				f.NewDeploymentWithOpts("echopass",
					framework.HTTPBunImage,
					80,
					1,
					nil,
					nil,
					envs,
					volumeMount,
					volume,
					false)

				f.EnsureIngress(ingressDef)

				assert.Nil(ginkgo.GinkgoT(), err)
				framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

				f.WaitForNginxServer(host,
					func(server string) bool {
						return strings.Contains(server, "listen 442")
					})
			})

			ginkgo.It("should pass unknown traffic to default backend and handle known traffic", func() {
				/* This one should not receive traffic as it does not contain passthrough annotation */
				hostBad := "noannotationnopassthrough.com"
				urlBad := "https://" + net.JoinHostPort(hostBad, "443")
				ingBad := f.EnsureIngress(framework.NewSingleIngressWithTLS(hostBad,
					"/",
					hostBad,
					[]string{hostBad},
					f.Namespace,
					echoName,
					80,
					nil))
				tlsConfigBad, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
					ingBad.Spec.TLS[0].Hosts,
					ingBad.Spec.TLS[0].SecretName,
					ingBad.Namespace)
				assert.Nil(ginkgo.GinkgoT(), err)
				framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfigBad)

				f.WaitForNginxServer(hostBad,
					func(server string) bool {
						return strings.Contains(server, "listen 442")
					})

				//nolint:gosec // Ignore the gosec error in testing
				f.HTTPTestClientWithTLSConfig(&tls.Config{ServerName: host, InsecureSkipVerify: true}).
					GET("/").
					WithURL(url).
					ForceResolve(f.GetNginxIP(), 443).
					Expect().
					Status(http.StatusOK)

				//nolint:gosec // Ignore the gosec error in testing
				f.HTTPTestClientWithTLSConfig(&tls.Config{ServerName: hostBad, InsecureSkipVerify: true}).
					GET("/").
					WithURL(urlBad).
					ForceResolve(f.GetNginxIP(), 443).
					Expect().
					Status(http.StatusNotFound)

				f.HTTPTestClientWithTLSConfig(tlsConfig).
					GET("/").
					WithURL(url).
					ForceResolve(f.GetNginxIP(), 443).
					Expect().
					Status(http.StatusOK)

				f.HTTPTestClientWithTLSConfig(tlsConfigBad).
					GET("/").
					WithURL(urlBad).
					ForceResolve(f.GetNginxIP(), 443).
					Expect().
					Status(http.StatusNotFound)
			})

			ginkgo.Context("on throttled connections", func() {
				throttleMiddleware := func(next httpexpect.DialContextFunc) httpexpect.DialContextFunc {
					return func(ctx context.Context, network, addr string) (net.Conn, error) {
						// Wrap the connection with a throttled writer to simulate real
						// world traffic where streaming data may arrive in chunks
						conn, err := next(ctx, network, addr)
						return &writeThrottledConn{
							Conn:      conn,
							chunkSize: len(host) / 3,
						}, err
					}
				}

				ginkgo.It("should handle known traffic without Host header", func() {
					f.HTTPTestClientWithTLSConfig(tlsConfig).
						GET("/").
						WithURL(url).
						ForceResolve(f.GetNginxIP(), 443).
						WithDialContextMiddleware(throttleMiddleware).
						Expect().
						Status(http.StatusOK)
				})

				ginkgo.It("should handle insecure traffic without Host header", func() {
					//nolint:gosec // Ignore the gosec error in testing
					f.HTTPTestClientWithTLSConfig(&tls.Config{ServerName: host, InsecureSkipVerify: true}).
						GET("/").
						WithURL(url).
						ForceResolve(f.GetNginxIP(), 443).
						WithDialContextMiddleware(throttleMiddleware).
						Expect().
						Status(http.StatusOK)
				})

				ginkgo.It("should handle known traffic with Host header", func() {
					f.HTTPTestClientWithTLSConfig(tlsConfig).
						GET("/").
						WithURL(url).
						WithHeader("Host", host).
						ForceResolve(f.GetNginxIP(), 443).
						WithDialContextMiddleware(throttleMiddleware).
						Expect().
						Status(http.StatusOK)
				})

				ginkgo.It("should handle insecure traffic with Host header", func() {
					//nolint:gosec // Ignore the gosec error in testing
					f.HTTPTestClientWithTLSConfig(&tls.Config{ServerName: host, InsecureSkipVerify: true}).
						GET("/").
						WithURL(url).
						WithHeader("Host", host).
						ForceResolve(f.GetNginxIP(), 443).
						WithDialContextMiddleware(throttleMiddleware).
						Expect().
						Status(http.StatusOK)
				})
			})
		})
	})
})

type writeThrottledConn struct {
	net.Conn
	chunkSize int
}

// Write writes data to the connection `chunkSize` bytes (or less) at a time.
func (c *writeThrottledConn) Write(b []byte) (n int, err error) {
	for i := 0; i < len(b); i += c.chunkSize {
		n, err := c.Conn.Write(b[i:min(i+c.chunkSize, len(b))])
		if err != nil {
			return i + n, err
		}
	}
	return len(b), nil
}
