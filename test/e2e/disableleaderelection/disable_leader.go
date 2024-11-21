/*
Copyright 2024 The Kubernetes Authors.

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

package disableleaderelection

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Disable Leader] Routing works when leader election was disabled", func() {
	f := framework.NewDefaultFramework("disableleaderelection")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should create multiple ingress routings rules when leader election has disabled", func() {
		host1 := "leader.election.disabled.com"
		host2 := "leader.election.disabled2.com"

		ing1 := framework.NewSingleIngress(host1, "/foo", host1, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing1)

		ing2 := framework.NewSingleIngress(host2, "/ping", host2, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing2)

		f.WaitForNginxServer(host1,
			func(server string) bool {
				return strings.Contains(server, host1) &&
					strings.Contains(server, "location /foo")
			})

		f.WaitForNginxServer(host2,
			func(server string) bool {
				return strings.Contains(server, host2) &&
					strings.Contains(server, "location /ping")
			})

		f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host1).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", host1).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host2).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/ping").
			WithHeader("Host", host2).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/pong").
			WithHeader("Host", host2).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/ping").
			WithHeader("Host", host1).
			Expect().
			Status(http.StatusNotFound)
	})
})
