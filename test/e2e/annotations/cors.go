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
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	originHost = "http://origin.com:8080"
	corsHost   = "cors.foo.com"
)

var _ = framework.DescribeAnnotation("cors-*", func() {
	f := framework.NewDefaultFramework("cors")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(2))
	})

	ginkgo.It("should enable cors", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, "set $cors_allow_methods 'GET, PUT, POST, DELETE, PATCH, OPTIONS';") &&
				strings.Contains(server, "set $cors_allow_headers 'DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization';") &&
				strings.Contains(server, "set $cors_max_age '1728000';") &&
				strings.Contains(server, "set $cors_allow_credentials true;")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should set cors methods to only allow POST, GET", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":        "true",
			"nginx.ingress.kubernetes.io/cors-allow-methods": "POST, GET",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, "set $cors_allow_methods 'POST, GET';")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"POST, GET"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should set cors max-age", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":  "true",
			"nginx.ingress.kubernetes.io/cors-max-age": "200",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, "set $cors_max_age '200';")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"200"})
	})

	ginkgo.It("should disable cors allow credentials", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":            "true",
			"nginx.ingress.kubernetes.io/cors-allow-credentials": "false",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return !strings.Contains(server, "set $cors_allow_credentials true;")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			NotContainsKey("Access-Control-Allow-Credentials").
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should allow origin for cors", func() {
		host := corsHost
		origin := "https://origin.cors.com:8080"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin.cors.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			NotContainsKey("Access-Control-Allow-Origin").
			NotContainsKey("Access-Control-Allow-Credentials").
			NotContainsKey("Access-Control-Allow-Methods").
			NotContainsKey("Access-Control-Allow-Headers").
			NotContainsKey("Access-Control-Max-Age")
	})

	ginkgo.It("should allow headers for cors", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":        "true",
			"nginx.ingress.kubernetes.io/cors-allow-headers": "DNT, User-Agent",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, "set $cors_allow_headers 'DNT, User-Agent';")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT, User-Agent"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should expose headers for cors", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":         "true",
			"nginx.ingress.kubernetes.io/cors-expose-headers": "X-CustomResponseHeader, X-CustomSecondHeader",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, "set $cors_expose_headers 'X-CustomResponseHeader, X-CustomSecondHeader';")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Expose-Headers", []string{"X-CustomResponseHeader, X-CustomSecondHeader"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should allow - single origin for multiple cors values", func() {
		host := corsHost
		origin := "https://origin.cors.com:8080"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin.cors.com:8080, https://origin2.cors.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should not allow - single origin for multiple cors values", func() {
		host := corsHost
		origin := "http://no.origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin2.cors.com, https://origin.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().
			NotContainsKey("Access-Control-Allow-Origin").
			NotContainsKey("Access-Control-Allow-Credentials").
			NotContainsKey("Access-Control-Allow-Methods").
			NotContainsKey("Access-Control-Allow-Headers").
			NotContainsKey("Access-Control-Max-Age")
	})

	ginkgo.It("should allow correct origins - single origin for multiple cors values", func() {
		host := corsHost
		badOrigin := "origin.cors.com:8080"
		origin1 := "https://origin2.cors.com"
		origin2 := "https://origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "origin.cors.com:8080, https://origin2.cors.com, https://origin.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", badOrigin).
			Expect().
			Headers().
			NotContainsKey("Access-Control-Allow-Origin").
			NotContainsKey("Access-Control-Allow-Credentials").
			NotContainsKey("Access-Control-Allow-Methods").
			NotContainsKey("Access-Control-Allow-Headers").
			NotContainsKey("Access-Control-Max-Age")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin1).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin1}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin2}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should allow wildcard origin", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "*",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should not break functionality with wildcard and extra domain", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "*, foo.bar.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should not allow - single origin with port and origin without port", func() {
		host := corsHost
		origin := originHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin2.cors.com, http://origin.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().
			NotContainsKey("Access-Control-Allow-Origin").
			NotContainsKey("Access-Control-Allow-Credentials").
			NotContainsKey("Access-Control-Allow-Methods").
			NotContainsKey("Access-Control-Allow-Headers").
			NotContainsKey("Access-Control-Max-Age")
	})

	ginkgo.It("should not allow - single origin without port and origin with required port", func() {
		host := corsHost
		origin := "http://origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin.cors.com:8080, http://origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().
			NotContainsKey("Access-Control-Allow-Origin").
			NotContainsKey("Access-Control-Allow-Credentials").
			NotContainsKey("Access-Control-Allow-Methods").
			NotContainsKey("Access-Control-Allow-Headers").
			NotContainsKey("Access-Control-Max-Age")
	})

	ginkgo.It("should allow - matching origin with wildcard origin (2 subdomains)", func() {
		host := corsHost
		origin := "http://foo.origin.cors.com"
		origin2 := "http://bar-foo.origin.cors.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://*.origin.cors.com, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin2}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})

	ginkgo.It("should not allow - unmatching origin with wildcard origin (2 subdomains)", func() {
		host := corsHost
		origin := "http://bar.foo.origin.cors.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://*.origin.cors.com, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().
			NotContainsKey("Access-Control-Allow-Origin").
			NotContainsKey("Access-Control-Allow-Credentials").
			NotContainsKey("Access-Control-Allow-Methods").
			NotContainsKey("Access-Control-Allow-Headers").
			NotContainsKey("Access-Control-Max-Age")
	})

	ginkgo.It("should allow - origins with non-http[s] protocols", func() {
		host := corsHost
		origin := "test://localhost"
		origin2 := "tauri://localhost:3000"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "test://localhost, tauri://localhost:3000",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin2}).
			ValueEqual("Access-Control-Allow-Credentials", []string{"true"}).
			ValueEqual("Access-Control-Allow-Methods",
				[]string{"GET, PUT, POST, DELETE, PATCH, OPTIONS"}).
			ValueEqual("Access-Control-Allow-Headers",
				[]string{"DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}).
			ValueEqual("Access-Control-Max-Age", []string{"1728000"})
	})
})
