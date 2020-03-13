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

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("[Security] modsecurity-snippet", func() {
	f := framework.NewDefaultFramework("modsecurity-snippet")

	ginkgo.It("should add value of modsecurity-snippet setting to nginx config", func() {
		expectedComment := "# modsecurity snippet"

		f.SetNginxConfigMapData(map[string]string{
			"enable-modsecurity":  "true",
			"modsecurity-snippet": expectedComment,
		})

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedComment)
			})
	})
})
