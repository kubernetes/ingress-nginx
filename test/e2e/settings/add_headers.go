/*
Copyright 2018 The Kubernetes Authors.

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
	"strings"

	. "github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("add-headers and proxy-set-headers", func() {
	f := framework.NewDefaultFramework("add-headers")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should reload after an update to the add-headers configmap", func() {
		host := "configmap-change"
		addHeadersConfigMap := "add-headers-1"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		f.CreateConfigMap(addHeadersConfigMap, map[string]string{
			"Foo": "bar",
		})

		f.UpdateNginxConfigMapData("add-headers", f.Namespace+"/"+addHeadersConfigMap)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `add_header Foo "bar";`)
			})

		f.SetConfigMapData(addHeadersConfigMap, map[string]string{
			"Foo": "baz",
			"Rar": "quux",
		})

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `add_header Foo "baz";`) &&
					strings.Contains(cfg, `add_header Rar "quux";`)
			})
	})

	It("should reload when the add-headers configmap key changes", func() {
		host := "configmap-key-change"
		addHeadersConfigMap := "add-headers-1"
		otheraddHeadersConfigMap := "add-headers-2"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		f.UpdateNginxConfigMapData("add-headers", f.Namespace+"/"+addHeadersConfigMap)

		f.CreateConfigMap(addHeadersConfigMap, map[string]string{
			"Hi": "ho",
		})

		f.CreateConfigMap(otheraddHeadersConfigMap, map[string]string{
			"Ho": "hum",
		})

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `add_header Hi "ho";`)
			})

		f.UpdateNginxConfigMapData("add-headers", f.Namespace+"/"+otheraddHeadersConfigMap)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `add_header Ho "hum";`)
			})
	})
})
