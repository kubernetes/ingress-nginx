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

package settings

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("[Security] block-*", func() {
	f := framework.NewDefaultFramework("global-access-block")

	host := "global-access-block"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))
	})

	ginkgo.It("should block CIDRs defined in the ConfigMap", func() {
		f.UpdateNginxConfigMapData("block-cidrs", "172.16.0.0/12,192.168.0.0/16,10.0.0.0/8")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "deny 172.16.0.0/12;") &&
					strings.Contains(cfg, "deny 192.168.0.0/16;") &&
					strings.Contains(cfg, "deny 10.0.0.0/8;")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should block User-Agents defined in the ConfigMap", func() {
		f.UpdateNginxConfigMapData("block-user-agents", "~*chrome\\/68\\.0\\.3440\\.106\\ safari\\/537\\.36,AlphaBot")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "~*chrome\\/68\\.0\\.3440\\.106\\ safari\\/537\\.36 1;") &&
					strings.Contains(cfg, "AlphaBot 1;")
			})

		// Should be blocked
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36").
			Expect().
			Status(http.StatusForbidden)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "AlphaBot").
			Expect().
			Status(http.StatusForbidden)

		// Shouldn't be blocked
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 11_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.0 Mobile/15E148 Safari/604.1").
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should block Referers defined in the ConfigMap", func() {
		f.UpdateNginxConfigMapData("block-referers", "~*example\\.com,qwerty")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "~*example\\.com 1;") &&
					strings.Contains(cfg, "qwerty 1;")
			})

		// Should be blocked
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Referer", "example.com").
			Expect().
			Status(http.StatusForbidden)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Referer", "qwerty").
			Expect().
			Status(http.StatusForbidden)

		// Shouldn't be blocked
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Referer", "qwerty123").
			Expect().
			Status(http.StatusOK)
	})
})
