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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/ingress-nginx/test/e2e/framework"
	"regexp"
	"strings"
)

var _ = framework.IngressNginxDescribe("Annotations - IPWhiteList", func() {
	f := framework.NewDefaultFramework("ipwhitelist")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(2)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should set valid ip whitelist range", func() {
		host := "ipwhitelist.foo.com"
		nameSpace := f.IngressController.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/whitelist-source-range": "18.0.0.0/8, 56.0.0.0/8",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		denyRegex := regexp.MustCompile("geo \\$the_real_ip \\$deny_[A-Za-z]{32}")
		denyString := ""

		err = f.WaitForNginxConfiguration(
			func(conf string) bool {

				match := denyRegex.FindStringSubmatch(conf)
				// If no match found, return false
				if !(len(match) > 0) {
					return false
				}

				denyString = strings.Replace(match[0], "geo $the_real_ip ", "", -1)
				return strings.Contains(conf, match[0])
			})
		Expect(err).NotTo(HaveOccurred())

		ipOne := "18.0.0.0/8 0;"
		ipTwo := "56.0.0.0/8 0;"

		err = f.WaitForNginxConfiguration(
			func(conf string) bool {
				return strings.Contains(conf, ipOne) && strings.Contains(conf, ipTwo)
			})
		Expect(err).NotTo(HaveOccurred())

		denyStatement := "if (" + denyString + ")"
		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, denyStatement)
			})
		Expect(err).NotTo(HaveOccurred())
	})
})
