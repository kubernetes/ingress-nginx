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

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("main-snippet", func() {
	f := framework.NewDefaultFramework("main-snippet")
	if framework.IsCrossplane() {
		return
	}
	mainSnippet := "main-snippet"

	ginkgo.It("should add value of main-snippet setting to nginx config", func() {
		expectedComment := "# main snippet"
		f.UpdateNginxConfigMapData(mainSnippet, expectedComment)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedComment)
			})
	})
})
