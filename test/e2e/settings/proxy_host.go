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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Proxy host variable", func() {
	test := "proxy-host"
	f := framework.NewDefaultFramework(test)

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	It("should exist a proxy_host", func() {
		upstreamName := fmt.Sprintf("%v-http-svc-80", f.IngressController.Namespace)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_headers "Custom-Header: $proxy_host"`,
		}
		f.EnsureIngress(framework.NewSingleIngress(test, "/", test, f.IngressController.Namespace, "http-svc", 80, &annotations))

		f.WaitForNginxConfiguration(
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", test)) &&
					strings.Contains(server, "set $proxy_host $proxy_upstream_name")
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", test).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Custom-Header")).Should(Equal(upstreamName))
	})

	It("should exist a proxy_host using the upstream-vhost annotation value", func() {
		upstreamName := fmt.Sprintf("%v-http-svc-80", f.IngressController.Namespace)
		upstreamVHost := "different.host"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost":        upstreamVHost,
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_headers "Custom-Header: $proxy_host"`,
		}
		f.EnsureIngress(framework.NewSingleIngress(test, "/", test, f.IngressController.Namespace, "http-svc", 80, &annotations))

		f.WaitForNginxConfiguration(
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", test)) &&
					strings.Contains(server, fmt.Sprintf("set $proxy_host $proxy_upstream_name"))
			})

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", test).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Custom-Header")).Should(Equal(upstreamName))
	})
})
