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
		f.NewDeployment("http-svc-canary", "gcr.io/kubernetes-e2e-test-images/echoserver:2.2", 8080, 1)
	})

	Context("when canary is created", func() {
		It("should response with a 200 status from the mainline upstream when requests are made to the mainline ingress", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
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
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))
		})

		It("should return 404 status for requests to the canary if no matching ingress is found", func() {
			host := "foo"

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)

			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
		})

		/*

			TODO: This test needs improvements made to the e2e framework so that deployment updates work in order to successfully run

			It("should return the correct status codes when endpoints are unavailable", func() {
				host := "foo"
				annotations := map[string]string{}

				ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
				f.EnsureIngress(ing)

				f.WaitForNginxServer(host,
					func(server string) bool {
						return Expect(server).Should(ContainSubstring("server_name foo"))
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

				By("returning a 503 status when the mainline deployment has 0 replicas and a request is sent to the canary")

				f.NewEchoDeploymentWithReplicas(0)

				resp, _, errs := gorequest.New().
					Get(f.IngressController.HTTPURL).
					Set("Host", host).
					Set("CanaryByHeader", "always").
					End()

				Expect(errs).Should(BeEmpty())
				Expect(resp.StatusCode).Should(Equal(http.StatusServiceUnavailable))

				By("returning a 200 status when the canary deployment has 0 replicas and a request is sent to the mainline ingress")

				f.NewEchoDeploymentWithReplicas(1)
				f.NewDeployment("http-svc-canary", "gcr.io/kubernetes-e2e-test-images/echoserver:2.2", 8080, 0)

				resp, _, errs = gorequest.New().
					Get(f.IngressController.HTTPURL).
					Set("Host", host).
					Set("CanaryByHeader", "never").
					End()

				Expect(errs).Should(BeEmpty())
				Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			})
		*/

		It("should route requests to the correct upstream if mainline ingress is created before the canary ingress", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
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

			By("routing requests destined for the mainline ingress to the maineline upstream")
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests destined for the canary ingress to the canary upstream")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})

		It("should route requests to the correct upstream if mainline ingress is created after the canary ingress", func() {
			host := "foo"

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
				})

			By("routing requests destined for the mainline ingress to the mainelin upstream")
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests destined for the canary ingress to the canary upstream")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})

		It("should route requests to the correct upstream if the mainline ingress is modified", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
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

			modAnnotations := map[string]string{
				"foo": "bar",
			}

			modIng := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &modAnnotations)

			f.EnsureIngress(modIng)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
				})

			By("routing requests destined fro the mainline ingress to the mainline upstream")

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests destined for the canary ingress to the canary upstream")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})

		It("should route requests to the correct upstream if the canary ingress is modified", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
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

			modCanaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader2",
			}

			modCanaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary", 80, &modCanaryAnnotations)
			f.EnsureIngress(modCanaryIng)

			time.Sleep(waitForLuaSync)

			By("routing requests destined for the mainline ingress to the mainline upstream")

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader2", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests destined for the canary ingress to the canary upstream")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader2", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})
	})

	Context("when canaried by header with no value", func() {
		It("should route requests to the correct upstream", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
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

			By("routing requests to the canary upstream when header is set to 'always'")

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when header is set to 'never'")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when header is set to anything else")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "badheadervalue").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))
		})
	})

	Context("when canaried by header with value", func() {
		It("should route requests to the correct upstream", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":                 "true",
				"nginx.ingress.kubernetes.io/canary-by-header":       "CanaryByHeader",
				"nginx.ingress.kubernetes.io/canary-by-header-value": "DoCanary",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			By("routing requests to the canary upstream when header is set to 'DoCanary'")

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "DoCanary").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when header is set to 'always'")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "always").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when header is set to 'never'")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "never").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when header is set to anything else")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "otherheadervalue").
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))
		})
	})

	Context("when canaried by header with value and cookie", func() {
		It("should route requests to the correct upstream", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":                 "true",
				"nginx.ingress.kubernetes.io/canary-by-header":       "CanaryByHeader",
				"nginx.ingress.kubernetes.io/canary-by-header-value": "DoCanary",
				"nginx.ingress.kubernetes.io/canary-by-cookie":       "CanaryByCookie",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			By("routing requests to the canary upstream when header value does not match and cookie is set to 'always'")
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				Set("CanaryByHeader", "otherheadervalue").
				AddCookie(&http.Cookie{Name: "CanaryByCookie", Value: "always"}).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))
		})
	})

	Context("when canaried by cookie", func() {
		It("should route requests to the correct upstream", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-cookie": "Canary-By-Cookie",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			By("routing requests to the canary upstream when cookie is set to 'always'")
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				AddCookie(&http.Cookie{Name: "Canary-By-Cookie", Value: "always"}).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when cookie is set to 'never'")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				AddCookie(&http.Cookie{Name: "Canary-By-Cookie", Value: "never"}).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("routing requests to the mainline upstream when cookie is set to anything else")

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				AddCookie(&http.Cookie{Name: "Canary-By-Cookie", Value: "badcookievalue"}).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))
		})
	})

	// TODO: add testing for canary-weight 0 < weight < 100
	Context("when canaried by weight", func() {
		It("should route requests to the correct upstream", func() {
			host := "foo"
			annotations := map[string]string{}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("server_name foo"))
				})

			canaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":        "true",
				"nginx.ingress.kubernetes.io/canary-weight": "0",
			}

			canaryIngName := fmt.Sprintf("%v-canary", host)

			canaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary",
				80, &canaryAnnotations)
			f.EnsureIngress(canaryIng)

			time.Sleep(waitForLuaSync)

			By("returning requests from the mainline only when weight is equal to 0")

			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc"))
			Expect(body).ShouldNot(ContainSubstring("http-svc-canary"))

			By("returning requests from the canary only when weight is equal to 100")

			modCanaryAnnotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":        "true",
				"nginx.ingress.kubernetes.io/canary-weight": "100",
			}

			modCanaryIng := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary", 80, &modCanaryAnnotations)

			f.EnsureIngress(modCanaryIng)

			time.Sleep(waitForLuaSync)

			resp, body, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring("http-svc-canary"))

		})
	})

	Context("Single canary Ingress", func() {
		It("should not use canary as a catch-all server", func() {
			host := "foo"
			canaryIngName := fmt.Sprintf("%v-canary", host)
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			ing := framework.NewSingleCatchAllIngress(canaryIngName, f.IngressController.Namespace, "http-svc-canary", 80, &annotations)
			f.EnsureIngress(ing)

			ing = framework.NewSingleCatchAllIngress(host, f.IngressController.Namespace, "http-svc", 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxServer("_",
				func(server string) bool {
					upstreamName := fmt.Sprintf(`set $proxy_upstream_name "%s-%s-%s";`, f.IngressController.Namespace, "http-svc", "80")
					canaryUpstreamName := fmt.Sprintf(`set $proxy_upstream_name "%s-%s-%s";`, f.IngressController.Namespace, "http-svc-canary", "80")
					return Expect(server).Should(ContainSubstring(`set $ingress_name "`+host+`";`)) &&
						Expect(server).ShouldNot(ContainSubstring(`set $proxy_upstream_name "upstream-default-backend";`)) &&
						Expect(server).ShouldNot(ContainSubstring(canaryUpstreamName)) &&
						Expect(server).Should(ContainSubstring(upstreamName))
				})
		})

		It("should not use canary with domain as a server", func() {
			host := "foo"
			canaryIngName := fmt.Sprintf("%v-canary", host)
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/canary":           "true",
				"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
			}

			ing := framework.NewSingleIngress(canaryIngName, "/", host, f.IngressController.Namespace, "http-svc-canary", 80, &annotations)
			f.EnsureIngress(ing)

			otherHost := "bar"
			ing = framework.NewSingleIngress(otherHost, "/", otherHost, f.IngressController.Namespace, "http-svc", 80, nil)
			f.EnsureIngress(ing)

			time.Sleep(waitForLuaSync)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return Expect(cfg).Should(ContainSubstring("server_name "+otherHost)) &&
					Expect(cfg).ShouldNot(ContainSubstring("server_name "+host))
			})
		})
	})
})
