/*
Copyright 2023 The Kubernetes Authors.

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

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("gzip", func() {
	f := framework.NewDefaultFramework("gzip")

	host := "gzip"

	ginkgo.BeforeEach(func() {
		f.NewHttpbunDeployment()
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBunService, 80, nil))
	})

	ginkgo.It("should be disabled by default", func() {
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "gzip on;")
			},
		)

		f.HTTPTestClient().
			GET("/xml").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentEncoding()
	})

	ginkgo.It("should be enabled with default settings", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				defaultCfg := config.NewDefault()
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, fmt.Sprintf("gzip_comp_level %d;", defaultCfg.GzipLevel)) &&
					!strings.Contains(cfg, "gzip_disable") &&
					strings.Contains(cfg, "gzip_http_version 1.1;") &&
					strings.Contains(cfg, fmt.Sprintf("gzip_min_length %d;", defaultCfg.GzipMinLength)) &&
					strings.Contains(cfg, fmt.Sprintf("gzip_types %s;", defaultCfg.GzipTypes)) &&
					strings.Contains(cfg, "gzip_proxied any;") &&
					strings.Contains(cfg, "gzip_vary on;")
			},
		)

		f.HTTPTestClient().
			GET("/xml").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentEncoding("gzip")
	})

	ginkgo.It("should set gzip_comp_level to 4", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-level", "4")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, "gzip_comp_level 4;")
			},
		)

		f.HTTPTestClient().
			GET("/xml").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentEncoding("gzip")
	})

	ginkgo.It("should set gzip_disable to msie6", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-disable", "msie6")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					(strings.Contains(cfg, `gzip_disable "msie6";`) || strings.Contains(cfg, `gzip_disable msie6;`))
			},
		)

		f.HTTPTestClient().
			GET("/xml").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			WithHeader("User-Agent", "Mozilla/4.8 [en] (Windows NT 5.1; U)").
			Expect().
			Status(http.StatusOK).
			ContentEncoding("gzip")

		f.HTTPTestClient().
			GET("/xml").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			WithHeader("User-Agent", "Mozilla/45.0 (compatible; MSIE 6.0; Windows NT 5.1)").
			Expect().
			Status(http.StatusOK).
			ContentEncoding()
	})

	ginkgo.It("should set gzip_min_length to 100", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-min-length", "100")
		f.UpdateNginxConfigMapData("gzip-types", "application/octet-stream")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, "gzip_min_length 100;") &&
					strings.Contains(cfg, "gzip_types application/octet-stream;")
			},
		)

		f.HTTPTestClient().
			GET("/bytes/99").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentType("application/octet-stream").
			ContentEncoding()

		f.HTTPTestClient().
			GET("/bytes/100").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentType("application/octet-stream").
			ContentEncoding("gzip")
	})

	ginkgo.It("should set gzip_types to text/html", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-types", "text/html")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, "gzip_types text/html;")
			},
		)

		f.HTTPTestClient().
			GET("/xml").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentType("application/xml").
			ContentEncoding()

		f.HTTPTestClient().
			GET("/html").
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "gzip").
			Expect().
			Status(http.StatusOK).
			ContentType("text/html").
			ContentEncoding("gzip")
	})
})
