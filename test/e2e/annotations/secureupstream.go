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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/ingress-nginx/test/e2e/framework"
	"strings"
)

var _ = framework.IngressNginxDescribe("Annotations - SecureUpstream", func() {
	f := framework.NewDefaultFramework("secureupstream")

	BeforeEach(func() {
		f.NewEchoDeployment()
		f.NewGRPCFortuneTellerDeployment()
	})

	AfterEach(func() {
	})

	It("should set secure-verify-ca-secret with https backend", func() {
		host := "secureupstream.foo.com"
		nameSpace := f.IngressController.Namespace
		ca := host + "-ca"

		_, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			ca,
			nameSpace,
			true)
		Expect(err).ToNot(HaveOccurred())

		_, err = framework.CreateIngressTLSSecret(
			f.KubeClientSet,
			[]string{host},
			host,
			nameSpace)
		Expect(err).ToNot(HaveOccurred())

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol":        "HTTPS",
			"nginx.ingress.kubernetes.io/secure-verify-ca-secret": ca,
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, nameSpace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		proxySslVerify := "proxy_ssl_verify on;"
		proxySslTrustedCertificate := fmt.Sprintf("proxy_ssl_trusted_certificate /etc/ingress-controller/ssl/ca-%s-%s.pem;", nameSpace, ca)
		proxySslVerifyDepth := "proxy_ssl_verify_depth 2;"

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, proxySslTrustedCertificate) && strings.Contains(server, proxySslVerify) && strings.Contains(server, proxySslVerifyDepth)
			})
	})

	It("should set secure-verify-ca-secret with grpcs backend", func() {
		host := "secureupstream.foo.com"
		nameSpace := f.IngressController.Namespace
		ca := host + "-ca"

		_, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			ca,
			nameSpace,
			true)
		Expect(err).ToNot(HaveOccurred())

		_, err = framework.CreateIngressTLSSecret(
			f.KubeClientSet,
			[]string{host},
			host,
			nameSpace)
		Expect(err).ToNot(HaveOccurred())

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol":        "GRPCS",
			"nginx.ingress.kubernetes.io/secure-verify-ca-secret": ca,
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, nameSpace, "fortune-teller", 80, &annotations)
		f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		proxySslVerify := "proxy_ssl_verify on;"
		proxySslTrustedCertificate := fmt.Sprintf("proxy_ssl_trusted_certificate /etc/ingress-controller/ssl/ca-%s-%s.pem", nameSpace, ca)
		proxySslVerifyDepth := "proxy_ssl_verify_depth 2;"

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, proxySslTrustedCertificate) && strings.Contains(server, proxySslVerify) && strings.Contains(server, proxySslVerifyDepth)
			})
	})
})
