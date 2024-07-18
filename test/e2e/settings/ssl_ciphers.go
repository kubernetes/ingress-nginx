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

var _ = framework.DescribeSetting("ssl-ciphers", func() {
	f := framework.NewDefaultFramework("ssl-ciphers")

	ginkgo.It("Add ssl ciphers", func() {
		wlKey := "ssl-ciphers"
		wlValue := "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP"

		f.UpdateNginxConfigMapData(wlKey, wlValue)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, fmt.Sprintf("ssl_ciphers '%s';", wlValue)) || strings.Contains(cfg, fmt.Sprintf("ssl_ciphers %s;", wlValue))
		})
	})
})
