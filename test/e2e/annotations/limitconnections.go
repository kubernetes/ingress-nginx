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

package annotations

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("Annotation - limit-connections", func() {
	f := framework.NewDefaultFramework("limit-connections")

	ginkgo.BeforeEach(func() {
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("should limit-connections", func() {
		host := "limit-connections"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, nil)
		f.EnsureIngress(ing)
		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
		})

		// prerequisite
		configKey := "variables-hash-bucket-size"
		configValue := "256"
		f.UpdateNginxConfigMapData(configKey, configValue)
		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("variables_hash_bucket_size %s;", configValue))
		})

		// limit connections
		annotations := make(map[string]string)
		connectionLimit := 8
		annotations["nginx.ingress.kubernetes.io/limit-connections"] = strconv.Itoa(connectionLimit)
		ing.SetAnnotations(annotations)
		f.UpdateIngress(ing)
		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, "limit_conn") && strings.Contains(server, fmt.Sprintf("conn %v;", connectionLimit))
		})

		type asyncResult struct {
			index  int
			status int
		}

		response := make(chan *asyncResult)
		maxRequestNumber := 20
		for currentRequestNumber := 0; currentRequestNumber < maxRequestNumber; currentRequestNumber++ {
			go func(requestNumber int, responeChannel chan *asyncResult) {
				resp := f.HTTPTestClient().
					GET("/sleep/10").
					WithHeader("Host", host).
					Expect().
					Raw()

				code := 0
				if resp != nil {
					code = resp.StatusCode
				}
				responeChannel <- &asyncResult{requestNumber, code}
			}(currentRequestNumber, response)

		}
		var results []asyncResult
		for {
			res := <-response
			results = append(results, *res)
			if len(results) >= maxRequestNumber {
				break
			}
		}

		close(response)

		okRequest := 0
		failedRequest := 0
		requestError := 0
		expectedFail := maxRequestNumber - connectionLimit
		for _, result := range results {
			if result.status == 200 {
				okRequest++
			} else if result.status == 503 {
				failedRequest++
			} else {
				requestError++
			}
		}
		assert.Equal(ginkgo.GinkgoT(), okRequest, connectionLimit, "expecting the ok (200) requests number should equal the connection limit")
		assert.Equal(ginkgo.GinkgoT(), failedRequest, expectedFail, "expecting the failed (503) requests number should equal the expected to fail request number")
		assert.Equal(ginkgo.GinkgoT(), requestError, 0, "expecting the failed (other than 503) requests number should equal zero")
	})
})
