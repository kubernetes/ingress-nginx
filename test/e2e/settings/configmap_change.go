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
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Configmap change", func() {
	f := framework.NewDefaultFramework("configmap-change")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should reload after an update in the configuration", func() {
		host := "configmap-change"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		ginkgo.By("adding a whitelist-source-range")

		f.UpdateNginxConfigMapData("whitelist-source-range", "1.1.1.1")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "allow 1.1.1.1;")
			})

		ginkgo.By("changing error-log-level")

		f.UpdateNginxConfigMapData("error-log-level", "debug")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.ContainsAny(cfg, "error_log  /var/log/nginx/error.log debug;") ||
					strings.ContainsAny(cfg, "error_log /var/log/nginx/error.log debug;")
			})

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, "Backend successfully reloaded")
	})
})
