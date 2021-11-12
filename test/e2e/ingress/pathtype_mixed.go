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

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Ingress] [PathType] mix Exact and Prefix paths", func() {
	f := framework.NewDefaultFramework("mixed")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	var exactPathType = networking.PathTypeExact

	ginkgo.It("should choose the correct location", func() {

		host := "mixed.path"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_input_headers "pathType: exact";more_set_input_headers "pathheader: /";`,
		}
		ing := framework.NewSingleIngress("exact-root", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &exactPathType
		f.EnsureIngress(ing)

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_input_headers "pathType: prefix";more_set_input_headers "pathheader: /";`,
		}
		ing = framework.NewSingleIngress("prefix-root", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location = /") &&
					strings.Contains(server, "location /")
			})

		ginkgo.By("Checking exact request to /")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.NotContains(ginkgo.GinkgoT(), body, "pathtype=prefix")
		assert.Contains(ginkgo.GinkgoT(), body, "pathtype=exact")
		assert.Contains(ginkgo.GinkgoT(), body, "pathheader=/")

		ginkgo.By("Checking prefix request to /bar")
		body = f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "pathtype=prefix")
		assert.NotContains(ginkgo.GinkgoT(), body, "pathtype=exact")
		assert.Contains(ginkgo.GinkgoT(), body, "pathheader=/")

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_input_headers "pathType: exact";more_set_input_headers "pathheader: /foo";`,
		}
		ing = framework.NewSingleIngress("exact-foo", "/foo", host, f.Namespace, framework.EchoService, 80, annotations)
		ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &exactPathType
		f.EnsureIngress(ing)

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_input_headers "pathType: prefix";more_set_input_headers "pathheader: /foo";`,
		}
		ing = framework.NewSingleIngress("prefix-foo", "/foo", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location = /foo") &&
					strings.Contains(server, "location /foo/")
			})

		ginkgo.By("Checking exact request to /foo")
		body = f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.NotContains(ginkgo.GinkgoT(), body, "pathtype=prefix")
		assert.Contains(ginkgo.GinkgoT(), body, "pathtype=exact")
		assert.Contains(ginkgo.GinkgoT(), body, "pathheader=/foo")

		ginkgo.By("Checking prefix request to /foo/bar")
		body = f.HTTPTestClient().
			GET("/foo/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "pathtype=prefix")
		assert.Contains(ginkgo.GinkgoT(), body, "pathheader=/foo")

		ginkgo.By("Checking prefix request to /foobar")
		body = f.HTTPTestClient().
			GET("/foobar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "pathtype=prefix")
		assert.Contains(ginkgo.GinkgoT(), body, "pathheader=/")
	})
})
