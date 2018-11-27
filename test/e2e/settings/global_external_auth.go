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
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Global External Auth", func() {
	f := framework.NewDefaultFramework("global-external-auth")

	host := "global-external-auth"

	echoServiceName := "http-svc"

	globalExternalAuthURLSetting := "global-auth-url"

	fooPath := "/foo"
	barPath := "/bar"

	noAuthSetting := "no-auth-locations"
	noAuthLocations := barPath

	enableGlobalExternalAuthAnnotation := "nginx.ingress.kubernetes.io/enable-global-auth"

	BeforeEach(func() {
		f.NewEchoDeployment()
		f.NewHttpbinDeployment()
	})

	AfterEach(func() {
	})

	Context("when global external authentication is configured", func() {

		BeforeEach(func() {
			globalExternalAuthURL := fmt.Sprintf("http://httpbin.%s.svc.cluster.local:80/status/401", f.Namespace)

			By("Adding an ingress rule for /foo")
			fooIng := framework.NewSingleIngress("foo-ingress", fooPath, host, f.Namespace, echoServiceName, 80, nil)
			f.EnsureIngress(fooIng)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("location /foo"))
				})

			By("Adding an ingress rule for /bar")
			barIng := framework.NewSingleIngress("bar-ingress", barPath, host, f.Namespace, echoServiceName, 80, nil)
			f.EnsureIngress(barIng)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("location /bar"))
				})

			By("Adding a global-auth-url to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthURLSetting, globalExternalAuthURL)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(globalExternalAuthURL))
				})
		})

		It("should return status code 401 when request any protected service", func() {

			By("Sending a request to protected service /foo")
			fooResp, _, _ := gorequest.New().
				Get(f.GetURL(framework.HTTP)+fooPath).
				Set("Host", host).
				End()
			Expect(fooResp.StatusCode).Should(Equal(http.StatusUnauthorized))

			By("Sending a request to protected service /bar")
			barResp, _, _ := gorequest.New().
				Get(f.GetURL(framework.HTTP)+barPath).
				Set("Host", host).
				End()
			Expect(barResp.StatusCode).Should(Equal(http.StatusUnauthorized))
		})

		It("should return status code 200 when request whitelisted (via no-auth-locations) service and 401 when request protected service", func() {

			By("Adding a no-auth-locations for /bar to configMap")
			f.UpdateNginxConfigMapData(noAuthSetting, noAuthLocations)

			By("Sending a request to protected service /foo")
			fooResp, _, _ := gorequest.New().
				Get(f.GetURL(framework.HTTP)+fooPath).
				Set("Host", host).
				End()
			Expect(fooResp.StatusCode).Should(Equal(http.StatusUnauthorized))

			By("Sending a request to whitelisted service /bar")
			barResp, _, _ := gorequest.New().
				Get(f.GetURL(framework.HTTP)+barPath).
				Set("Host", host).
				End()
			Expect(barResp.StatusCode).Should(Equal(http.StatusOK))
		})

		It("should return status code 200 when request whitelisted (via ingress annotation) service and 401 when request protected service", func() {

			By("Adding an ingress rule for /bar with annotation enable-global-auth = false")
			annotations := map[string]string{
				enableGlobalExternalAuthAnnotation: "false",
			}
			barIng := framework.NewSingleIngress("bar-ingress", barPath, host, f.Namespace, echoServiceName, 80, &annotations)
			f.EnsureIngress(barIng)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("location /bar"))
				})

			By("Sending a request to protected service /foo")
			fooResp, _, _ := gorequest.New().
				Get(f.GetURL(framework.HTTP)+fooPath).
				Set("Host", host).
				End()
			Expect(fooResp.StatusCode).Should(Equal(http.StatusUnauthorized))

			By("Sending a request to whitelisted service /bar")
			barResp, _, _ := gorequest.New().
				Get(f.GetURL(framework.HTTP)+barPath).
				Set("Host", host).
				End()
			Expect(barResp.StatusCode).Should(Equal(http.StatusOK))
		})

		It(`should proxy_method method when global-auth-method is configured`, func() {

			globalExternalAuthMethodSetting := "global-auth-method"
			globalExternalAuthMethod := "GET"

			By("Adding a global-auth-method to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthMethodSetting, globalExternalAuthMethod)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("proxy_method"))
				})
		})

		It(`should add custom error page when global-auth-signin url is configured`, func() {

			globalExternalAuthSigninSetting := "global-auth-signin"
			globalExternalAuthSignin := "http://foo.com/global-error-page"

			By("Adding a global-auth-signin to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthSigninSetting, globalExternalAuthSignin)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("error_page 401 = "))
				})
		})

		It(`should add auth headers when global-auth-response-headers is configured`, func() {

			globalExternalAuthResponseHeadersSetting := "global-auth-response-headers"
			globalExternalAuthResponseHeaders := "Foo, Bar"

			By("Adding a global-auth-response-headers to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthResponseHeadersSetting, globalExternalAuthResponseHeaders)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("auth_request_set $authHeader0 $upstream_http_foo;")) &&
						Expect(server).Should(ContainSubstring("auth_request_set $authHeader1 $upstream_http_bar;"))
				})
		})

		It(`should set request-redirect when global-auth-request-redirect is configured`, func() {

			globalExternalAuthRequestRedirectSetting := "global-auth-request-redirect"
			globalExternalAuthRequestRedirect := "Foo-Redirect"

			By("Adding a global-auth-request-redirect to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthRequestRedirectSetting, globalExternalAuthRequestRedirect)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(globalExternalAuthRequestRedirect))
				})
		})

		It(`should set snippet when global external auth is configured`, func() {

			globalExternalAuthSnippetSetting := "global-auth-snippet"
			globalExternalAuthSnippet := "proxy_set_header My-Custom-Header 42;"

			By("Adding a global-auth-snippet to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthSnippetSetting, globalExternalAuthSnippet)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(globalExternalAuthSnippet))
				})
		})

	})

})
