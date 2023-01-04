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
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("plugins", func() {
	f := framework.NewDefaultFramework("plugins")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should exist a x-hello-world header", func() {
		f.UpdateNginxConfigMapData("plugins", "hello_world, invalid")

		host := "example.com"
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxConfiguration(
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host)) &&
					strings.Contains(server, `plugins.init({ "hello_world","invalid" })`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("User-Agent", "hello").
			Expect().
			Status(http.StatusOK).
			Body().Contains("x-hello-world=1")
	})
})
