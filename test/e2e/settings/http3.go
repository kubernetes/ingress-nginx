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

var _ = framework.DescribeSetting("http3", func() {
	f := framework.NewDefaultFramework("http3")

	ginkgo.It("should disable HTTP/3", func() {
		host := "http3.com"
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.UpdateNginxConfigMapData("use-http3", "false")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return !strings.Contains(cfg, "quic;")
		})
	})

	ginkgo.It("should enable HTTP/3", func() {
		host := "http3.com"
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.UpdateNginxConfigMapData("use-http3", "true")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic;")
		})
	})

	ginkgo.It("should set http3_max_concurrent_streams value", func() {
		f.UpdateNginxConfigMapData("http3-max-concurrent-streams", "256")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_max_concurrent_streams 256;")
		})
	})

	ginkgo.It("should set http3_stream_buffer_size value", func() {
		f.UpdateNginxConfigMapData("http3-stream-buffer-size", "128k")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_stream_buffer_size 128k;")
		})
	})
})
