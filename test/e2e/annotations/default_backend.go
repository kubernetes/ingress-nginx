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
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - custom default-backend", func() {
	f := framework.NewDefaultFramework("default-backend")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	Context("when default backend annotation is enabled", func() {
		It("should use a custom default backend as upstream", func() {
			host := "default-backend"
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/default-backend": "http-svc",
			}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "invalid", 80, &annotations)
			f.EnsureIngress(ing)

			time.Sleep(5 * time.Second)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(fmt.Sprintf("server_name %v", host)))
				})

			uri := "/alma/armud"
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL+uri).
				Set("Host", host).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			Expect(body).To(ContainSubstring("x-code=503"))
			Expect(body).To(ContainSubstring(fmt.Sprintf("x-ingress-name=%s", host)))
			Expect(body).To(ContainSubstring("x-service-name=invalid"))
			Expect(body).To(ContainSubstring(fmt.Sprintf("x-original-uri=%s", uri)))
		})
	})
})
