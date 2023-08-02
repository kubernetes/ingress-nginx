/*
Copyright 2022 The Kubernetes Authors.

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

package endpointslices

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Endpointslices] long service name", func() {
	f := framework.NewDefaultFramework("endpointslices")
	host := "longsvcname.foo.com"
	name := "long-name-foobar-foobar-foobar-foobar-foobar-bar-foo-bar-foobuz"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithName(name))
	})

	ginkgo.It("should return 200 when service name has max allowed number of characters 63", func() {

		annotations := make(map[string]string)
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, name, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %s", host))
		})

		ginkgo.By("checking if the service is reached")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})
