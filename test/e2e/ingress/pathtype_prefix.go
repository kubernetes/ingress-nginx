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

package ingress

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Ingress] [PathType] prefix checks", func() {
	f := framework.NewDefaultFramework("prefix")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should return 404 when prefix /aaa does not match request /aaaccc", func() {
		host := "prefix.path"

		ing := framework.NewSingleIngress("exact", "/aaa", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location /aaa")
			})

		f.HTTPTestClient().
			GET("/aaa").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/aaacccc").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/aaa/cccc").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/aaa/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})
