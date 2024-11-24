/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[TCP] tcp-services", func() {
	f := framework.NewDefaultFramework("tcp")
	if framework.IsCrossplane() {
		return
	}
	var ip string

	ginkgo.BeforeEach(func() {
		ip = f.GetNginxIP()
	})

	ginkgo.It("should expose a TCP service", func() {
		f.NewEchoDeployment()

		cm := f.GetConfigMap(f.Namespace, "tcp-services")
		cm.Data = map[string]string{
			"8080": fmt.Sprintf("%v/%v:80", f.Namespace, framework.EchoService),
		}
		f.EnsureConfigMap(cm)

		svc := f.GetService(f.Namespace, "nginx-ingress-controller")
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       framework.EchoService,
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
		})
		_, err := f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Update(context.TODO(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf(`ngx.var.proxy_upstream_name="tcp-%v-%v-80"`,
					f.Namespace, framework.EchoService))
			})

		f.HTTPTestClient().
			GET("/").
			WithURL(fmt.Sprintf("http://%v:8080", ip)).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should expose an ExternalName TCP service", func() {
		// Setup:
		// - Create an external name service for DNS lookups on port 5353. Point it to google's DNS server
		// - Expose port 5353 on the nginx ingress NodePort service to open a hole for this test
		// - Update the `tcp-services` configmap to proxy traffic to the configured external name service

		// Create an external service for DNS
		externalService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dns-external-name-svc",
				Namespace: f.Namespace,
			},

			Spec: corev1.ServiceSpec{
				ExternalName: "google-public-dns-a.google.com",
				Ports: []corev1.ServicePort{
					{
						Name:       "dns-external-name-svc",
						Port:       5353,
						TargetPort: intstr.FromInt(53),
					},
				},
				Type: corev1.ServiceTypeExternalName,
			},
		}
		f.EnsureService(externalService)

		// Expose the `external name` port on the `ingress-nginx-controller` service
		svc := f.GetService(f.Namespace, "nginx-ingress-controller")
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "dns-svc",
			Port:       5353,
			TargetPort: intstr.FromInt(5353),
		})
		_, err := f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Update(context.TODO(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		// Update the TCP configmap to link port 5353 to the DNS external name service
		config := f.GetConfigMap(f.Namespace, "tcp-services")
		config.Data = map[string]string{
			"5353": fmt.Sprintf("%v/dns-external-name-svc:5353", f.Namespace),
		}
		f.EnsureConfigMap(config)

		// Validate that the generated nginx config contains the expected `proxy_upstream_name` value
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf(`ngx.var.proxy_upstream_name="tcp-%v-dns-external-name-svc-5353"`, f.Namespace))
			})

		// Execute the test. Use the `external name` service to resolve a domain name.
		resolver := net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "tcp", fmt.Sprintf("%v:5353", ip))
			},
		}

		// add retries to LookupHost to avoid random e2e errors
		retry := wait.Backoff{
			Steps:    10,
			Duration: 2 * time.Second,
			Factor:   0.8,
			Jitter:   0.2,
		}

		var ips []string
		var errRetry error
		err = wait.ExponentialBackoff(retry, func() (bool, error) {
			ips, errRetry = resolver.LookupHost(context.Background(), "google-public-dns-b.google.com")
			if errRetry == nil {
				return true, nil
			}

			return false, nil
		})
		//nolint:staticcheck // TODO: will replace it since wait.ErrWaitTimeout is deprecated
		if err == wait.ErrWaitTimeout {
			err = errRetry
		}

		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error from DNS resolver")
		assert.Contains(ginkgo.GinkgoT(), ips, "8.8.4.4")
	})

	ginkgo.It("should reload after an update in the configuration", func() {
		ginkgo.By("setting up a first deployment")
		f.NewEchoDeployment(framework.WithDeploymentName("first-service"))

		cm := f.GetConfigMap(f.Namespace, "tcp-services")
		cm.Data = map[string]string{
			"8080": fmt.Sprintf("%v/first-service:80", f.Namespace),
		}
		f.EnsureConfigMap(cm)

		checksumRegex := regexp.MustCompile(`Configuration checksum:\s+(\d+)`)
		checksum := ""

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				// before returning, extract the current checksum
				match := checksumRegex.FindStringSubmatch(cfg)
				if len(match) > 0 {
					checksum = match[1]
				}

				return strings.Contains(cfg, fmt.Sprintf(`ngx.var.proxy_upstream_name="tcp-%v-first-service-80"`,
					f.Namespace))
			})
		assert.NotEmpty(ginkgo.GinkgoT(), checksum)

		ginkgo.By("updating the tcp service to a second deployment")
		f.NewEchoDeployment(framework.WithDeploymentName("second-service"))

		cm = f.GetConfigMap(f.Namespace, "tcp-services")
		cm.Data["8080"] = fmt.Sprintf("%v/second-service:80", f.Namespace)
		f.EnsureConfigMap(cm)

		newChecksum := ""
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				match := checksumRegex.FindStringSubmatch(cfg)
				if len(match) > 0 {
					newChecksum = match[1]
				}

				return strings.Contains(cfg, fmt.Sprintf(`ngx.var.proxy_upstream_name="tcp-%v-second-service-80"`,
					f.Namespace))
			})
		assert.NotEqual(ginkgo.GinkgoT(), checksum, newChecksum)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, "Backend successfully reloaded")
	})
})
