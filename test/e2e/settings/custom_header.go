/*
Copyright 2020 The Kubernetes Authors.

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
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("add-headers", func() {
	f := framework.NewDefaultFramework("custom-header")
	host := "custom-header"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)
	})

	ginkgo.It("Add a custom header", func() {
		customHeader := "X-A-Custom-Header"
		customHeaderValue := "customHeaderValue"

		h := make(map[string]string)
		h[customHeader] = customHeaderValue

		cfgMap := "add-headers-configmap"

		f.CreateConfigMap(cfgMap, h)

		f.UpdateNginxConfigMapData("add-headers", fmt.Sprintf("%v/%v", f.Namespace, cfgMap))

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("more_set_headers \"%s: %s\";", customHeader, customHeaderValue))
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header(customHeader).Contains(customHeaderValue)
	})

	ginkgo.It("Add multiple custom headers", func() {
		firstCustomHeader := "X-First"
		firstCustomHeaderValue := "Prepare for trouble!"
		secondCustomHeader := "X-Second"
		secondCustomHeaderValue := "And make it double!"

		h := make(map[string]string)
		h[firstCustomHeader] = firstCustomHeaderValue
		h[secondCustomHeader] = secondCustomHeaderValue

		cfgMap := "add-headers-configmap-two"

		f.CreateConfigMap(cfgMap, h)

		f.UpdateNginxConfigMapData("add-headers", fmt.Sprintf("%v/%v", f.Namespace, cfgMap))

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("more_set_headers \"%s: %s\";", firstCustomHeader, firstCustomHeaderValue)) &&
				strings.Contains(server, fmt.Sprintf("more_set_headers \"%s: %s\";", secondCustomHeader, secondCustomHeaderValue))
		})

		resp := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Raw()

		assert.Equal(ginkgo.GinkgoT(), resp.Header.Get(firstCustomHeader), firstCustomHeaderValue)
		assert.Equal(ginkgo.GinkgoT(), resp.Header.Get(secondCustomHeader), secondCustomHeaderValue)
	})
})
