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
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Affinity/Sticky Sessions", func() {
	f := framework.NewDefaultFramework("affinity")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(2)
	})

	AfterEach(func() {
	})

	It("should set sticky cookie SERVERID", func() {
		host := "sticky.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":            "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-name": "SERVERID",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		md5Regex := regexp.MustCompile("SERVERID=[0-9a-f]{32}")
		match := md5Regex.FindStringSubmatch(resp.Header.Get("Set-Cookie"))
		Expect(len(match)).Should(BeNumerically("==", 1))

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring(match[0]))
	})

	It("should set sticky cookie with sha1 hash", func() {
		host := "sticky.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":            "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-hash": "sha1",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		sha1Regex := regexp.MustCompile("INGRESSCOOKIE=[0-9a-f]{40}")
		match := sha1Regex.FindStringSubmatch(resp.Header.Get("Set-Cookie"))
		Expect(len(match)).Should(BeNumerically("==", 1))

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring(match[0]))
	})

	It("should set the path to /something on the generated cookie", func() {
		host := "example.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":            "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-name": "SERVERID",
		}

		ing := framework.NewSingleIngress(host, "/something", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/something").
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("Path=/something"))
	})

	It("does not set the path to / on the generated cookie if there's more than one rule referring to the same backend", func() {
		host := "example.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":            "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-name": "SERVERID",
		}

		f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        host,
				Namespace:   f.IngressController.Namespace,
				Annotations: annotations,
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/something",
										Backend: v1beta1.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
									{
										Path: "/somewhereelese",
										Backend: v1beta1.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/something").
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("Path=/something;"))

		resp, _, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+"/somewhereelese").
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("Path=/somewhereelese;"))
	})

	It("should set cookie with expires", func() {
		host := "cookie.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":               "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-name":    "ExpiresCookie",
			"nginx.ingress.kubernetes.io/session-cookie-expires": "172800",
			"nginx.ingress.kubernetes.io/session-cookie-max-age": "259200",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		local, _ := time.LoadLocation("GMT")
		duration, _ := time.ParseDuration("48h")
		expected := time.Now().In(local).Add(duration).Format("Mon, 02-Jan-06 15:04")
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring(fmt.Sprintf("Expires=%s", expected)))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("Max-Age=259200"))
	})

	It("should work with use-regex annotation and session-cookie-path", func() {
		host := "cookie.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":            "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-name": "SERVERID",
			"nginx.ingress.kubernetes.io/use-regex":           "true",
			"nginx.ingress.kubernetes.io/session-cookie-path": "/foo/bar",
		}

		ing := framework.NewSingleIngress(host, "/foo/.*", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo/bar").
			Set("Host", host).
			End()

		md5Regex := regexp.MustCompile("SERVERID=[0-9a-f]{32}")
		match := md5Regex.FindStringSubmatch(resp.Header.Get("Set-Cookie"))
		Expect(len(match)).Should(BeNumerically("==", 1))

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring(match[0]))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("Path=/foo/bar"))
	})

	It("should warn user when use-regex is true and session-cookie-path is not set", func() {
		host := "cookie.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity":            "cookie",
			"nginx.ingress.kubernetes.io/session-cookie-name": "SERVERID",
			"nginx.ingress.kubernetes.io/use-regex":           "true",
		}

		ing := framework.NewSingleIngress(host, "/foo/.*", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo/bar").
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		logs, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(ContainSubstring(`session-cookie-path should be set when use-regex is true`))
	})

	It("should not set affinity across all server locations when using separate ingresses", func() {
		host := "cookie.foo.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/affinity": "cookie",
		}
		ing1 := framework.NewSingleIngress("ingress1", "/foo/bar", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing1)

		ing2 := framework.NewSingleIngress("ingress2", "/foo", host, f.IngressController.Namespace, "http-svc", 80, &map[string]string{})
		f.EnsureIngress(ing2)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location /foo/bar`) && strings.Contains(server, `location /foo`)
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo").
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(Equal(""))

		resp, _, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo/bar").
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("Path=/foo/bar"))
	})
})
