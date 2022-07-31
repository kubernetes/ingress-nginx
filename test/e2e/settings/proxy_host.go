/*
Copyright 2018 The Kubernetes Authors.

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

var _ = framework.IngressNginxDescribe("Dynamic $proxy_host", func() {
	test := "proxy-host"
	f := framework.NewDefaultFramework(test)

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should exist a proxy_host", func() {
		upstreamName := fmt.Sprintf("%v-%v-80", f.Namespace, framework.EchoService)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_headers "Custom-Header: $proxy_host"`,
		}
		f.EnsureIngress(framework.NewSingleIngress(test, "/", test, f.Namespace, framework.EchoService, 80, annotations))

		f.WaitForNginxConfiguration(
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", test)) &&
					strings.Contains(server, "set $proxy_host $proxy_upstream_name")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", test).
			Expect().
			Status(http.StatusOK).
			Header("Custom-Header").Equal(upstreamName)
	})

	ginkgo.It("should exist a proxy_host using the upstream-vhost annotation value", func() {
		upstreamName := fmt.Sprintf("%v-%v-80", f.Namespace, framework.EchoService)
		upstreamVHost := "different.host"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost":        upstreamVHost,
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_headers "Custom-Header: $proxy_host"`,
		}
		f.EnsureIngress(framework.NewSingleIngress(test, "/", test, f.Namespace, framework.EchoService, 80, annotations))

		f.WaitForNginxConfiguration(
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", test)) &&
					strings.Contains(server, fmt.Sprintf("set $proxy_host $proxy_upstream_name"))
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", test).
			Expect().
			Status(http.StatusOK).
			Header("Custom-Header").Equal(upstreamName)
	})
})
