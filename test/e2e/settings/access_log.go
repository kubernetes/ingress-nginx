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

	"github.com/onsi/ginkgo"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("access-log", func() {
	f := framework.NewDefaultFramework("access-log")

	ginkgo.Context("access-log-path", func() {

		ginkgo.It("use the default configuration", func() {
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /var/log/nginx/access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /var/log/nginx/access.log log_stream")
				})
		})

		ginkgo.It("use the specified configuration", func() {
			f.UpdateNginxConfigMapData("access-log-path", "/tmp/access.log")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /tmp/access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /tmp/access.log log_stream")
				})
		})
	})

	ginkgo.Context("http-access-log-path", func() {

		ginkgo.It("use the specified configuration", func() {
			f.UpdateNginxConfigMapData("http-access-log-path", "/tmp/http-access.log")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /tmp/http-access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /var/log/nginx/access.log log_stream")
				})
		})
	})

	ginkgo.Context("stream-access-log-path", func() {

		ginkgo.It("use the specified configuration", func() {
			f.UpdateNginxConfigMapData("stream-access-log-path", "/tmp/stream-access.log")
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /tmp/stream-access.log log_stream") &&
						strings.Contains(cfg, "access_log /var/log/nginx/access.log upstreaminfo")
				})
		})
	})

	ginkgo.Context("http-access-log-path & stream-access-log-path", func() {

		ginkgo.It("use the specified configuration", func() {
			f.SetNginxConfigMapData(map[string]string{
				"http-access-log-path":   "/tmp/http-access.log",
				"stream-access-log-path": "/tmp/stream-access.log",
			})
			f.WaitForNginxConfiguration(
				func(cfg string) bool {
					return strings.Contains(cfg, "access_log /tmp/http-access.log upstreaminfo") &&
						strings.Contains(cfg, "access_log /tmp/stream-access.log log_stream")
				})
		})
	})
})
