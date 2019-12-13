/*
Copyright 2019 The Kubernetes Authors.

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
	"strings"

	. "github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Mirror", func() {
	f := framework.NewDefaultFramework("mirror")
	host := "mirror.foo.com"

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should set mirror-uri to /mirror", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/mirror-uri": "/mirror",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "mirror /mirror;") && strings.Contains(server, "mirror_request_body on;")
			})
	})

	It("should disable mirror-request-body", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/mirror-uri":          "/mirror",
			"nginx.ingress.kubernetes.io/mirror-request-body": "off",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "mirror /mirror;") && strings.Contains(server, "mirror_request_body off;")
			})
	})
})
