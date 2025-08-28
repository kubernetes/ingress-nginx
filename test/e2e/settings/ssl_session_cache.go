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
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("ssl-session-cache", func() {
	f := framework.NewDefaultFramework("ssl-session-cache")

	ginkgo.It("should have default ssl_session_cache and ssl_session_timeout values", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "ssl_session_cache shared:SSL:10m;") &&
				strings.Contains(cfg, "ssl_session_timeout 10m;")
		})
	})

	ginkgo.It("should disable ssl_session_cache", func() {
		f.UpdateNginxConfigMapData("ssl-session-cache", "false")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return !strings.Contains(cfg, "ssl_session_cache")
		})
	})

	ginkgo.It("should set ssl_session_cache value", func() {
		f.UpdateNginxConfigMapData("ssl-session-cache-size", "20m")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "ssl_session_cache shared:SSL:20m;")
		})
	})

	ginkgo.It("should set ssl_session_timeout value", func() {
		f.UpdateNginxConfigMapData("ssl-session-timeout", "30m")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "ssl_session_timeout 30m;")
		})
	})
})
