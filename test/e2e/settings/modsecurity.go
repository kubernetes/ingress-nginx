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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("ModSecurity Global", func() {
	f := framework.NewDefaultFramework("modsecurityglobal")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should enable modsecurity", func() {
		host := "modsecurityglobal.foo.com"

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.UpdateNginxConfigMapData("modsecurity", "true")
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "modsecurity on;") && strings.Contains(cfg, "modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;")
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should enable modsecurity with owasp core rules", func() {
		host := "modsecurityglobal.foo.com"

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.UpdateNginxConfigMapData("modsecurity", "true")
		Expect(err).NotTo(HaveOccurred())

		err = f.UpdateNginxConfigMapData("enable-owasp-modsecurity-crs", "true")
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "modsecurity on;") && strings.Contains(cfg, "modsecurity_rules_file /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf;")
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should enable modsecurity with remote rules", func() {
		host := "modsecurityglobal.foo.com"

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.UpdateNginxConfigMapData("modsecurity", "true")
		Expect(err).NotTo(HaveOccurred())

		err = f.UpdateNginxConfigMapData("remote-modsecurity-rules-key", "secure-key")
		Expect(err).NotTo(HaveOccurred())

		err = f.UpdateNginxConfigMapData("remote-modsecurity-rules-location", "https://secure.com/yeah")
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "modsecurity on;") && strings.Contains(cfg, "modsecurity_rules_remote secure-key https://secure.com/yeah;")
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should enable modsecurity with snippet", func() {
		host := "modsecurityglobal.foo.com"

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.UpdateNginxConfigMapData("modsecurity", "true")
		Expect(err).NotTo(HaveOccurred())

		snippet := "SecRuleEngine On\nSecDebugLog /tmp/modsec_debug.log\nSecDebugLogLevel 9\nSecRuleRemoveById 10"

		// TODO:
		err = f.UpdateNginxConfigMapData("modsecurity-snippet", snippet)
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "modsecurity on;") && strings.Contains(cfg, "modsecurity_rules")
			})
		Expect(err).NotTo(HaveOccurred())
	})
})
