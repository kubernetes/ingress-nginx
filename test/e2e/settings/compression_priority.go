/*
Copyright 2025 The Kubernetes Authors.

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

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("compression-priority", func() {
	f := framework.NewDefaultFramework("compression-priority")

	host := "compression-priority"

	ginkgo.BeforeEach(func() {
		f.NewHttpbunDeployment()
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBunService, 80, nil))
	})

	ginkgo.It("should prefer brotli when compression-priority is set to brotli,gzip", func() {
		f.UpdateNginxConfigMapData("enable-brotli", "true")
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("compression-priority", "brotli,gzip")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf("server_name %v", host)) &&
					strings.Contains(cfg, "brotli on;") &&
					!strings.Contains(cfg, "gzip on;")
			},
		)

		f.HTTPTestClient().
			GET("/html").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "br").
			Expect().
			Status(http.StatusOK).
			ContentEncoding("br")

		// gzip should not be used when brotli is preferred
		f.HTTPTestClient().
			GET("/html").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentEncoding()
	})

	ginkgo.It("should prefer gzip when compression-priority is set to gzip,brotli", func() {
		f.UpdateNginxConfigMapData("enable-brotli", "true")
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("compression-priority", "gzip,brotli")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf("server_name %v", host)) &&
					strings.Contains(cfg, "gzip on;") &&
					!strings.Contains(cfg, "brotli on;")
			},
		)

		f.HTTPTestClient().
			GET("/html").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentEncoding("gzip")

		// brotli should not be used when gzip is preferred
		f.HTTPTestClient().
			GET("/html").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "br").
			Expect().
			Status(http.StatusOK).
			ContentEncoding()
	})
})
