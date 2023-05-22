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
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var _ = framework.DescribeSetting("[Load Balancer] least-connections", func() {
	f := framework.NewDefaultFramework("leastconn")
	ginkgo.BeforeEach(func() {
		f.NewAlwaysSlowEchoDeploymentWithOptions(100, framework.WithDeploymentReplicas(1), framework.WithName("echo-fast"), framework.WithServiceName("leastconn-slow"))
		f.NewAlwaysSlowEchoDeploymentWithOptions(1800, framework.WithDeploymentReplicas(1), framework.WithName("echo-slow"), framework.WithServiceName("leastconn-slow"))

		f.SetNginxConfigMapData(map[string]string{
			"worker-processes": "2",
			"load-balance":     "least_connections"},
		)
	})

	ginkgo.It("does not fail requests", func() {
		host := "load-balance.com"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, "leastconn-slow", 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name load-balance.com")
			})

		algorithm, err := f.GetLbAlgorithm("leastconn-slow", 80)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Equal(ginkgo.GinkgoT(), "least_connections", algorithm)

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

	ginkgo.It("sends fewer requests to a slower server", func() {
		host := "load-balance.com"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, "leastconn-slow", 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name load-balance.com")
			})

		algorithm, err := f.GetLbAlgorithm("leastconn-slow", 80)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Equal(ginkgo.GinkgoT(), "least_connections", algorithm)

		re, _ := regexp.Compile(fmt.Sprintf(`%v.*`, framework.EchoService))
		replicaRequestCount := map[string]int{}
		reqCount := 30

		var wg sync.WaitGroup
		wg.Add(reqCount)
		results := make(chan string, reqCount)
		for i := 0; i < reqCount; i++ {
			time.Sleep(100 * time.Millisecond)
			go func() {
				defer wg.Done()
				body := f.HTTPTestClient().
					GET("/").
					WithHeader("Host", host).
					Expect().
					Status(http.StatusOK).Body().Raw()
				replica := re.FindString(body)
				assert.NotEmpty(ginkgo.GinkgoT(), replica)
				results <- replica
			}()
		}
		wg.Wait()
		close(results)

		for r := range results {
			if _, ok := replicaRequestCount[r]; !ok {
				replicaRequestCount[r] = 1
			} else {
				replicaRequestCount[r]++
			}
		}

		framework.Logf("Request distribution: %v", replicaRequestCount)

		replicaCount := len(replicaRequestCount)
		assert.Equal(ginkgo.GinkgoT(), replicaCount, 2, "expected responses from two replicas")

		values := make([]int, 2)
		i := 0
		for _, v := range replicaRequestCount {
			values[i] = v
			i++
		}
		sort.Ints(values)
		// we expect to see at least twice as many requests to the echo server compared to the slow echo server
		assert.GreaterOrEqual(ginkgo.GinkgoT(), values[1], 2*values[0], "expected at least twice as many responses from the faster server")
	})
})
