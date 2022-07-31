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

package ingress

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Ingress] DeepInspection", func() {
	f := framework.NewDefaultFramework("deep-inspection")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should drop whole ingress if one path matches invalid regex", func() {
		host := "inspection123.com"

		ingInvalid := framework.NewSingleIngress("invalidregex", "/bla{alias /var/run/secrets/;}location ~* ^/abcd", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ingInvalid)
		ingValid := framework.NewSingleIngress("valid", "/xpto", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ingValid)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location /xpto") &&
					!strings.Contains(server, "location /bla")
			})

		f.HTTPTestClient().
			GET("/xpto").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/bla").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/abcd/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})
})
