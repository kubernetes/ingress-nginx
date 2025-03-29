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
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	disable               = "false"
	noAuthLocationSetting = "no-auth-locations"
)

var _ = framework.DescribeSetting("[Security] global-auth-url", func() {
	f := framework.NewDefaultFramework(
		"global-external-auth",
		framework.WithHTTPBunEnabled(),
	)

	host := "global-external-auth"

	echoServiceName := framework.EchoService

	globalExternalAuthURLSetting := "global-auth-url"

	fooPath := "/foo"
	barPath := "/bar"

	noAuthSetting := noAuthLocationSetting
	noAuthLocations := barPath

	enableGlobalExternalAuthAnnotation := "nginx.ingress.kubernetes.io/enable-global-auth"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.Context("when global external authentication is configured", func() {
		ginkgo.BeforeEach(func() {
			globalExternalAuthURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:80/status/401", framework.HTTPBunService, f.Namespace)

			ginkgo.By("Adding an ingress rule for /foo")
			fooIng := framework.NewSingleIngress("foo-ingress", fooPath, host, f.Namespace, echoServiceName, 80, nil)
			f.EnsureIngress(fooIng)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "location /foo")
				})

			ginkgo.By("Adding an ingress rule for /bar")
			barIng := framework.NewSingleIngress("bar-ingress", barPath, host, f.Namespace, echoServiceName, 80, nil)
			f.EnsureIngress(barIng)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "location /bar")
				})

			ginkgo.By("Adding a global-auth-url to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthURLSetting, globalExternalAuthURL)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, globalExternalAuthURL)
				})
		})

		ginkgo.It("should return status code 401 when request any protected service", func() {
			ginkgo.By("Sending a request to protected service /foo")
			f.HTTPTestClient().
				GET(fooPath).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusUnauthorized)

			ginkgo.By("Sending a request to protected service /bar")
			f.HTTPTestClient().
				GET(barPath).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusUnauthorized)
		})

		ginkgo.It("should return status code 200 when request whitelisted (via no-auth-locations) service and 401 when request protected service", func() {
			ginkgo.By("Adding a no-auth-locations for /bar to configMap")
			f.UpdateNginxConfigMapData(noAuthSetting, noAuthLocations)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, noAuthLocations)
				})

			ginkgo.By("Sending a request to protected service /foo")
			f.HTTPTestClient().
				GET(fooPath).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusUnauthorized)

			ginkgo.By("Sending a request to whitelisted service /bar")
			f.HTTPTestClient().
				GET(barPath).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It("should return status code 200 when request whitelisted (via ingress annotation) service and 401 when request protected service", func() {
			ginkgo.By("Adding an ingress rule for /bar with annotation enable-global-auth = false")
			err := framework.UpdateIngress(f.KubeClientSet, f.Namespace, "bar-ingress", func(ingress *networking.Ingress) error {
				ingress.ObjectMeta.Annotations[enableGlobalExternalAuthAnnotation] = disable
				return nil
			})
			assert.Nil(ginkgo.GinkgoT(), err)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "location /bar")
				})

			ginkgo.By("Sending a request to protected service /foo")
			f.HTTPTestClient().
				GET(fooPath).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusUnauthorized)

			ginkgo.By("Sending a request to whitelisted service /bar")
			f.HTTPTestClient().
				GET(barPath).
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It("should still return status code 200 after auth backend is deleted using cache", func() {
			globalExternalAuthCacheKeySetting := "global-auth-cache-key"
			globalExternalAuthCacheKey := fooHost
			globalExternalAuthCacheDurationSetting := "global-auth-cache-duration"
			globalExternalAuthCacheDuration := "200 201 401 30m"
			globalExternalAuthURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:80/status/200", framework.HTTPBunService, f.Namespace)

			ginkgo.By("Adding a global-auth-cache-key to configMap")
			f.SetNginxConfigMapData(map[string]string{
				globalExternalAuthCacheKeySetting:      globalExternalAuthCacheKey,
				globalExternalAuthCacheDurationSetting: globalExternalAuthCacheDuration,
				globalExternalAuthURLSetting:           globalExternalAuthURL,
			})

			cacheRegex := regexp.MustCompile(`\$cache_key.*foo`)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return cacheRegex.MatchString(server) &&
						strings.Contains(server, `proxy_cache_valid 200 201 401 30m;`)
				})

			f.HTTPTestClient().
				GET(barPath).
				WithHeader("Host", host).
				WithBasicAuth("user", "password").
				Expect().
				Status(http.StatusOK)

			err := f.DeleteDeployment(framework.HTTPBunService)
			assert.Nil(ginkgo.GinkgoT(), err)
			framework.Sleep()

			f.HTTPTestClient().
				GET(barPath).
				WithHeader("Host", host).
				WithBasicAuth("user", "password").
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It(`should proxy_method method when global-auth-method is configured`, func() {
			globalExternalAuthMethodSetting := "global-auth-method"
			globalExternalAuthMethod := "GET"

			ginkgo.By("Adding a global-auth-method to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthMethodSetting, globalExternalAuthMethod)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "proxy_method")
				})
		})

		ginkgo.It(`should add custom error page when global-auth-signin url is configured`, func() {
			globalExternalAuthSigninSetting := "global-auth-signin"
			globalExternalAuthSignin := "http://foo.com/global-error-page"

			ginkgo.By("Adding a global-auth-signin to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthSigninSetting, globalExternalAuthSignin)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "error_page 401 = ")
				})
		})

		ginkgo.It(`should add auth headers when global-auth-response-headers is configured`, func() {
			globalExternalAuthResponseHeadersSetting := "global-auth-response-headers"
			globalExternalAuthResponseHeaders := "Foo, Bar"

			ginkgo.By("Adding a global-auth-response-headers to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthResponseHeadersSetting, globalExternalAuthResponseHeaders)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, "auth_request_set $authHeader0 $upstream_http_foo;") &&
						strings.Contains(server, "auth_request_set $authHeader1 $upstream_http_bar;")
				})
		})

		ginkgo.It(`should set request-redirect when global-auth-request-redirect is configured`, func() {
			globalExternalAuthRequestRedirectSetting := "global-auth-request-redirect"
			globalExternalAuthRequestRedirect := "Foo-Redirect"

			ginkgo.By("Adding a global-auth-request-redirect to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthRequestRedirectSetting, globalExternalAuthRequestRedirect)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, globalExternalAuthRequestRedirect)
				})
		})

		ginkgo.It(`should set snippet when global external auth is configured`, func() {
			if framework.IsCrossplane() {
				ginkgo.Skip("crossplane does not support snippets")
			}
			globalExternalAuthSnippetSetting := "global-auth-snippet"
			globalExternalAuthSnippet := "proxy_set_header My-Custom-Header 42;"

			ginkgo.By("Adding a global-auth-snippet to configMap")
			f.UpdateNginxConfigMapData(globalExternalAuthSnippetSetting, globalExternalAuthSnippet)
			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, globalExternalAuthSnippet)
				})
		})
	})

	ginkgo.Context("cookie set by external authentication server", func() {
		host := "global-external-auth-check-cookies"
		var ing1, ing2 *networking.Ingress

		cfg := `#
events {
	worker_connections  1024;
	multi_accept on;
}

http {
	default_type 'text/plain';
	client_max_body_size 0;

	server {
		access_log on;
		access_log /dev/stdout;

		listen 80;

		location ~ ^/cookies/set/(?<key>.*)/(?<value>.*) {
			content_by_lua_block {
				ngx.header['Set-Cookie'] = {ngx.var.key.."="..ngx.var.value}
				ngx.say("OK")
			}
		}

		location / {
			return 200;
		}

		location /error {
			return 503;
		}
	}
}
`
		ginkgo.BeforeEach(func() {
			f.NGINXWithConfigDeployment("http-cookie-with-error", cfg)

			e, err := f.KubeClientSet.CoreV1().Endpoints(f.Namespace).Get(context.TODO(), "http-cookie-with-error", metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			assert.GreaterOrEqual(ginkgo.GinkgoT(), len(e.Subsets), 1, "expected at least one endpoint")
			assert.GreaterOrEqual(ginkgo.GinkgoT(), len(e.Subsets[0].Addresses), 1, "expected at least one address ready in the endpoint")

			nginxIP := e.Subsets[0].Addresses[0].IP

			f.UpdateNginxConfigMapData(globalExternalAuthURLSetting, fmt.Sprintf("http://%s/cookies/set/alma/armud", nginxIP))

			ing1 = framework.NewSingleIngress(host, "/", host, f.Namespace, "http-cookie-with-error", 80, nil)
			f.EnsureIngress(ing1)

			ing2 = framework.NewSingleIngress(host+"-error", "/error", host, f.Namespace, "http-cookie-with-error", 80, nil)
			f.EnsureIngress(ing2)

			f.WaitForNginxServer(host, func(server string) bool {
				return strings.Contains(server, "server_name "+host)
			})
		})

		ginkgo.It("user retains cookie by default", func() {
			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				WithQuery("a", "b").
				WithQuery("c", "d").
				Expect().
				Status(http.StatusOK).
				Header("Set-Cookie").Contains("alma=armud")
		})

		ginkgo.It("user does not retain cookie if upstream returns error status code", func() {
			f.HTTPTestClient().
				GET("/error").
				WithHeader("Host", host).
				WithQuery("a", "b").
				WithQuery("c", "d").
				Expect().
				Status(http.StatusServiceUnavailable).
				Header("Set-Cookie").Contains("")
		})

		ginkgo.It("user with global-auth-always-set-cookie key in configmap retains cookie if upstream returns error status code", func() {
			f.UpdateNginxConfigMapData("global-auth-always-set-cookie", "true")

			f.HTTPTestClient().
				GET("/error").
				WithHeader("Host", host).
				WithQuery("a", "b").
				WithQuery("c", "d").
				Expect().
				Status(http.StatusServiceUnavailable).
				Header("Set-Cookie").Contains("alma=armud")
		})
	})
})
