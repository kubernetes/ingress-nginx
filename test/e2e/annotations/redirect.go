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
	"net/http"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

func noRedirectPolicyFunc(gorequest.Request, []gorequest.Request) error {
	return http.ErrUseLastResponse
}

var _ = framework.IngressNginxDescribe("Annotations - Redirect", func() {
	f := framework.NewDefaultFramework("redirect")

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should respond with a standard redirect code", func() {
		By("setting permanent-redirect annotation")

		host := "redirect"
		redirectPath := "/something"
		redirectURL := "http://redirect.example.com"

		annotations := map[string]string{"nginx.ingress.kubernetes.io/permanent-redirect": redirectURL}

		ing := framework.NewSingleIngress(host, redirectPath, host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("if ($uri ~* %s) {", redirectPath)) &&
					strings.Contains(server, fmt.Sprintf("return 301 %s;", redirectURL))
			})

		By("sending request to redirected URL path")

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+redirectPath).
			Set("Host", host).
			RedirectPolicy(noRedirectPolicyFunc).
			End()

		Expect(errs).To(BeNil())
		Expect(resp.StatusCode).Should(BeNumerically("==", http.StatusMovedPermanently))
		Expect(resp.Header.Get("Location")).Should(Equal(redirectURL))
		Expect(body).Should(ContainSubstring("nginx/"))
	})

	It("should respond with a custom redirect code", func() {
		By("setting permanent-redirect-code annotation")

		host := "redirect"
		redirectPath := "/something"
		redirectURL := "http://redirect.example.com"
		redirectCode := http.StatusFound

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/permanent-redirect":      redirectURL,
			"nginx.ingress.kubernetes.io/permanent-redirect-code": strconv.Itoa(redirectCode),
		}

		ing := framework.NewSingleIngress(host, redirectPath, host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("if ($uri ~* %s) {", redirectPath)) &&
					strings.Contains(server, fmt.Sprintf("return %d %s;", redirectCode, redirectURL))
			})

		By("sending request to redirected URL path")

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+redirectPath).
			Set("Host", host).
			RedirectPolicy(noRedirectPolicyFunc).
			End()

		Expect(errs).To(BeNil())
		Expect(resp.StatusCode).Should(BeNumerically("==", redirectCode))
		Expect(resp.Header.Get("Location")).Should(Equal(redirectURL))
		Expect(body).Should(ContainSubstring("nginx/"))
	})
})
