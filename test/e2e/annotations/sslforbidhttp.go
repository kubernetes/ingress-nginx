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
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("ssl-forbid-http", func() {
	f := framework.NewDefaultFramework("sslforbidhttp")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should send forbidden errors for http when tls is present", func() {
		host := "sslforbid.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/ssl-forbid-http": "true",
			"nginx.ingress.kubernetes.io/ssl-redirect":    "true",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should pass through for http when tls is absent", func() {
		host := "sslforbidnotls.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/ssl-forbid-http": "true",
			"nginx.ingress.kubernetes.io/ssl-redirect":    "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:80", host)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)
	})
})
