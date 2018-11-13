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
	"regexp"
	"strings"

	"github.com/parnurzeal/gorequest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Load Balance - Round Robin", func() {
	f := framework.NewDefaultFramework("round-robin")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(3)
		f.UpdateNginxConfigMapData("worker-processes", "1")
	})

	AfterEach(func() {
		f.UpdateNginxConfigMapData("worker-processes", "")
	})

	It("should evenly distribute requests with round-robin (default algorithm)", func() {
		host := "load-balance.com"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name load-balance.com")
			})

		re, _ := regexp.Compile(`http-svc.*`)
		replicaRequestCount := map[string]int{}

		for i := 0; i < 600; i++ {
			_, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
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

		for _, v := range replicaRequestCount {
			Expect(v).Should(Equal(200))
		}
	})
})
