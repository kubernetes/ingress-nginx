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
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("[SSL] TLS protocols, ciphers and headers)", func() {
	f := framework.NewDefaultFramework("settings-tls")
	host := "settings-tls"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.Context("should configure TLS protocol", func() {
		var (
			sslCiphers   string
			sslProtocols string
			testCiphers  string
			tlsConfig    *tls.Config
		)

		ginkgo.BeforeEach(func() {
			sslCiphers = "ssl-ciphers"
			sslProtocols = "ssl-protocols"

			// Two ciphers supported by each of TLSv1.2 and TLSv1.
			// https://www.openssl.org/docs/man1.1.0/apps/ciphers.html - "CIPHER SUITE NAMES"
			testCiphers = "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA"

			ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))
			tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)
		})

		ginkgo.It("setting cipher suite", func() {
			f.SetNginxConfigMapData(map[string]string{
				sslCiphers:   testCiphers,
				sslProtocols: "TLSv1.2",
			})

			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, fmt.Sprintf("ssl_ciphers '%s';", testCiphers))
				})

			resp := f.HTTPTestClientWithTLSConfig(tlsConfig).
				GET("/").
				WithURL(f.GetURL(framework.HTTPS)).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Raw()

			assert.Equal(ginkgo.GinkgoT(), int(resp.TLS.Version), tls.VersionTLS12)
			assert.Equal(ginkgo.GinkgoT(), resp.TLS.CipherSuite, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		})
		ginkgo.It("enforcing TLS v1.0", func() {
			f.SetNginxConfigMapData(map[string]string{
				sslCiphers:   testCiphers,
				sslProtocols: "TLSv1",
			})

			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "ssl_protocols TLSv1;")
				})

			resp := f.HTTPTestClientWithTLSConfig(tlsConfig).
				GET("/").
				WithURL(f.GetURL(framework.HTTPS)).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Raw()

			assert.Equal(ginkgo.GinkgoT(), int(resp.TLS.Version), tls.VersionTLS10)
			assert.Equal(ginkgo.GinkgoT(), resp.TLS.CipherSuite, tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA)
		})
	})

	ginkgo.Context("should configure HSTS policy header", func() {
		var (
			tlsConfig *tls.Config
		)

		const (
			hstsMaxAge            = "hsts-max-age"
			hstsIncludeSubdomains = "hsts-include-subdomains"
			hstsPreload           = "hsts-preload"
		)

		ginkgo.BeforeEach(func() {
			ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))
			tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)
		})

		ginkgo.It("setting max-age parameter", func() {
			f.UpdateNginxConfigMapData(hstsMaxAge, "86400")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`hsts_max_age = 86400,`))
			})

			f.HTTPTestClientWithTLSConfig(tlsConfig).
				GET("/").
				WithURL(f.GetURL(framework.HTTPS)).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Header("Strict-Transport-Security").Equal("max-age=86400; includeSubDomains")
		})

		ginkgo.It("setting includeSubDomains parameter", func() {
			f.SetNginxConfigMapData(map[string]string{
				hstsMaxAge:            "86400",
				hstsIncludeSubdomains: "false",
			})

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`hsts_include_subdomains = false,`))
			})

			f.HTTPTestClientWithTLSConfig(tlsConfig).
				GET("/").
				WithURL(f.GetURL(framework.HTTPS)).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Header("Strict-Transport-Security").Equal("max-age=86400")
		})

		ginkgo.It("setting preload parameter", func() {
			f.SetNginxConfigMapData(map[string]string{
				hstsMaxAge:            "86400",
				hstsPreload:           "true",
				hstsIncludeSubdomains: "false",
			})

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`hsts_preload = true,`))
			})

			f.HTTPTestClientWithTLSConfig(tlsConfig).
				GET("/").
				WithURL(f.GetURL(framework.HTTPS)).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Header("Strict-Transport-Security").Equal("max-age=86400; preload")
		})

		ginkgo.It("overriding what's set from the upstream", func() {
			f.SetNginxConfigMapData(map[string]string{
				hstsMaxAge:            "86400",
				hstsPreload:           "true",
				hstsIncludeSubdomains: "false",
			})

			// we can not use gorequest here because it flattens the duplicate headers
			// and specifically in case of Strict-Transport-Security it ignore extra headers
			// intead of concatenating, rightfully. And I don't know of any API it provides for getting raw headers.
			curlCmd := fmt.Sprintf("curl -I -k --fail --silent --resolve settings-tls:443:127.0.0.1 https://settings-tls%v", "?hsts=true")
			output, err := f.ExecIngressPod(curlCmd)
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.Contains(ginkgo.GinkgoT(), output, "strict-transport-security: max-age=86400; preload")
			// this is what the upstream sets
			assert.NotContains(ginkgo.GinkgoT(), output, "strict-transport-security: max-age=3600; preload")
		})

	})

	ginkgo.Context("ports or X-Forwarded-Host check during HTTP tp HTTPS redirection", func() {
		ginkgo.It("should not use ports during the HTTP to HTTPS redirection", func() {
			ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))
			tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusPermanentRedirect).
				Header("Location").Equal(fmt.Sprintf("https://%v", host))
		})

		ginkgo.It("should not use ports or X-Forwarded-Host during the HTTP to HTTPS redirection", func() {
			f.UpdateNginxConfigMapData("use-forwarded-headers", "true")

			ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))
			tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				WithHeader("X-Forwarded-Host", "example.com:80").
				Expect().
				Status(http.StatusPermanentRedirect).
				Header("Location").Equal("https://example.com")
		})
	})
})
