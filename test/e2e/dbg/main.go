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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Debug Tool", func() {
	f := framework.NewDefaultFramework("debug-tool")
	host := "foo.com"

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	AfterEach(func() {
	})

	It("should list the backend servers", func() {
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return Expect(cfg).Should(ContainSubstring(host))
		})

		cmd := "/dbg backends list"
		output, err := f.ExecIngressPod(cmd)
		Expect(err).Should(BeNil())

		// Should be 2: the default and the echo deployment
		numUpstreams := len(strings.Split(strings.Trim(string(output), "\n"), "\n"))
		Expect(numUpstreams).Should(Equal(2))

	})

	It("should get information for a specific backend server", func() {
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return Expect(cfg).Should(ContainSubstring(host))
		})

		cmd := "/dbg backends list"
		output, err := f.ExecIngressPod(cmd)
		Expect(err).Should(BeNil())

		backends := strings.Split(string(output), "\n")
		Expect(len(backends)).Should(BeNumerically(">", 0))

		getCmd := "/dbg backends get " + backends[0]
		output, err = f.ExecIngressPod(getCmd)

		var f map[string]interface{}
		unmarshalErr := json.Unmarshal([]byte(output), &f)
		Expect(unmarshalErr).Should(BeNil())

		// Check that the backend we've gotten has the same name as the one we requested
		Expect(backends[0]).Should(Equal(f["name"].(string)))
	})

	It("should produce valid JSON for /dbg general", func() {
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)

		cmd := "/dbg general"
		output, err := f.ExecIngressPod(cmd)
		Expect(err).Should(BeNil())

		var f interface{}
		unmarshalErr := json.Unmarshal([]byte(output), &f)
		Expect(unmarshalErr).Should(BeNil())
	})
})
