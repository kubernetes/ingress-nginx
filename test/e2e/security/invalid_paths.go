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

package security

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	validPath   = "/xpto/~user/t-e_st.exe"
	invalidPath = "/foo/bar/;xpto"
	regexPath   = "/foo/bar/(.+)"
	host        = "securitytest.com"
)

var (
	annotationRegex = map[string]string{
		"nginx.ingress.kubernetes.io/use-regex": "true",
	}
)

var _ = framework.IngressNginxDescribe("[Security] validate path fields", func() {
	f := framework.NewDefaultFramework("validate-path")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should accept an ingress with valid path", func() {

		ing := framework.NewSingleIngress(host, validPath, host, f.Namespace, framework.EchoService, 80, nil)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.HTTPTestClient().
			GET(validPath).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should drop an ingress with invalid path", func() {

		ing := framework.NewSingleIngress(host, invalidPath, host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.HTTPTestClient().
			GET(invalidPath).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})

	ginkgo.It("should drop an ingress with regex path and regex disabled", func() {

		ing := framework.NewSingleIngress(host, regexPath, host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.HTTPTestClient().
			GET("/foo/bar/lalala").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})

	ginkgo.It("should accept an ingress with regex path and regex enabled", func() {

		ing := framework.NewSingleIngress(host, regexPath, host, f.Namespace, framework.EchoService, 80, annotationRegex)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.HTTPTestClient().
			GET("/foo/bar/lalala").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should reject an ingress with invalid path and regex enabled", func() {

		ing := framework.NewSingleIngress(host, invalidPath, host, f.Namespace, framework.EchoService, 80, annotationRegex)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		f.HTTPTestClient().
			GET(invalidPath).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusNotFound)
	})
})
