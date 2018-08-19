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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Rewrite", func() {
	f := framework.NewDefaultFramework("rewrite")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(1)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should rewrite request URL", func() {
		By("setting rewrite-target annotation")

		host := "rewrite.foo.com"
		annotations := map[string]string{"nginx.ingress.kubernetes.io/rewrite-target": "/"}
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:8080/", host)

		ing := framework.NewSingleIngress(host, "/something", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "rewrite (?i)/something/(.*) /$1 break;") &&
					strings.Contains(server, "rewrite (?i)/something$ / break;")
			})
		Expect(err).NotTo(HaveOccurred())

		By("sending request to Ingress rule path (lowercase)")

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/something").
			Set("Host", host).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))

		By("sending request to Ingress rule path (mixed case)")

		resp, body, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+"/SomeThing").
			Set("Host", host).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))
	})

	It("should write rewrite logs", func() {
		By("setting enable-rewrite-log annotation")

		host := "rewrite.bar.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target":     "/",
			"nginx.ingress.kubernetes.io/enable-rewrite-log": "true",
		}

		ing := framework.NewSingleIngress(host, "/something", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "rewrite_log on;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/something").
			Set("Host", host).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		logs, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(ContainSubstring(`"(?i)/something$" matches "/something", client:`))
		Expect(logs).To(ContainSubstring(`rewritten data: "/", args: "",`))
	})
})
