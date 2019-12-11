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

package ssl

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("sslredirect", func() {
	f := framework.NewDefaultFramework("sslredirect")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should redirect from HTTP to HTTPS when secret is missing", func() {
		host := "redirect.com"

		_ = f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name redirect.com") &&
					strings.Contains(server, "listen 443") &&
					strings.Contains(server, "listen 80")
			})

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Set("Host", host).
			RedirectPolicy(func(_ gorequest.Request, _ []gorequest.Request) error {
				return http.ErrUseLastResponse
			}).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusPermanentRedirect))

		location, err := (*http.Response)(resp).Location()
		Expect(err).Should(BeNil())
		Expect(location.String()).Should(Equal("https://redirect.com/"))
	})
})
