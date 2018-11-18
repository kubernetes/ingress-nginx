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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - CORS", func() {
	f := framework.NewDefaultFramework("cors")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(2)
	})

	AfterEach(func() {
	})

	It("should enable cors", func() {
		host := "cors.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS';"))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Origin: *';"))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Headers: DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization';"))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Max-Age: 1728000';"))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Credentials: true';"))
			})

		uri := "/"
		resp, _, errs := gorequest.New().
			Options(f.IngressController.HTTPURL+uri).
			Set("Host", host).
			End()
		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusNoContent))
	})

	It("should set cors methods to only allow POST, GET", func() {
		host := "cors.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":        "true",
			"nginx.ingress.kubernetes.io/cors-allow-methods": "POST, GET",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Methods: POST, GET';"))
			})
	})

	It("should set cors max-age", func() {
		host := "cors.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":  "true",
			"nginx.ingress.kubernetes.io/cors-max-age": "200",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Max-Age: 200';"))
			})
	})

	It("should disable cors allow credentials", func() {
		host := "cors.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":            "true",
			"nginx.ingress.kubernetes.io/cors-allow-credentials": "false",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).ShouldNot(ContainSubstring("more_set_headers 'Access-Control-Allow-Credentials: true';"))
			})
	})

	It("should allow origin for cors", func() {
		host := "cors.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin.cors.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Origin: https://origin.cors.com:8080';"))
			})
	})

	It("should allow headers for cors", func() {
		host := "cors.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":        "true",
			"nginx.ingress.kubernetes.io/cors-allow-headers": "DNT, User-Agent",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("more_set_headers 'Access-Control-Allow-Headers: DNT, User-Agent';"))
			})
	})
})
