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

	"github.com/onsi/ginkgo/v2"
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

	ginkgo.Context("Check log-format-escape-json and log-format-escape-none", func() {

		ginkgo.It("should not configure log-format escape by default", func() {
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return !strings.Contains(cfg, "log_format upstreaminfo escape")
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

		ginkgo.It("should enable the log-format-escape-none", func() {
			f.UpdateNginxConfigMapData("log-format-escape-none", "true")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "log_format upstreaminfo escape=none")
				})
		})

		ginkgo.It("should disable the log-format-escape-none", func() {
			f.UpdateNginxConfigMapData("log-format-escape-none", "false")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return !strings.Contains(cfg, "log_format upstreaminfo escape=none")
				})
		})
	})

	ginkgo.Context("Check log-format-upstream with log-format-escape-json and log-format-escape-none", func() {

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

		ginkgo.It("log-format default escape", func() {
			f.SetNginxConfigMapData(map[string]string{
				"log-format-escape-json": "false",
				"log-format-upstream":    "\"{\"my_header3\":\"$http_header3\", \"my_header4\":\"$http_header4\"}\"",
			})

			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return !strings.Contains(cfg, "log_format upstreaminfo escape")
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				WithHeader("header3", `Here is "header3" with default escape`).
				Expect().
				Status(http.StatusOK)

			logs, err := f.NginxLogs()
			assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
			assert.Contains(ginkgo.GinkgoT(), logs, `{"my_header3":"Here is \x22header3\x22 with default escape", "my_header4":"-"}`)
		})

		ginkgo.It("log-format-escape-none enabled", func() {
			f.SetNginxConfigMapData(map[string]string{
				"log-format-escape-none": "true",
				"log-format-upstream":    "\"{\"my_header5\":\"$http_header5\", \"my_header6\":\"$http_header6\"}\"",
			})

			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "log_format upstreaminfo escape=none")
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				WithHeader("header5", `Here is "header5" with none escape`).
				Expect().
				Status(http.StatusOK)

			logs, err := f.NginxLogs()
			assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
			assert.Contains(ginkgo.GinkgoT(), logs, `{"my_header5":"Here is "header5" with none escape", "my_header6":""}`)
		})
	})
})
