/*
Copyright 2023 The Kubernetes Authors.

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

var _ = framework.DescribeSetting("aio-write", func() {
	f := framework.NewDefaultFramework("aio-write")

	ginkgo.It("should be enabled by default", func() {
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "aio_write on")
			})
	})

	ginkgo.It("should be enabled when setting is true", func() {
		f.UpdateNginxConfigMapData("enable-aio-write", "true")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "aio_write on")
			})
	})

	ginkgo.It("should be disabled when setting is false", func() {
		f.UpdateNginxConfigMapData("enable-aio-write", "false")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "aio_write on")
			})
	})
})
