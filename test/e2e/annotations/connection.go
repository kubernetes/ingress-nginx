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

var _ = framework.IngressNginxDescribe("Annotations - Connection", func() {
	f := framework.NewDefaultFramework("connection")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(2)
	})

	AfterEach(func() {
	})

	It("set connection header to keep-alive", func() {
		host := "connection.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/connection-proxy-header": "keep-alive",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring(`proxy_set_header Connection keep-alive;`))
			})

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("connection=keep-alive")))
	})
})
