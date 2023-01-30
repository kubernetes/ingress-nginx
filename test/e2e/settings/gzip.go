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
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("gzip", func() {
	f := framework.NewDefaultFramework("gzip")

	ginkgo.It("should be disabled by default", func() {
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "gzip on;")
			})
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
			})
	})

	ginkgo.It("should set gzip_comp_level to 4", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-level", "4")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, "gzip_comp_level 4;")
			})
	})

	ginkgo.It("should set gzip_disable to msie6", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-disable", "msie6")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, `gzip_disable "msie6";`)
			})
	})

	ginkgo.It("should set gzip_min_length to 100", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-min-length", "100")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, "gzip_min_length 100;")
			})
	})

	ginkgo.It("should set gzip_types to application/javascript", func() {
		f.UpdateNginxConfigMapData("use-gzip", "true")
		f.UpdateNginxConfigMapData("gzip-types", "application/javascript")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "gzip on;") &&
					strings.Contains(cfg, "gzip_types application/javascript;")
			})
	})
})
