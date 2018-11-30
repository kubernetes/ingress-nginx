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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Alias", func() {
	f := framework.NewDefaultFramework("alias")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should return status code 200 for host 'foo' and 404 for 'bar'", func() {
		host := "foo"
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name foo"))
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", host)))

		resp, body, errs = gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", "bar").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
		Expect(body).Should(ContainSubstring("404 Not Found"))
	})

	It("should return status code 200 for host 'foo' and 'bar'", func() {
		host := "foo"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-alias": "bar",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name foo"))
			})

		hosts := []string{"foo", "bar"}
		for _, host := range hosts {
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", host)))
		}
	})
})
