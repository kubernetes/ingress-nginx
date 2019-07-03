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
	"regexp"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Load Balance - EWMA", func() {
	f := framework.NewDefaultFramework("ewma")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(3)
		f.UpdateNginxConfigMapData("worker-processes", "2")
		f.UpdateNginxConfigMapData("load-balance", "ewma")
	})

	It("does not fail requests", func() {
		host := "load-balance.com"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name load-balance.com")
			})
		time.Sleep(waitForLuaSync)

		algorithm, err := f.GetLbAlgorithm("http-svc", 80)
		Expect(err).Should(BeNil())
		Expect(algorithm).Should(Equal("ewma"))

		re, _ := regexp.Compile(`http-svc.*`)
		replicaRequestCount := map[string]int{}

		for i := 0; i < 30; i++ {
			_, body, errs := gorequest.New().
				Get(f.GetURL(framework.HTTP)).
				Set("Host", host).
				End()
			Expect(errs).Should(BeEmpty())

			replica := re.FindString(body)
			Expect(replica).ShouldNot(Equal(""))

			if _, ok := replicaRequestCount[replica]; !ok {
				replicaRequestCount[replica] = 1
			} else {
				replicaRequestCount[replica]++
			}
		}
		framework.Logf("Request distribution: %v", replicaRequestCount)

		actualCount := 0
		for _, v := range replicaRequestCount {
			actualCount += v
		}
		Expect(actualCount).Should(Equal(30))
	})
})
