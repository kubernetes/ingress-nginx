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
	"strings"

	. "github.com/onsi/ginkgo"
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

	It("update lua shared dict", func() {
		ingress := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ingress)
		By("update shared dict")
		f.UpdateNginxConfigMapData("lua-shared-dicts", "configuration_data:123,certificate_data:456")
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "lua_shared_dict configuration_data 123M; lua_shared_dict certificate_data 456M;")
		})
	})
})
