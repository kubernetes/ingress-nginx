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

var _ = framework.IngressNginxDescribe("Annotations - Fromtowwwredirect", func() {
	f := framework.NewDefaultFramework("fromtowwwredirect")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(2)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should redirect from www", func() {
		By("setting up server for redirect from www")
		host := "fromtowwwredirect.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/from-to-www-redirect": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return Expect(cfg).Should(ContainSubstring(`server_name www.fromtowwwredirect.bar.com;`)) &&
					Expect(cfg).Should(ContainSubstring(`return 308 $scheme://fromtowwwredirect.bar.com$request_uri;`))
			})
		Expect(err).NotTo(HaveOccurred())

		By("sending request to www.fromtowwwredirect.bar.com")

		gorequest.New().
			Get(fmt.Sprintf("%s/%s", f.IngressController.HTTPURL, "foo")).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", fmt.Sprintf("%s.%s", "www", host)).
			End()

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		Expect(log).To(ContainSubstring(fmt.Sprintf(` "GET /foo HTTP/1.1" 308 171 "-" "Go-http-client/1.1"`)))
	})
})
