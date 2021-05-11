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
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("use-forwarded-headers", func() {
	f := framework.NewDefaultFramework("forwarded-headers")

	setting := "use-forwarded-headers"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.UpdateNginxConfigMapData(setting, "false")
	})

	ginkgo.It("should trust X-Forwarded headers when setting is true", func() {
		host := "forwarded-headers"

		f.UpdateNginxConfigMapData(setting, "true")

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-headers") &&
					strings.Contains(server, "proxy_set_header X-Forwarded-Proto $pass_access_scheme;")
			})

		ginkgo.By("ensuring single values are parsed correctly")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-Port", "1234").
			WithHeader("X-Forwarded-Proto", "myproto").
			WithHeader("X-Forwarded-Scheme", "myproto").
			WithHeader("X-Forwarded-For", "1.2.3.4").
			WithHeader("X-Forwarded-Host", "myhost").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=myhost"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-host=myhost"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-proto=myproto"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-scheme=myproto"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-port=1234"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-for=1.2.3.4"))

		ginkgo.By("ensuring that first entry in X-Forwarded-Host is used as the best host")
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-Port", "1234").
			WithHeader("X-Forwarded-Proto", "myproto").
			WithHeader("X-Forwarded-For", "1.2.3.4").
			WithHeader("X-Forwarded-Host", "myhost.com, another.host,example.net").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=myhost.com"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-host=myhost.com"))
	})

	ginkgo.It("should not trust X-Forwarded headers when setting is false", func() {
		host := "forwarded-headers"

		f.UpdateNginxConfigMapData(setting, "false")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-headers") &&
					strings.Contains(server, "proxy_set_header X-Forwarded-Proto $pass_access_scheme;")
			})

		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-Port", "1234").
			WithHeader("X-Forwarded-Proto", "myproto").
			WithHeader("X-Forwarded-Scheme", "myproto").
			WithHeader("X-Forwarded-For", "1.2.3.4").
			WithHeader("X-Forwarded-Host", "myhost").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=forwarded-headers"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-port=80"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-proto=http"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-scheme=http"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-original-forwarded-for=1.2.3.4"))
		assert.NotContains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=myhost"))
		assert.NotContains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-host=myhost"))
		assert.NotContains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-proto=myproto"))
		assert.NotContains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-scheme=myproto"))
		assert.NotContains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-port=1234"))
		assert.NotContains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-for=1.2.3.4"))
	})
})
