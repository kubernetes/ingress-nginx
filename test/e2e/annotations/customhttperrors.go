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

var _ = framework.IngressNginxDescribe("Annotations - custom-http-errors", func() {
	f := framework.NewDefaultFramework("custom-http-errors")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(2)
	})

	AfterEach(func() {
	})

	It("should set proxy_intercept_errors", func() {
		host := "customerrors.foo.com"

		errorCodes := []string{"404", "500"}

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-http-errors": strings.Join(errorCodes, ","),
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_intercept_errors on;"))
			})
	})

	It("should create error routes", func() {
		host := "customerrors.foo.com"
		errorCodes := []string{"404", "500"}

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-http-errors": strings.Join(errorCodes, ","),
		}

		ing := framework.NewSingleIngress(host, "/test", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		for _, code := range errorCodes {
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(fmt.Sprintf("@custom_%s", code)))
				})
		}
	})

	It("should set up error_page routing", func() {
		host := "customerrors.foo.com"
		errorCodes := []string{"404", "500"}

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-http-errors": strings.Join(errorCodes, ","),
		}

		ing := framework.NewSingleIngress(host, "/test", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		for _, code := range errorCodes {
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(fmt.Sprintf("error_page %s = @custom_%s", code, code)))
				})
		}
	})

	It("should create only one of each error route", func() {
		host := "customerrors.foo.com"
		errorCodes := [][]string{{"404", "500"}, {"400", "404"}}

		for i, codeSet := range errorCodes {
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/custom-http-errors": strings.Join(codeSet, ","),
			}

			ing := framework.NewSingleIngress(
				fmt.Sprintf("%s-%d", host, i), fmt.Sprintf("/test-%d", i),
				host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)
		}

		for _, codeSet := range errorCodes {
			for _, code := range codeSet {
				f.WaitForNginxServer(host,
					func(server string) bool {
						count := strings.Count(server, fmt.Sprintf("location @custom_%s", code))
						return Expect(count).Should(Equal(1))
					})
			}
		}
	})
})
