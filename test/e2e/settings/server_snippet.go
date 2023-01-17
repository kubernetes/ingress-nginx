/*
Copyright 2021 The Kubernetes Authors.

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
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("configmap server-snippet", func() {
	f := framework.NewDefaultFramework("cm-server-snippet")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should add value of server-snippet setting to all ingress config", func() {
		host := "serverglobalsnippet1.foo.com"
		hostAnnots := "serverannotssnippet1.foo.com"

		f.SetNginxConfigMapData(map[string]string{
			"server-snippet": `
			more_set_headers "Globalfoo: Foooo";`,
		})

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-snippet": `
				more_set_headers "Foo: Bar";
				more_set_headers "Xpto: Lalala";`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		ing1 := framework.NewSingleIngress(hostAnnots, "/", hostAnnots, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing1)

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "Globalfoo: Foooo`) &&
					!strings.Contains(server, `more_set_headers "Foo: Bar";`) &&
					!strings.Contains(server, `more_set_headers "Xpto: Lalala";`)
			})

		f.WaitForNginxServer(hostAnnots,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "Globalfoo: Foooo`) &&
					strings.Contains(server, `more_set_headers "Foo: Bar";`) &&
					strings.Contains(server, `more_set_headers "Xpto: Lalala";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Globalfoo", []string{"Foooo"}).
			NotContainsKey("Foo").
			NotContainsKey("Xpto")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", hostAnnots).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Foo", []string{"Bar"}).
			ValueEqual("Xpto", []string{"Lalala"}).
			ValueEqual("Globalfoo", []string{"Foooo"})
	})

	ginkgo.It("should add global server-snippet and drop annotations per admin config", func() {
		host := "serverglobalsnippet2.foo.com"
		hostAnnots := "serverannotssnippet2.foo.com"

		f.SetNginxConfigMapData(map[string]string{
			"allow-snippet-annotations": "false",
			"server-snippet": `
			more_set_headers "Globalfoo: Foooo";`,
		})

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-snippet": `
				more_set_headers "Foo: Bar";
				more_set_headers "Xpto: Lalala";`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		ing1 := framework.NewSingleIngress(hostAnnots, "/", hostAnnots, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing1)

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "Globalfoo: Foooo`) &&
					!strings.Contains(server, `more_set_headers "Foo: Bar";`) &&
					!strings.Contains(server, `more_set_headers "Xpto: Lalala";`)
			})

		f.WaitForNginxServer(hostAnnots,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "Globalfoo: Foooo`) &&
					!strings.Contains(server, `more_set_headers "Foo: Bar";`) &&
					!strings.Contains(server, `more_set_headers "Xpto: Lalala";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Globalfoo", []string{"Foooo"}).
			NotContainsKey("Foo").
			NotContainsKey("Xpto")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", hostAnnots).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Globalfoo", []string{"Foooo"}).
			NotContainsKey("Foo").
			NotContainsKey("Xpto")
	})
})
