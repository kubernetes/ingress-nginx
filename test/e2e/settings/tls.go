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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

func noRedirectPolicyFunc(gorequest.Request, []gorequest.Request) error {
	return http.ErrUseLastResponse
}

var _ = framework.IngressNginxDescribe("Settings - TLS)", func() {
	f := framework.NewDefaultFramework("settings-tls")
	host := "settings-tls"

	BeforeEach(func() {
		f.NewEchoDeployment()
		f.UpdateNginxConfigMapData("use-forwarded-headers", "false")
	})

	AfterEach(func() {
	})

	It("should configure TLS protocol", func() {
		sslCiphers := "ssl-ciphers"
		sslProtocols := "ssl-protocols"

		// Two ciphers supported by each of TLSv1.2 and TLSv1.
		// https://www.openssl.org/docs/man1.1.0/apps/ciphers.html - "CIPHER SUITE NAMES"
		testCiphers := "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA"

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "http-svc", 80, nil))
		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).NotTo(HaveOccurred())

		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

		By("setting cipher suite")
		f.UpdateNginxConfigMapData(sslCiphers, testCiphers)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf("ssl_ciphers '%s';", testCiphers))
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTPS)).
			TLSClientConfig(tlsConfig).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.TLS.Version).Should(BeNumerically("==", tls.VersionTLS12))
		Expect(resp.TLS.CipherSuite).Should(BeNumerically("==", tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384))

		By("enforcing TLS v1.0")
		f.UpdateNginxConfigMapData(sslProtocols, "TLSv1")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "ssl_protocols TLSv1;")
			})

		resp, _, errs = gorequest.New().
			Get(f.GetURL(framework.HTTPS)).
			TLSClientConfig(tlsConfig).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.TLS.Version).Should(BeNumerically("==", tls.VersionTLS10))
		Expect(resp.TLS.CipherSuite).Should(BeNumerically("==", tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA))
	})

	It("should configure HSTS policy header", func() {
		hstsMaxAge := "hsts-max-age"
		hstsIncludeSubdomains := "hsts-include-subdomains"
		hstsPreload := "hsts-preload"

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "http-svc", 80, nil))
		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).NotTo(HaveOccurred())

		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

		By("setting max-age parameter")
		f.UpdateNginxConfigMapData(hstsMaxAge, "86400")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "Strict-Transport-Security: max-age=86400; includeSubDomains\"")
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTPS)).
			TLSClientConfig(tlsConfig).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Strict-Transport-Security")).Should(ContainSubstring("max-age=86400"))

		By("setting includeSubDomains parameter")
		f.UpdateNginxConfigMapData(hstsIncludeSubdomains, "false")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "Strict-Transport-Security: max-age=86400\"")
			})

		resp, _, errs = gorequest.New().
			Get(f.GetURL(framework.HTTPS)).
			TLSClientConfig(tlsConfig).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Strict-Transport-Security")).ShouldNot(ContainSubstring("includeSubDomains"))

		By("setting preload parameter")
		f.UpdateNginxConfigMapData(hstsPreload, "true")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "Strict-Transport-Security: max-age=86400; preload\"")
			})

		resp, _, errs = gorequest.New().
			Get(f.GetURL(framework.HTTPS)).
			TLSClientConfig(tlsConfig).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Strict-Transport-Security")).Should(ContainSubstring("preload"))
	})

	It("should not use ports during the HTTP to HTTPS redirection", func() {
		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "http-svc", 80, nil))
		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).NotTo(HaveOccurred())

		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

		resp, _, errs := gorequest.New().
			Get(fmt.Sprintf(f.GetURL(framework.HTTP))).
			Retry(10, 1*time.Second, http.StatusNotFound).
			RedirectPolicy(noRedirectPolicyFunc).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusPermanentRedirect))
		Expect(resp.Header.Get("Location")).Should(Equal(fmt.Sprintf("https://%v/", host)))
	})

	It("should not use ports or X-Forwarded-Host during the HTTP to HTTPS redirection", func() {
		f.UpdateNginxConfigMapData("use-forwarded-headers", "true")

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "http-svc", 80, nil))
		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).NotTo(HaveOccurred())

		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

		resp, _, errs := gorequest.New().
			Get(fmt.Sprintf(f.GetURL(framework.HTTP))).
			Retry(10, 1*time.Second, http.StatusNotFound).
			RedirectPolicy(noRedirectPolicyFunc).
			Set("Host", host).
			Set("X-Forwarded-Host", "example.com:80").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusPermanentRedirect))
		Expect(resp.Header.Get("Location")).Should(Equal("https://example.com/"))
	})
})
