/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Flag] custom HTTP and HTTPS ports", func() {

	host := "forwarded-headers"

	f := framework.NewDefaultFramework("forwarded-port-headers", framework.WithHTTPBunEnabled())

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()

		f.WaitForNginxServer("_",
			func(server string) bool {
				return strings.Contains(server, "listen 1443")
			})
	})

	ginkgo.Context("with a plain HTTP ingress", func() {
		ginkgo.It("should set X-Forwarded-Port headers accordingly when listening on a non-default HTTP port", func() {

			ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "server_name forwarded-headers")
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Body().
				Contains(fmt.Sprintf("x-forwarded-port=%d", 1080))
		})
	})

	ginkgo.Context("with a TLS enabled ingress", func() {

		ginkgo.It("should set X-Forwarded-Port header to 443", func() {

			ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "server_name forwarded-headers")
				})

			f.HTTPTestClientWithTLSConfig(tlsConfig).
				GET("/").
				WithURL(f.GetURL(framework.HTTPS)).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Body().
				Contains("x-forwarded-port=443")
		})

		ginkgo.Context("when external authentication is configured", func() {

			ginkgo.It("should set the X-Forwarded-Port header to 443", func() {
				annotations := map[string]string{
					"nginx.ingress.kubernetes.io/auth-url":    fmt.Sprintf("http://%s/basic-auth/user/password", f.HTTPBunIP),
					"nginx.ingress.kubernetes.io/auth-signin": "http://$host/auth/start",
				}

				ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations)

				f.EnsureIngress(ing)

				tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
					ing.Spec.TLS[0].Hosts,
					ing.Spec.TLS[0].SecretName,
					ing.Namespace)
				assert.Nil(ginkgo.GinkgoT(), err)

				framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

				f.WaitForNginxServer(host,
					func(server string) bool {
						return strings.Contains(server, "server_name forwarded-headers")
					})

				f.HTTPTestClientWithTLSConfig(tlsConfig).
					GET("/").
					WithURL(f.GetURL(framework.HTTPS)).
					WithHeader("Host", host).
					WithBasicAuth("user", "password").
					Expect().
					Status(http.StatusOK).
					Body().
					Contains("x-forwarded-port=443")
			})
		})
	})
})
