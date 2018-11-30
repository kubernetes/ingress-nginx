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
	"net/url"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Auth", func() {
	f := framework.NewDefaultFramework("auth")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should return status code 200 when no authentication is configured", func() {
		host := "auth"

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", host)))
	})

	It("should return status code 503 when authentication is configured with an invalid secret", func() {
		host := "auth"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": "something",
			"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusServiceUnavailable))
		Expect(body).Should(ContainSubstring("503 Service Temporarily Unavailable"))
	})

	It("should return status code 401 when authentication is configured but Authorization header is not configured", func() {
		host := "auth"

		s := f.EnsureSecret(buildSecret("foo", "bar", "test", f.IngressController.Namespace))

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": s.Name,
			"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusUnauthorized))
		Expect(body).Should(ContainSubstring("401 Authorization Required"))
	})

	It("should return status code 401 when authentication is configured and Authorization header is sent with invalid credentials", func() {
		host := "auth"

		s := f.EnsureSecret(buildSecret("foo", "bar", "test", f.IngressController.Namespace))

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": s.Name,
			"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			SetBasicAuth("user", "pass").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusUnauthorized))
		Expect(body).Should(ContainSubstring("401 Authorization Required"))
	})

	It("should return status code 200 when authentication is configured and Authorization header is sent", func() {
		host := "auth"

		s := f.EnsureSecret(buildSecret("foo", "bar", "test", f.IngressController.Namespace))

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": s.Name,
			"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			SetBasicAuth("foo", "bar").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})

	It("should return status code 500 when authentication is configured with invalid content and Authorization header is sent", func() {
		host := "auth"

		s := f.EnsureSecret(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: f.IngressController.Namespace,
				},
				Data: map[string][]byte{
					// invalid content
					"auth": []byte("foo:"),
				},
				Type: corev1.SecretTypeOpaque,
			},
		)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": s.Name,
			"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			SetBasicAuth("foo", "bar").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusInternalServerError))
	})

	It(`should set snippet "proxy_set_header My-Custom-Header 42;" when external auth is configured`, func() {
		host := "auth"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-url": "http://foo.bar/basic-auth/user/password",
			"nginx.ingress.kubernetes.io/auth-snippet": `
				proxy_set_header My-Custom-Header 42;`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring(`proxy_set_header My-Custom-Header 42;`))
			})
	})

	It(`should not set snippet "proxy_set_header My-Custom-Header 42;" when external auth is not configured`, func() {
		host := "auth"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-snippet": `
				proxy_set_header My-Custom-Header 42;`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).ShouldNot(ContainSubstring(`proxy_set_header My-Custom-Header 42;`))
			})
	})

	Context("when external authentication is configured", func() {
		host := "auth"

		BeforeEach(func() {
			f.NewHttpbinDeployment()

			var httpbinIP string

			err := framework.WaitForEndpoints(f.KubeClientSet, framework.DefaultTimeout, "httpbin", f.IngressController.Namespace, 1)
			Expect(err).NotTo(HaveOccurred())

			e, err := f.KubeClientSet.CoreV1().Endpoints(f.IngressController.Namespace).Get("httpbin", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			httpbinIP = e.Subsets[0].Addresses[0].IP

			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/auth-url":    fmt.Sprintf("http://%s/basic-auth/user/password", httpbinIP),
				"nginx.ingress.kubernetes.io/auth-signin": "http://$host/auth/start",
			}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host, func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth"))
			})
		})

		It("should return status code 200 when signed in", func() {
			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Retry(10, 1*time.Second, http.StatusNotFound).
				Set("Host", host).
				SetBasicAuth("user", "password").
				End()

			for _, err := range errs {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})

		It("should redirect to signin url when not signed in", func() {
			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Retry(10, 1*time.Second, http.StatusNotFound).
				Set("Host", host).
				RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
					return http.ErrUseLastResponse
				}).
				Param("a", "b").
				Param("c", "d").
				End()

			for _, err := range errs {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(resp.StatusCode).Should(Equal(http.StatusFound))
			Expect(resp.Header.Get("Location")).Should(Equal(fmt.Sprintf("http://%s/auth/start?rd=http://%s%s", host, host, url.QueryEscape("/?a=b&c=d"))))
		})
	})
})

// TODO: test Digest Auth
//   401
//   Realm name
//   Auth ok
//   Auth error

func buildSecret(username, password, name, namespace string) *corev1.Secret {
	out, err := exec.Command("openssl", "passwd", "-crypt", password).CombinedOutput()
	encpass := fmt.Sprintf("%v:%s\n", username, out)
	Expect(err).NotTo(HaveOccurred())

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       name,
			Namespace:                  namespace,
			DeletionGracePeriodSeconds: framework.NewInt64(1),
		},
		Data: map[string][]byte{
			"auth": []byte(encpass),
		},
		Type: corev1.SecretTypeOpaque,
	}
}
