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

package annotations

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("proxy-intercept-errors", func() {
	f := framework.NewDefaultFramework("proxy-intercept-errors")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("configures Nginx correctly", func() {
		pieState := "true"
		host := "pie.foo.com"

		errorCodes := []string{"404"}

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-http-errors":     strings.Join(errorCodes, ","),
			"nginx.ingress.kubernetes.io/proxy-intercept-errors": pieState,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		var serverConfig string
		f.WaitForNginxServer(host, func(sc string) bool {
			serverConfig = sc
			return strings.Contains(serverConfig, fmt.Sprintf("server_name %s", host))
		})

		ginkgo.By("turning on proxy_intercept_errors directive")
		assert.Contains(ginkgo.GinkgoT(), serverConfig, "proxy_intercept_errors on;")

		// --
		pieState = "false"

		ginkgo.By("updating configuration when only proxy-intercept-errors value changes")
		err := framework.UpdateIngress(f.KubeClientSet, f.Namespace, host, func(ingress *networking.Ingress) error {
			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/proxy-intercept-errors"] = pieState
			return nil
		})
		assert.Nil(ginkgo.GinkgoT(), err)

		f.WaitForNginxServer(host, func(sc string) bool {
			if serverConfig != sc {
				serverConfig = sc
				return true
			}
			return false
		})

		ginkgo.By("turning off proxy_intercept_errors directive")
		assert.NotContains(ginkgo.GinkgoT(), serverConfig, "proxy_intercept_errors on;")

	})
})
