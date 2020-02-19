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

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Geoip2", func() {
	f := framework.NewDefaultFramework("geoip2")

	host := "geoip2"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should only allow requests from specific countries", func() {
		ginkgo.Skip("GeoIP test are temporarily disabled")

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

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "if ($blocked_country)")
			})

		// Should be blocked
		usIP := "8.8.8.8"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", usIP).
			Expect().
			Status(http.StatusForbidden)

		// Shouldn't be blocked
		australianIP := "1.1.1.1"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", australianIP).
			Expect().
			Status(http.StatusOK)
	})
})
