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

package annotations

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - ProxySSL", func() {
	f := framework.NewDefaultFramework("proxyssl")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(2)
	})

	AfterEach(func() {
	})

	It("should set valid proxy-ssl-secret", func() {
		host := "proxyssl.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-ssl-secret": f.Namespace + "/" + host,
		}

		_, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		Expect(err).ToNot(HaveOccurred())

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, &annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "DEFAULT", "TLSv1 TLSv1.1 TLSv1.2", "off", 1)
	})

	It("should set valid proxy-ssl-secret, proxy-ssl-verify to on, and proxy-ssl-verify-depth to 2", func() {
		host := "proxyssl.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-ssl-secret":       f.Namespace + "/" + host,
			"nginx.ingress.kubernetes.io/proxy-ssl-verify":       "on",
			"nginx.ingress.kubernetes.io/proxy-ssl-verify-depth": "2",
		}

		_, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		Expect(err).ToNot(HaveOccurred())

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, &annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "DEFAULT", "TLSv1 TLSv1.1 TLSv1.2", "on", 2)
	})

	It("should set valid proxy-ssl-secret, proxy-ssl-ciphers to HIGH:!AES", func() {
		host := "proxyssl.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-ssl-secret":  f.Namespace + "/" + host,
			"nginx.ingress.kubernetes.io/proxy-ssl-ciphers": "HIGH:!AES",
		}

		_, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		Expect(err).ToNot(HaveOccurred())

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, &annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "HIGH:!AES", "TLSv1 TLSv1.1 TLSv1.2", "off", 1)
	})

	It("should set valid proxy-ssl-secret, proxy-ssl-protocols", func() {
		host := "proxyssl.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-ssl-secret":    f.Namespace + "/" + host,
			"nginx.ingress.kubernetes.io/proxy-ssl-protocols": "TLSv1.2 TLSv1.3",
		}

		_, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		Expect(err).ToNot(HaveOccurred())

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, &annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "DEFAULT", "TLSv1.2 TLSv1.3", "off", 1)
	})
})

func assertProxySSL(f *framework.Framework, host, ciphers, protocols, verify string, depth int) {
	certFile := fmt.Sprintf("/etc/ingress-controller/ssl/%s-%s.pem", f.Namespace, host)
	f.WaitForNginxServer(host,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("proxy_ssl_certificate %s;", certFile)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_certificate_key %s;", certFile)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_trusted_certificate %s;", certFile)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_ciphers %s;", ciphers)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_protocols %s;", protocols)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_verify %s;", verify)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_verify_depth %d;", depth))
		})
}
