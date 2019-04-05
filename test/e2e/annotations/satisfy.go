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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	extensions "k8s.io/api/extensions/v1beta1"
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
})
