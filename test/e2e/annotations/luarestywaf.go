/*
Copyright 2017 The Kubernetes Authors.

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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - lua-resty-waf", func() {
	f := framework.NewDefaultFramework("luarestywaf")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when lua-resty-waf is enabled", func() {
		It("should return 403 for a malicious request that matches a default WAF rule and 200 for other requests", func() {
			host := "foo"
			createIngress(f, host, map[string]string{"nginx.ingress.kubernetes.io/lua-resty-waf": "active"})

			url := fmt.Sprintf("%s?msg=<A href=\"http://mysite.com/\">XSS</A>", f.IngressController.HTTPURL)
			resp, _, errs := gorequest.New().
				Get(url).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
		It("should not apply ignored rulesets", func() {
			host := "foo"
			createIngress(f, host, map[string]string{
				"nginx.ingress.kubernetes.io/lua-resty-waf":                 "active",
				"nginx.ingress.kubernetes.io/lua-resty-waf-ignore-rulesets": "41000_sqli, 42000_xss"})

			url := fmt.Sprintf("%s?msg=<A href=\"http://mysite.com/\">XSS</A>", f.IngressController.HTTPURL)
			resp, _, errs := gorequest.New().
				Get(url).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
		It("should apply configured extra rules", func() {
			host := "foo"
			createIngress(f, host, map[string]string{
				"nginx.ingress.kubernetes.io/lua-resty-waf": "active",
				"nginx.ingress.kubernetes.io/lua-resty-waf-extra-rules": `[=[
						{ "access": [
								{ "actions": { "disrupt" : "DENY" },
								"id": 10001,
								"msg": "my custom rule",
								"operator": "STR_CONTAINS",
								"pattern": "foo",
								"vars": [ { "parse": [ "values", 1 ], "type": "REQUEST_ARGS" } ] }
							],
							"body_filter": [],
							"header_filter":[]
						}
					]=]`,
			})

			url := fmt.Sprintf("%s?msg=my-message", f.IngressController.HTTPURL)
			resp, _, errs := gorequest.New().
				Get(url).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			url = fmt.Sprintf("%s?msg=my-foo-message", f.IngressController.HTTPURL)
			resp, _, errs = gorequest.New().
				Get(url).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})
	Context("when lua-resty-waf is not enabled", func() {
		It("should return 200 even for a malicious request", func() {
			host := "foo"
			createIngress(f, host, map[string]string{})

			url := fmt.Sprintf("%s?msg=<A href=\"http://mysite.com/\">XSS</A>", f.IngressController.HTTPURL)
			resp, _, errs := gorequest.New().
				Get(url).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
		It("should run in simulate mode", func() {
			host := "foo"
			createIngress(f, host, map[string]string{"nginx.ingress.kubernetes.io/lua-resty-waf": "simulate"})

			url := fmt.Sprintf("%s?msg=<A href=\"http://mysite.com/\">XSS</A>", f.IngressController.HTTPURL)
			resp, _, errs := gorequest.New().
				Get(url).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			time.Sleep(5 * time.Second)
			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).To(ContainSubstring("Request score greater than score threshold"))
		})
	})
})

func createIngress(f *framework.Framework, host string, annotations map[string]string) {
	ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, &annotations))
	Expect(err).NotTo(HaveOccurred())
	Expect(ing).NotTo(BeNil())

	err = f.WaitForNginxServer(host,
		func(server string) bool {
			return Expect(server).Should(ContainSubstring("server_name foo")) &&
				Expect(server).ShouldNot(ContainSubstring("return 503"))
		})
	Expect(err).NotTo(HaveOccurred())

	resp, body, errs := gorequest.New().
		Get(f.IngressController.HTTPURL).
		Set("Host", host).
		End()

	Expect(len(errs)).Should(Equal(0))
	Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", host)))
}
