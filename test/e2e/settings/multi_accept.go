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

var _ = framework.DescribeSetting("enable-multi-accept", func() {
	multiAccept := "enable-multi-accept"
	f := framework.NewDefaultFramework(multiAccept)

	ginkgo.It("should be enabled by default", func() {
		expectedDirective := "multi_accept on;"
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedDirective)
			})
	})

	ginkgo.It("should be enabled when set to true", func() {
		expectedDirective := "multi_accept on;"
		f.UpdateNginxConfigMapData(multiAccept, "true")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedDirective)
			})
	})

	ginkgo.It("should be disabled when set to false", func() {
		expectedDirective := "multi_accept off;"
		f.UpdateNginxConfigMapData(multiAccept, "false")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedDirective)
			})
	})
})
