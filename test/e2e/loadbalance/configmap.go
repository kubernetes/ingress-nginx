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

package loadbalance

import (
	"encoding/json"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	waitForLuaSync = 5 * time.Second
)

var _ = framework.IngressNginxDescribe("Load Balance - Configmap value", func() {
	f := framework.NewDefaultFramework("lb-configmap")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	AfterEach(func() {
	})

	It("should apply the configmap load-balance setting", func() {
		host := "load-balance.com"

		f.UpdateNginxConfigMapData("load-balance", "ewma")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name load-balance.com")
			})
		time.Sleep(waitForLuaSync)

		getCmd := "/dbg backends all"
		output, err := f.ExecIngressPod(getCmd)
		Expect(err).Should(BeNil())

		var backends []map[string]interface{}
		unmarshalErr := json.Unmarshal([]byte(output), &backends)
		Expect(unmarshalErr).Should(BeNil())

		for _, backend := range backends {
			if backend["name"].(string) != "upstream-default-backend" {
				lb, ok := backend["load-balance"].(string)
				Expect(ok).Should(Equal(true))
				Expect(lb).Should(Equal("ewma"))
			}
		}
	})
})
