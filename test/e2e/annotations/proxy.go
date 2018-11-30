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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Proxy", func() {
	f := framework.NewDefaultFramework("proxy")
	host := "proxy.foo.com"

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(2)
	})

	AfterEach(func() {
	})

	It("should set proxy_redirect to off", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-redirect-from": "off",
			"nginx.ingress.kubernetes.io/proxy-redirect-to":   "goodbye.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_redirect off;"))
			})
	})

	It("should set proxy_redirect to default", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-redirect-from": "default",
			"nginx.ingress.kubernetes.io/proxy-redirect-to":   "goodbye.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_redirect default;"))
			})
	})

	It("should set proxy_redirect to hello.com goodbye.com", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-redirect-from": "hello.com",
			"nginx.ingress.kubernetes.io/proxy-redirect-to":   "goodbye.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_redirect hello.com goodbye.com;"))
			})
	})

	It("should set proxy client-max-body-size to 8m", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-body-size": "8m",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("client_max_body_size 8m;"))
			})
	})

	It("should not set proxy client-max-body-size to incorrect value", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-body-size": "15r",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).ShouldNot(ContainSubstring("client_max_body_size 15r;"))
			})
	})

	It("should set valid proxy timeouts", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-connect-timeout": "50",
			"nginx.ingress.kubernetes.io/proxy-send-timeout":    "20",
			"nginx.ingress.kubernetes.io/proxy-read-timeout":    "20",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_connect_timeout 50s;") &&
					strings.Contains(server, "proxy_send_timeout 20s;") &&
					strings.Contains(server, "proxy_read_timeout 20s;")
			})
	})

	It("should not set invalid proxy timeouts", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-connect-timeout": "50k",
			"nginx.ingress.kubernetes.io/proxy-send-timeout":    "20k",
			"nginx.ingress.kubernetes.io/proxy-read-timeout":    "20k",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "proxy_connect_timeout 50ks;") &&
					!strings.Contains(server, "proxy_send_timeout 20ks;") &&
					!strings.Contains(server, "proxy_read_timeout 20ks;")
			})
	})

	It("should turn on proxy-buffering", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-buffering":   "on",
			"nginx.ingress.kubernetes.io/proxy-buffer-size": "8k",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_buffering on;") &&
					strings.Contains(server, "proxy_buffer_size 8k;") &&
					strings.Contains(server, "proxy_buffers 4 8k;") &&
					strings.Contains(server, "proxy_request_buffering on;")
			})
	})

	It("should turn off proxy-request-buffering", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-request-buffering": "off",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_request_buffering off;"))
			})
	})

	It("should build proxy next upstream", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-next-upstream":       "error timeout http_502",
			"nginx.ingress.kubernetes.io/proxy-next-upstream-tries": "5",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_next_upstream error timeout http_502;") &&
					strings.Contains(server, "proxy_next_upstream_tries 5;")
			})
	})

	It("should setup proxy cookies", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-cookie-domain": "localhost example.org",
			"nginx.ingress.kubernetes.io/proxy-cookie-path":   "/one/ /",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_cookie_domain localhost example.org;") &&
					strings.Contains(server, "proxy_cookie_path /one/ /;")
			})
	})
})
