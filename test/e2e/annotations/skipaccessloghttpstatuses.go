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

package annotations

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("skip-access-log-http-statuses", func() {
	f := framework.NewDefaultFramework("skipaccessloghttpstatuses")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("skip-access-log-http-statuses 200 literal, 200 OK", func() {
		host := "skipaccessloghttpstatuses.go.foo.com"

		f.UpdateNginxConfigMapData("skip-access-log-http-statuses", "200")
		ing := framework.NewSingleIngress(host, "/prefixOne", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(ngx string) bool {
			return strings.Contains(ngx, `~200 0;`)
		})

		f.HTTPTestClient().
			GET("/prefixOne").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.NotContains(ginkgo.GinkgoT(), logs, `GET /prefixOne HTTP/1.1" 200`)
	})

	ginkgo.It("skip-access-log-http-statuses ^2.. regex, 200 OK", func() {
		host := "skipaccessloghttpstatuses.go.foo.com"

		f.UpdateNginxConfigMapData("skip-access-log-http-statuses", "^2..")
		ing := framework.NewSingleIngress(host, "/prefixOne", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(ngx string) bool {
			return strings.Contains(ngx, `~^2.. 0;`)
		})

		f.HTTPTestClient().
			GET("/prefixOne").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.NotContains(ginkgo.GinkgoT(), logs, `GET /prefixOne HTTP/1.1" 200`)
	})

	ginkgo.It("skip-access-log-http-statuses ^2.. regex, 404 Not Found", func() {
		host := "skipaccessloghttpstatuses.go.foo.com"

		f.UpdateNginxConfigMapData("skip-access-log-http-statuses", "^2..")
		ing := framework.NewSingleIngress(host, "/prefixOne", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(ngx string) bool {
			return strings.Contains(ngx, `~^2.. 0;`)
		})

		f.HTTPTestClient().
			GET("/404").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, `GET /404 HTTP/1.1" 404`)
	})
})
