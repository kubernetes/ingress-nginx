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
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - from-to-www-redirect", func() {
	f := framework.NewDefaultFramework("fromtowwwredirect")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	AfterEach(func() {
	})

	It("should redirect from www HTTP to HTTP", func() {
		By("setting up server for redirect from www")
		host := "fromtowwwredirect.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/from-to-www-redirect": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return Expect(cfg).Should(ContainSubstring(`server_name www.fromtowwwredirect.bar.com;`)) &&
					Expect(cfg).Should(ContainSubstring(`return 308 $scheme://fromtowwwredirect.bar.com$request_uri;`))
			})

		By("sending request to www.fromtowwwredirect.bar.com")

		resp, _, errs := gorequest.New().
			Get(fmt.Sprintf("%s/%s", f.IngressController.HTTPURL, "foo")).
			Retry(10, 1*time.Second, http.StatusNotFound).
			RedirectPolicy(noRedirectPolicyFunc).
			Set("Host", fmt.Sprintf("%s.%s", "www", host)).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusPermanentRedirect))
		Expect(resp.Header.Get("Location")).Should(Equal("http://fromtowwwredirect.bar.com/foo"))
	})

	It("should redirect from www HTTPS to HTTPS", func() {
		By("setting up server for redirect from www")
		host := "fromtowwwredirect.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/from-to-www-redirect": "true",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host, fmt.Sprintf("www.%v", host)}, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).ToNot(HaveOccurred())

		f.WaitForNginxServer(fmt.Sprintf("www.%v", host),
			func(server string) bool {
				return Expect(server).Should(ContainSubstring(`server_name www.fromtowwwredirect.bar.com;`)) &&
					Expect(server).Should(ContainSubstring(fmt.Sprintf("/etc/ingress-controller/ssl/%v-fromtowwwredirect.bar.com.pem", f.IngressController.Namespace))) &&
					Expect(server).Should(ContainSubstring(`return 308 $scheme://fromtowwwredirect.bar.com$request_uri;`))
			})

		By("sending request to www.fromtowwwredirect.bar.com")

		h := fmt.Sprintf("%s.%s", "www", host)

		resp, _, errs := gorequest.New().
			TLSClientConfig(&tls.Config{
				InsecureSkipVerify: true,
				ServerName:         h,
			}).
			Get(f.IngressController.HTTPSURL).
			Retry(10, 1*time.Second, http.StatusNotFound).
			RedirectPolicy(noRedirectPolicyFunc).
			Set("host", h).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusPermanentRedirect))
		Expect(resp.Header.Get("Location")).Should(Equal("https://fromtowwwredirect.bar.com/"))
	})
})
