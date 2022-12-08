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
	"net/http"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("settings-global-rate-limit", func() {
	f := framework.NewDefaultFramework("global-rate-limit")
	host := "global-rate-limit"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("generates correct NGINX configuration", func() {
		annotations := make(map[string]string)
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		ginkgo.By("generating correct defaults")

		ngxCfg := ""
		f.WaitForNginxConfiguration(func(cfg string) bool {
			if strings.Contains(cfg, "global_throttle") {
				ngxCfg = cfg
				return true
			}
			return false
		})

		assert.Contains(ginkgo.GinkgoT(), ngxCfg, fmt.Sprintf(`global_throttle = { `+
			`memcached = { host = "%v", port = %d, connect_timeout = %d, max_idle_timeout = %d, `+
			`pool_size = %d, }, status_code = %d, }`,
			"", 11211, 50, 10000, 50, 429))

		f.HTTPTestClient().GET("/").WithHeader("Host", host).Expect().Status(http.StatusOK)

		ginkgo.By("applying customizations")

		memcachedHost := "memc.default.svc.cluster.local"
		memcachedPort := 11211
		memcachedConnectTimeout := 100
		memcachedMaxIdleTimeout := 5000
		memcachedPoolSize := 100
		statusCode := 503

		f.SetNginxConfigMapData(map[string]string{
			"global-rate-limit-memcached-host":             memcachedHost,
			"global-rate-limit-memcached-port":             strconv.Itoa(memcachedPort),
			"global-rate-limit-memcached-connect-timeout":  strconv.Itoa(memcachedConnectTimeout),
			"global-rate-limit-memcached-max-idle-timeout": strconv.Itoa(memcachedMaxIdleTimeout),
			"global-rate-limit-memcached-pool-size":        strconv.Itoa(memcachedPoolSize),
			"global-rate-limit-status-code":                strconv.Itoa(statusCode),
		})

		ngxCfg = ""
		f.WaitForNginxConfiguration(func(cfg string) bool {
			if strings.Contains(cfg, "global_throttle") {
				ngxCfg = cfg
				return true
			}
			return false
		})

		assert.Contains(ginkgo.GinkgoT(), ngxCfg, fmt.Sprintf(`global_throttle = { `+
			`memcached = { host = "%v", port = %d, connect_timeout = %d, max_idle_timeout = %d, `+
			`pool_size = %d, }, status_code = %d, }`,
			memcachedHost, memcachedPort, memcachedConnectTimeout, memcachedMaxIdleTimeout,
			memcachedPoolSize, statusCode))

		f.HTTPTestClient().GET("/").WithHeader("Host", host).Expect().Status(http.StatusOK)
	})
})
