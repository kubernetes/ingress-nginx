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
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Configmap - limit-rate", func() {
	f := framework.NewDefaultFramework("limit-rate")
	host := "limit-rate"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("Check limit-rate config", func() {
		annotation := make(map[string]string)
		annotation["nginx.ingress.kubernetes.io/proxy-buffering"] = "on"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotation)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
		})

		wlKey := "limit-rate"
		wlValue := "1"
		f.UpdateNginxConfigMapData(wlKey, wlValue)

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("limit_rate %sk;", wlValue))
		})

		wlValue = "90"
		f.UpdateNginxConfigMapData(wlKey, wlValue)

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("limit_rate %sk;", wlValue))
		})
	})
})
