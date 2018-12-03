/*
Copyright 2017 The Kubernetes Authors.

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - grpc", func() {
	f := framework.NewDefaultFramework("grpc")

	BeforeEach(func() {
		f.NewGRPCFortuneTellerDeployment()
	})

	Context("when grpc is enabled", func() {
		It("should use grpc_pass in the configuration file", func() {
			host := "grpc"

			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
			}

			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "fortune-teller", 50051, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring(fmt.Sprintf("server_name %v", host)))
				})

			f.WaitForNginxServer(host,
				func(server string) bool {
					return Expect(server).Should(ContainSubstring("grpc_pass")) &&
						Expect(server).Should(ContainSubstring("grpc_set_header")) &&
						Expect(server).ShouldNot(ContainSubstring("proxy_pass"))
				})
		})
	})
})
