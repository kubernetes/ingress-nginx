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

var _ = framework.DescribeSetting("reuse-port", func() {
	f := framework.NewDefaultFramework("reuse-port")

	host := "reuse-port"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)
	})

	ginkgo.It("reuse port should be enabled by default", func() {
		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, "reuseport")
		})
	})

	ginkgo.It("reuse port should be disabled", func() {
		f.UpdateNginxConfigMapData("reuse-port", "false")

		f.WaitForNginxConfiguration(func(server string) bool {
			return !strings.Contains(server, "reuseport")
		})
	})

	ginkgo.It("reuse port should be enabled", func() {
		f.UpdateNginxConfigMapData("reuse-port", "true")

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, "reuseport")
		})
	})
})
