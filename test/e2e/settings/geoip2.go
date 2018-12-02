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

package settings

import (
	"strings"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Geoip2", func() {
	f := framework.NewDefaultFramework("geoip2")

	host := "geoip2"

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	It("should only allow requests from specific countries", func() {
		f.UpdateNginxConfigMapData("use-geoip2", "true")

		httpSnippetAllowingOnlyAustralia :=
			`map $geoip2_city_country_code $blocked_country {
  default 1;
  AU 0;
}`
		f.UpdateNginxConfigMapData("http-snippet", httpSnippetAllowingOnlyAustralia)
		f.UpdateNginxConfigMapData("use-forwarded-headers", "true")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "map $geoip2_city_country_code $blocked_country")
			})

		configSnippet :=
			`if ($blocked_country) {
  return 403;
}`

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": configSnippet,
		}

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "if ($blocked_country)")
			})

		// Should be blocked
		usIP := "8.8.8.8"
		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			Set("X-Forwarded-For", usIP).
			End()
		Expect(errs).To(BeNil())
		Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))

		// Shouldn't be blocked
		australianIP := "1.1.1.1"
		resp, _, errs = gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			Set("X-Forwarded-For", australianIP).
			End()
		Expect(errs).To(BeNil())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})
})
