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

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Add no tls redirect locations", func() {
	f := framework.NewDefaultFramework("no-tls-redirect-locations")

	ginkgo.It("Check no tls redirect locations config", func() {
		host := "no-tls-redirect-locations"
		ing := framework.NewSingleIngress(host, "/check-no-tls", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(server string) bool {
			return !strings.Contains(server, fmt.Sprintf("force_no_ssl_redirect = true,"))
		})

		wlKey := "no-tls-redirect-locations"
		wlValue := "/check-no-tls"

		f.UpdateNginxConfigMapData(wlKey, wlValue)

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("force_no_ssl_redirect = true,"))
		})

	})
})
