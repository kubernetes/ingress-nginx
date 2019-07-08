/*
Copyright 2019 The Kubernetes Authors.

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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - SATISFY", func() {
	f := framework.NewDefaultFramework("satisfy")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should configure satisfy directive correctly", func() {
		host := "satisfy"
		annotationKey := "nginx.ingress.kubernetes.io/satisfy"

		annotations := map[string]string{
			"any": "any",
			"all": "all",
		}

		results := map[string]string{
			"any": "satisfy any",
			"all": "satisfy all",
		}

		initAnnotations := map[string]string{
			annotationKey: "all",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, &initAnnotations)
		f.EnsureIngress(ing)

		for key, result := range results {
			err := framework.UpdateIngress(f.KubeClientSet, f.Namespace, host, func(ingress *extensions.Ingress) error {
				ingress.ObjectMeta.Annotations[annotationKey] = annotations[key]
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(result))
				})

			resp, body, errs := gorequest.New().
				Get(f.GetURL(framework.HTTP)).
				Retry(10, 1*time.Second, http.StatusNotFound).
				Set("Host", host).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", host)))
		}
	})

	It("should allow multiple auth with satisfy any", func() {
		host := "auth"

		// setup external auth
		f.NewHttpbinDeployment()

		err := framework.WaitForEndpoints(f.KubeClientSet, framework.DefaultTimeout, "httpbin", f.Namespace, 1)
		Expect(err).NotTo(HaveOccurred())

		e, err := f.KubeClientSet.CoreV1().Endpoints(f.Namespace).Get("httpbin", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		httpbinIP := e.Subsets[0].Addresses[0].IP

		// create basic auth secret at ingress
		s := f.EnsureSecret(buildSecret("uname", "pwd", "basic-secret", f.Namespace))

		annotations := map[string]string{
			// annotations for basic auth at ingress
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": s.Name,
			"nginx.ingress.kubernetes.io/auth-realm":  "test basic auth",

			// annotations for external auth
			"nginx.ingress.kubernetes.io/auth-url":    fmt.Sprintf("http://%s/basic-auth/user/password", httpbinIP),
			"nginx.ingress.kubernetes.io/auth-signin": "http://$host/auth/start",

			// set satisfy any
			"nginx.ingress.kubernetes.io/satisfy": "any",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return Expect(server).Should(ContainSubstring("server_name auth"))
		})

		// with basic auth cred
		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			SetBasicAuth("uname", "pwd").End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		// reroute to signin if without basic cred
		resp, _, errs = gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
				return http.ErrUseLastResponse
			}).Param("a", "b").Param("c", "d").
			End()
		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusFound))
		Expect(resp.Header.Get("Location")).Should(Equal(fmt.Sprintf("http://%s/auth/start?rd=http://%s%s", host, host, url.QueryEscape("/?a=b&c=d"))))
	})
})
