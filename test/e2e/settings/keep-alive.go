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
	"regexp"
	"strings"

	"github.com/onsi/ginkgo"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("keep-alive keep-alive-requests", func() {
	f := framework.NewDefaultFramework("keep-alive")

	host := "keep-alive"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)
	})

	ginkgo.Context("Check the keep alive", func() {
		ginkgo.It("should set keepalive_timeout", func() {
			f.UpdateNginxConfigMapData("keep-alive", "140")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`keepalive_timeout 140s;`))
			})
		})

		ginkgo.It("should set keepalive_requests", func() {
			f.UpdateNginxConfigMapData("keep-alive-requests", "200")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`keepalive_requests 200;`))
			})

		})
	})

	ginkgo.Context("Check the upstream keep alive", func() {
		ginkgo.It("should set keepalive connection to upstream server", func() {
			f.UpdateNginxConfigMapData("upstream-keepalive-connections", "128")

			f.WaitForNginxConfiguration(func(server string) bool {
				match, _ := regexp.MatchString(`upstream\supstream_balancer\s\{[\s\S]*keepalive 128;`, server)
				return match
			})
		})

		ginkgo.It("should set keep alive connection timeout to upstream server", func() {
			f.UpdateNginxConfigMapData("upstream-keepalive-timeout", "120")

			f.WaitForNginxConfiguration(func(server string) bool {
				match, _ := regexp.MatchString(`upstream\supstream_balancer\s\{[\s\S]*keepalive_timeout\s*120s;`, server)
				return match
			})
		})

		ginkgo.It("should set the request count to upstream server through one keep alive connection", func() {
			f.UpdateNginxConfigMapData("upstream-keepalive-requests", "200")

			f.WaitForNginxConfiguration(func(server string) bool {
				match, _ := regexp.MatchString(`upstream\supstream_balancer\s\{[\s\S]*keepalive_requests\s*200;`, server)
				return match
			})
		})
	})
})
