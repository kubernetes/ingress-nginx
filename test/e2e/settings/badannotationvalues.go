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
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("Bad annotation values", func() {
	f := framework.NewDefaultFramework("bad-annotation")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("[BAD_ANNOTATIONS] should drop an ingress if there is an invalid character in some annotation", func() {
		host := "invalid-value-test"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `
			# abc { }`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.UpdateNginxConfigMapData("allow-snippet-annotations", "true")
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "something_forbidden,otherthing_forbidden,{")

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "# abc { }")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})

	ginkgo.It("[BAD_ANNOTATIONS] should drop an ingress if there is a forbidden word in some annotation", func() {
		host := "forbidden-value-test"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `
			  default_type text/plain;
              content_by_lua_block {
                ngx.say("Hello World")
            }`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.UpdateNginxConfigMapData("allow-snippet-annotations", "true")
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "something_forbidden,otherthing_forbidden,content_by_lua_block")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, `ngx.say("Hello World")`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})

	ginkgo.It("[BAD_ANNOTATIONS] should allow an ingress if there is a default blocklist config in place", func() {

		hostValid := "custom-allowed-value-test"
		annotationsValid := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `
			# bla_by_lua`,
		}

		ingValid := framework.NewSingleIngress(hostValid, "/", hostValid, f.Namespace, framework.EchoService, 80, annotationsValid)

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		f.EnsureIngress(ingValid)

		f.WaitForNginxServer(hostValid,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", hostValid))
			})

		f.WaitForNginxServer(hostValid,
			func(server string) bool {
				return strings.Contains(server, "# bla_by_lua")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", hostValid).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("[BAD_ANNOTATIONS] should drop an ingress if there is a custom blocklist config in place and allow others to pass", func() {
		host := "custom-forbidden-value-test"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `
			# something_forbidden`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "something_forbidden,otherthing_forbidden")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "# something_forbidden")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)

	})
})
