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

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Ingress] [PathType] mix Exact and Prefix paths", func() {
	f := framework.NewDefaultFramework("mixed")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	exactPathType := networking.PathTypeExact

	ginkgo.It("should choose the correct location", func() {
		host := "mixed.path"

		f.UpdateNginxConfigMapData("global-allowed-response-headers", "Pathtype,Pathheader")

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-headers": f.Namespace + "/custom-headers-exact",
		}

		f.CreateConfigMap("custom-headers-exact", map[string]string{
			"Pathtype":   "exact",
			"Pathheader": "/",
		})

		ing := framework.NewSingleIngress("exact-root", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &exactPathType
		f.EnsureIngress(ing)

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/custom-headers": f.Namespace + "/custom-headers-prefix",
		}

		f.CreateConfigMap("custom-headers-prefix", map[string]string{
			"Pathtype":   "prefix",
			"Pathheader": "/",
		})

		ing = framework.NewSingleIngress("prefix-root", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location = /") &&
					strings.Contains(server, "location /")
			})

		ginkgo.By("Checking exact request to /")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Pathtype", []string{"exact"}).ValueEqual("Pathheader", []string{"/"})

		ginkgo.By("Checking prefix request to /bar")
		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Pathtype", []string{"prefix"}).ValueEqual("Pathheader", []string{"/"})

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/custom-headers": f.Namespace + "/custom-headers-ex-foo",
		}

		f.CreateConfigMap("custom-headers-ex-foo", map[string]string{
			"Pathtype":   "exact",
			"Pathheader": "/foo",
		})
		ing = framework.NewSingleIngress("exact-foo", "/foo", host, f.Namespace, framework.EchoService, 80, annotations)
		ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &exactPathType
		f.EnsureIngress(ing)

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/custom-headers": f.Namespace + "/custom-headers-pr-foo",
		}

		f.CreateConfigMap("custom-headers-pr-foo", map[string]string{
			"Pathtype":   "prefix",
			"Pathheader": "/foo",
		})

		ing = framework.NewSingleIngress("prefix-foo", "/foo", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location = /foo") &&
					strings.Contains(server, "location /foo/")
			})

		ginkgo.By("Checking exact request to /foo")
		f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Pathtype", []string{"exact"}).ValueEqual("Pathheader", []string{"/foo"})

		ginkgo.By("Checking prefix request to /foo/bar")
		f.HTTPTestClient().
			GET("/foo/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Pathtype", []string{"prefix"}).ValueEqual("Pathheader", []string{"/foo"})

		ginkgo.By("Checking prefix request to /foobar")
		f.HTTPTestClient().
			GET("/foobar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Pathtype", []string{"prefix"}).ValueEqual("Pathheader", []string{"/"})

	})
})
