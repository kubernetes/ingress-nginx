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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	waitForLuaSync = 5 * time.Second
)

var _ = framework.IngressNginxDescribe("Annotations - canary", func() {
	f := framework.NewDefaultFramework("canary")

	BeforeEach(func() {
		// Deployment for main backend
		f.NewEchoDeployment()

		// Deployment for canary backend
		f.NewDeployment("http-svc-canary", "gcr.io/kubernetes-e2e-test-images/echoserver:2.1", 8080, 1)
	})

	Context("when canaried by header", func() {
		It("should route requests to the canary pod if header is set to 'always'", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo")) &&
						Expect(server).ShouldNot(ContainSubstring("return 503"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).ShouldNot(Equal(http.StatusNotFound))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})

		It("should not route requests to the canary pod if header is set to 'never'", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo")) &&
						Expect(server).ShouldNot(ContainSubstring("return 503"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).ShouldNot(Equal(http.StatusNotFound))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))
		})
	})

	Context("when canaried by cookie", func() {
		It("should route requests to the canary pod if cookie is set to 'always'", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo")) &&
						Expect(server).ShouldNot(ContainSubstring("return 503"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-cookie": "CanaryByCookie",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				AddCookie(&http.Cookie{Name: "CanaryByCookie", Value: "always"}).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).ShouldNot(Equal(http.StatusNotFound))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})

		It("should not route requests to the canary pod if cookie is set to 'never'", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			Expect(ing).NotTo(BeNil())

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo")) &&
						Expect(server).ShouldNot(ContainSubstring("return 503"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				AddCookie(&http.Cookie{Name: "CanaryByCookie", Value: "never"}).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).ShouldNot(Equal(http.StatusNotFound))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))
		})
	})
})
