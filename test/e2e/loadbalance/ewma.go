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
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("[Load Balancer] EWMA", func() {
	f := framework.NewDefaultFramework("ewma")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(3)
		f.SetNginxConfigMapData(map[string]string{
			"worker-processes": "2",
			"load-balance":     "ewma"},
		)
	})

	ginkgo.It("does not fail requests", func() {
		host := "load-balance.com"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name load-balance.com")
			})

		algorithm, err := f.GetLbAlgorithm(framework.EchoService, 80)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Equal(ginkgo.GinkgoT(), algorithm, "ewma")

		re, _ := regexp.Compile(fmt.Sprintf(`%v.*`, framework.EchoService))
		replicaRequestCount := map[string]int{}

		for i := 0; i < 30; i++ {
			body := f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).Body().Raw()

			replica := re.FindString(body)
			assert.NotEmpty(ginkgo.GinkgoT(), replica)

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
		assert.Equal(ginkgo.GinkgoT(), actualCount, 30)
	})
})
