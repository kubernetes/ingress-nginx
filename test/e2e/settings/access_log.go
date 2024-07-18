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
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("access-log", func() {
	f := framework.NewDefaultFramework("access-log")

	ginkgo.Context("access-log-path", func() {
		ginkgo.It("use the default configuration", func() {
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					if framework.IsCrossplane() {
						return strings.Contains(cfg, "access_log /var/log/nginx/access.log upstreaminfo") ||
							strings.Contains(cfg, "access_log syslog:server=127.0.0.1:11514 upstreaminfo")
					}
					return (strings.Contains(cfg, "access_log /var/log/nginx/access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /var/log/nginx/access.log log_stream")) ||
						(strings.Contains(cfg, "access_log syslog:server=127.0.0.1:11514 upstreaminfo") &&
							strings.Contains(cfg, "access_log syslog:server=127.0.0.1:11514 log_stream"))
				})
		})

		ginkgo.It("use the specified configuration", func() {
			f.UpdateNginxConfigMapData("access-log-path", "/tmp/nginx/access.log")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					if framework.IsCrossplane() {
						return strings.Contains(cfg, "access_log /tmp/nginx/access.log upstreaminfo")
					}
					return strings.Contains(cfg, "access_log /tmp/nginx/access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /tmp/nginx/access.log log_stream")
				})
		})
	})

	ginkgo.Context("http-access-log-path", func() {
		ginkgo.It("use the specified configuration", func() {
			f.UpdateNginxConfigMapData("http-access-log-path", "/tmp/nginx/http-access.log")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					if framework.IsCrossplane() {
						return strings.Contains(cfg, "access_log /tmp/nginx/http-access.log upstreaminfo")
					}
					return strings.Contains(cfg, "access_log /tmp/nginx/http-access.log upstreaminfo") &&
						(strings.Contains(cfg, "access_log /var/log/nginx/access.log log_stream") ||
							strings.Contains(cfg, "access_log syslog:server=127.0.0.1:11514 log_stream"))
				})
		})
	})

	ginkgo.Context("stream-access-log-path", func() {
		ginkgo.It("use the specified configuration", func() {
			if framework.IsCrossplane() {
				ginkgo.Skip("Crossplane does not support stream")
			}
			f.UpdateNginxConfigMapData("stream-access-log-path", "/tmp/nginx/stream-access.log")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /tmp/nginx/stream-access.log log_stream") &&
						(strings.Contains(cfg, "access_log /var/log/nginx/access.log upstreaminfo") ||
							strings.Contains(cfg, "access_log syslog:server=127.0.0.1:11514 upstreaminfo"))
				})
		})
	})

	ginkgo.Context("http-access-log-path & stream-access-log-path", func() {
		ginkgo.It("use the specified configuration", func() {
			if framework.IsCrossplane() {
				ginkgo.Skip("Crossplane does not support stream")
			}
			f.SetNginxConfigMapData(map[string]string{
				"http-access-log-path":   "/tmp/nginx/http-access.log",
				"stream-access-log-path": "/tmp/nginx/stream-access.log",
			})
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /tmp/nginx/http-access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /tmp/nginx/stream-access.log log_stream")
				})
		})
	})
})
