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

package settings

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("X-Forwarded headers", func() {
	f := framework.NewDefaultFramework("forwarded-headers")

	setting := "use-forwarded-headers"

	BeforeEach(func() {
		f.NewEchoDeployment()
		f.UpdateNginxConfigMapData(setting, "false")
	})

	AfterEach(func() {
	})

	It("should trust X-Forwarded headers when setting is true", func() {
		host := "forwarded-headers"

		f.UpdateNginxConfigMapData(setting, "true")

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-headers")
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			Set("X-Forwarded-Port", "1234").
			Set("X-Forwarded-Proto", "myproto").
			Set("X-Forwarded-For", "1.2.3.4").
			Set("X-Forwarded-Host", "myhost").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=myhost")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-host=myhost")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-proto=myproto")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-port=1234")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-for=1.2.3.4")))
	})
	It("should not trust X-Forwarded headers when setting is false", func() {
		host := "forwarded-headers"

		f.UpdateNginxConfigMapData(setting, "false")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-headers")
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			Set("X-Forwarded-Port", "1234").
			Set("X-Forwarded-Proto", "myproto").
			Set("X-Forwarded-For", "1.2.3.4").
			Set("X-Forwarded-Host", "myhost").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=forwarded-headers")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-port=80")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-proto=http")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-original-forwarded-for=1.2.3.4")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("host=myhost")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-host=myhost")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-proto=myproto")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-port=1234")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-for=1.2.3.4")))
	})
})
