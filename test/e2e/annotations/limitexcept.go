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

package annotations

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("limit-except", func() {
	f := framework.NewDefaultFramework("limitexcept")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should return 403 when the request is not allowed", func() {
		host := "foo.com"

        annotations := map[string]string{
            "nginx.ingress.kubernetes.io/server-snippet": `
                location = / {
                    return 403;
                }`,
            "nginx.ingress.kubernetes.io/configuration-snippet": `
                limit_except GET {
                    deny all;
                }`,
        }
    
        ing := framework.NewSingleIngress(host, "/foo", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

        f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location = / { return 403; }`) &&
					strings.Contains(server, `limit_except GET { deny all; }`)
			})

		ginkgo.By("sending request to foo.com")
		f.HTTPTestClient().
		    POST("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)

        f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)

        f.HTTPTestClient().
		    POST("/foo").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)

        f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})
