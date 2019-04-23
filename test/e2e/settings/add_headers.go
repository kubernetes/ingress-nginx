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

	It("should escape double quotes in the header values", func() {
		host := "escape"
		addHeadersConfigMap := "add-headers-1"
		proxySetHeadersConfigMap := "p-h-1"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		f.CreateConfigMap(addHeadersConfigMap, map[string]string{
			"NEL": `{"report_to":"network-errors","failure_fraction":0.01,"success_fraction":0.0001}`,
		})

		f.CreateConfigMap(proxySetHeadersConfigMap, map[string]string{
			"Quux": `{"ack":["quack"]}`,
		})

		f.UpdateNginxConfigMapData("add-headers", f.Namespace+"/"+addHeadersConfigMap)
		f.UpdateNginxConfigMapData("proxy-set-headers", f.Namespace+"/"+proxySetHeadersConfigMap)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `add_header NEL "{\"report_to\":\"network-errors\",\"failure_fraction\":0.01,\"success_fraction\":0.0001}";`) &&
					strings.Contains(cfg, `proxy_set_header Quux "{\"ack\":[\"quack\"]}"`)
			})
	})
})
