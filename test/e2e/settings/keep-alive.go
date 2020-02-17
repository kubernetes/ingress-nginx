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
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Settings - keep alive", func() {
	f := framework.NewDefaultFramework("keep-alive")

	host := "keep-alive"

	BeforeEach(func() {
		f.NewEchoDeployment()
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)
	})

	It("should set keepalive_timeout", func() {
		f.UpdateNginxConfigMapData("keep-alive", "140")

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf(`keepalive_timeout 140s;`))
		})
	})

	It("should set keepalive_requests", func() {
		f.UpdateNginxConfigMapData("keep-alive-requests", "200")

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf(`keepalive_requests 200;`))
		})

	})
})
