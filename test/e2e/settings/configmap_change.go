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
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Configmap change", func() {
	f := framework.NewDefaultFramework("configmap-change")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should reload after an update in the configuration", func() {
		host := "configmap-change"

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		wlKey := "whitelist-source-range"
		wlValue := "1.1.1.1"

		By("adding a whitelist-source-range")

		err = f.UpdateNginxConfigMapData(wlKey, wlValue)
		Expect(err).NotTo(HaveOccurred())

		checksumRegex := regexp.MustCompile("Configuration checksum:\\s+(\\d+)")
		checksum := ""

		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				// before returning, extract the current checksum
				match := checksumRegex.FindStringSubmatch(cfg)
				if len(match) > 0 {
					checksum = match[1]
				}

				return strings.Contains(cfg, "geo $the_real_ip $deny_") &&
					strings.Contains(cfg, "1.1.1.1 0")
			})
		Expect(err).NotTo(HaveOccurred())
		Expect(checksum).NotTo(BeEmpty())

		By("changing error-log-level")

		err = f.UpdateNginxConfigMapData("error-log-level", "debug")
		Expect(err).NotTo(HaveOccurred())

		newChecksum := ""
		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				match := checksumRegex.FindStringSubmatch(cfg)
				if len(match) > 0 {
					newChecksum = match[1]
				}

				return strings.ContainsAny(cfg, "error_log  /var/log/nginx/error.log debug;")
			})
		Expect(err).NotTo(HaveOccurred())

		Expect(checksum).NotTo(BeEquivalentTo(newChecksum))
	})
})
