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
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - X-Forwarded-Prefix", func() {
	f := framework.NewDefaultFramework("xforwardedprefix")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should set the X-Forwarded-Prefix to the annotation value", func() {
		host := "xfp.baz.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/x-forwarded-prefix": "/test/value",
			"nginx.ingress.kubernetes.io/rewrite-target":     "/foo",
		}

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, &annotations))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_set_header X-Forwarded-Prefix \"/test/value\";"))
			})

		uri := "/"
		resp, body, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)+uri).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring("x-forwarded-prefix=/test/value"))
	})

	It("should not add X-Forwarded-Prefix if the annotation value is empty", func() {
		host := "noxfp.baz.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/x-forwarded-prefix": "",
			"nginx.ingress.kubernetes.io/rewrite-target":     "/foo",
		}

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, &annotations))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(And(ContainSubstring(host), Not(ContainSubstring("proxy_set_header X-Forwarded-Prefix"))))
			})

		uri := "/"
		resp, body, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)+uri).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).To(Not(ContainSubstring("x-forwarded-prefix")))
	})
})
