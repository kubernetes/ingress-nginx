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

package dbg

import (
	"encoding/json"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Debug CLI", func() {
	f := framework.NewDefaultFramework("debug-tool")
	host := "foo.com"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should list the backend servers", func() {
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, host)
		})

		cmd := "/dbg backends list"
		output, err := f.ExecIngressPod(cmd)
		assert.Nil(ginkgo.GinkgoT(), err)

		// Should be 2: the default and the echo deployment
		numUpstreams := len(strings.Split(strings.Trim(string(output), "\n"), "\n"))
		assert.Equal(ginkgo.GinkgoT(), numUpstreams, 2)
	})

	ginkgo.It("should get information for a specific backend server", func() {
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, host)
		})

		cmd := "/dbg backends list"
		output, err := f.ExecIngressPod(cmd)
		assert.Nil(ginkgo.GinkgoT(), err)

		backends := strings.Split(string(output), "\n")
		assert.Greater(ginkgo.GinkgoT(), len(backends), 0)

		getCmd := "/dbg backends get " + backends[0]
		output, err = f.ExecIngressPod(getCmd)
		assert.Nil(ginkgo.GinkgoT(), err)

		var f map[string]interface{}
		err = json.Unmarshal([]byte(output), &f)
		assert.Nil(ginkgo.GinkgoT(), err)

		// Check that the backend we've gotten has the same name as the one we requested
		assert.Equal(ginkgo.GinkgoT(), backends[0], f["name"].(string))
	})

	ginkgo.It("should produce valid JSON for /dbg general", func() {
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		cmd := "/dbg general"
		output, err := f.ExecIngressPod(cmd)
		assert.Nil(ginkgo.GinkgoT(), err)

		var f interface{}
		err = json.Unmarshal([]byte(output), &f)
		assert.Nil(ginkgo.GinkgoT(), err)
	})
})
