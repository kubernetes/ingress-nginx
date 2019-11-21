/*
Copyright 2016 The Kubernetes Authors.

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("LuaSharedDict", func() {
	f := framework.NewDefaultFramework("lua-shared-dicts")
	host := "lua-shared-dicts"

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("configures lua shared dicts", func() {
		ingress := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ingress)

		f.UpdateNginxConfigMapData("lua-shared-dicts", "configuration_data:60,certificate_data:300, my_dict: 15 , invalid: 1a")

		ngxCfg := ""
		f.WaitForNginxConfiguration(func(cfg string) bool {
			ngxCfg = cfg
			return true
		})

		Expect(ngxCfg).Should(ContainSubstring("lua_shared_dict configuration_data 60M;"))
		Expect(ngxCfg).Should(ContainSubstring("lua_shared_dict certificate_data 20M;"))
		Expect(ngxCfg).Should(ContainSubstring("lua_shared_dict my_dict 15M;"))
	})
})
