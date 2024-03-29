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

	ginkgo.It("should test prefix path using simple regex pattern for /id/{int}", func() {
		host := "echo.com.br"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex": `true`,
		}

		ing := framework.NewSingleIngress(host, "/id/[0-9]+", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/id/1").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/id/12").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/id/123").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/id/aaa").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/id/123a").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should test prefix path using regex pattern for /id/{int} ignoring non-digits characters at end of string", func() {
		host := "echo.regex.br"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex": `true`,
		}

		ing := framework.NewSingleIngress(host, "/id/[0-9]+$", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/id/1").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/id/aaa").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/id/123a").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})

	ginkgo.It("should test prefix path using fixed path size regex pattern /id/{int}{3}", func() {
		host := "echo.regex.size.br"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex": `true`,
		}

		ing := framework.NewSingleIngress(host, "/id/[0-9]{3}$", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/id/99").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/id/123").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/id/9999").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/id/123a").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})

	ginkgo.It("should correctly route multi-segment path patterns", func() {
		host := "echo.multi.segment.br"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex": `true`,
		}

		ing := framework.NewSingleIngress(host, "/id/[0-9]+/post/[a-zA-Z]+$", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/id/123/post/abc").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/id/123/post/abc123").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

		f.HTTPTestClient().
			GET("/id/abc/post/abc").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})
})
