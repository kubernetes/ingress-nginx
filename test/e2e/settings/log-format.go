/*
Copyright 2020 The Kubernetes Authors.

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

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("log-format-*", func() {
	f := framework.NewDefaultFramework("log-format")

	host := "log-format"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))
	})

	ginkgo.Context("Check log-format-escape-json", func() {

		ginkgo.It("should disable the log-format-escape-json by default", func() {
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return !strings.Contains(cfg, "log_format upstreaminfo escape=json")
				})
		})

		ginkgo.It("should enable the log-format-escape-json", func() {
			f.UpdateNginxConfigMapData("log-format-escape-json", "true")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "log_format upstreaminfo escape=json")
				})
		})

		ginkgo.It("should disable the log-format-escape-json", func() {
			f.UpdateNginxConfigMapData("log-format-escape-json", "false")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return !strings.Contains(cfg, "log_format upstreaminfo escape=json")
				})
		})
	})

	ginkgo.Context("Check log-format-upstream with log-format-escape-json", func() {

		ginkgo.It("log-format-escape-json enabled", func() {
			f.SetNginxConfigMapData(map[string]string{
				"log-format-escape-json": "true",
				"log-format-upstream":    "\"{\"my_header1\":\"$http_header1\", \"my_header2\":\"$http_header2\"}\"",
			})

			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "log_format upstreaminfo escape=json")
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				WithHeader("header1", `Here is "header1" with json escape`).
				Expect().
				Status(http.StatusOK)

			logs, err := f.NginxLogs()
			assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
			assert.Contains(ginkgo.GinkgoT(), logs, `{"my_header1":"Here is \"header1\" with json escape", "my_header2":""}`)
		})

		ginkgo.It("log-format-escape-json disabled", func() {
			f.SetNginxConfigMapData(map[string]string{
				"log-format-escape-json": "false",
				"log-format-upstream":    "\"{\"my_header3\":\"$http_header3\", \"my_header4\":\"$http_header4\"}\"",
			})

			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return !strings.Contains(cfg, "log_format upstreaminfo escape=json")
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				WithHeader("header3", `Here is "header3" with json escape`).
				Expect().
				Status(http.StatusOK)

			logs, err := f.NginxLogs()
			assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
			assert.Contains(ginkgo.GinkgoT(), logs, `{"my_header3":"Here is \x22header3\x22 with json escape", "my_header4":"-"}`)
		})
	})
})
