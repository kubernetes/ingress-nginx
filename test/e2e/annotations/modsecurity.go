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
	"strings"

	. "github.com/onsi/ginkgo"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - ModSecurityLocation", func() {
	f := framework.NewDefaultFramework("modsecuritylocation")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should enable modsecurity", func() {
		host := "modsecurity.foo.com"
		nameSpace := f.IngressController.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;")
			})
	})

	It("should enable modsecurity with transaction ID and OWASP rules", func() {
		host := "modsecurity.foo.com"
		nameSpace := f.IngressController.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity":         "true",
			"nginx.ingress.kubernetes.io/enable-owasp-core-rules":    "true",
			"nginx.ingress.kubernetes.io/modsecurity-transaction-id": "modsecurity-$request_id",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "modsecurity_rules_file /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf;") &&
					strings.Contains(server, "modsecurity_transaction_id \"modsecurity-$request_id\";")
			})
	})

	It("should disable modsecurity", func() {
		host := "modsecurity.foo.com"
		nameSpace := f.IngressController.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity": "false",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "modsecurity on;")
			})
	})

	It("should enable modsecurity with snippet", func() {
		host := "modsecurity.foo.com"
		nameSpace := f.IngressController.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-modsecurity":  "true",
			"nginx.ingress.kubernetes.io/modsecurity-snippet": "SecRuleEngine On",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "modsecurity on;") &&
					strings.Contains(server, "SecRuleEngine On")
			})
	})
})
