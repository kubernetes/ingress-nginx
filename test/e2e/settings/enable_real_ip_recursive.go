/*
Copyright 2024 The Kubernetes Authors.

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

const forwardedForHost = "forwarded-for-header"

var _ = framework.DescribeSetting("enable-real-ip-recursive", func() {
	f := framework.NewDefaultFramework("enable-real-ip-recursive")

	setting := "enable-real-ip-recursive"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()

		f.SetNginxConfigMapData(map[string]string{
			"log-format-escape-json": "true",
			"log-format-upstream":    "clientip=\"$remote_addr\"",
			"use-forwarded-headers":  "true",
			setting:                  "false",
		})
	})

	ginkgo.It("should use the first IP in X-Forwarded-For header when setting is true", func() {
		host := forwardedForHost

		f.UpdateNginxConfigMapData(setting, "true")

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(conf string) bool {
			return strings.Contains(conf, "real_ip_recursive on;")
		})

		ginkgo.By("ensuring single values are parsed correctly")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "127.0.0.1, 1.2.3.4").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-for=127.0.0.1")

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, "clientip=\"127.0.0.1\"")
	})

	ginkgo.It("should use the last IP in X-Forwarded-For header when setting is false", func() {
		host := forwardedForHost

		f.UpdateNginxConfigMapData(setting, "false")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxConfiguration(func(conf string) bool {
			return strings.Contains(conf, "real_ip_recursive off;")
		})

		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "127.0.0.1, 1.2.3.4").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-for=1.2.3.4")

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, "clientip=\"1.2.3.4\"")
	})
})
