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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Approot", func() {
	f := framework.NewDefaultFramework("approot")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(2)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should redirect to /foo", func() {
		host := "approot.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/app-root": "/foo",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring(`if ($uri = /) {`)) &&
					Expect(server).Should(ContainSubstring(`return 302 /foo;`))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			RedirectPolicy(noRedirectPolicyFunc).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusFound))
		Expect(resp.Header.Get("Location")).Should(Equal("http://approot.bar.com/foo"))
	})
})
