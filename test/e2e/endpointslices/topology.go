/*
Copyright 2022 The Kubernetes Authors.

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

package endpointslices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribeSerial("[TopologyHints] topology aware routing", func() {
	f := framework.NewDefaultFramework("topology")
	host := "topology-svc.foo.com"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(2), framework.WithSvcTopologyAnnotations())
	})

	ginkgo.It("should return 200 when service has topology hints", func() {

		annotations := make(map[string]string)
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %s", host))
		})

		ginkgo.By("checking if the service is reached")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		slices, err := f.KubeClientSet.DiscoveryV1().EndpointSlices(f.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "kubernetes.io/service-name=echo",
			Limit:         1,
		})
		assert.Nil(ginkgo.GinkgoT(), err)

		// check if we have hints, really depends on k8s endpoint slice controller
		gotHints := true
		for _, ep := range slices.Items[0].Endpoints {
			if ep.Hints == nil || len(ep.Hints.ForZones) == 0 {
				gotHints = false
				break
			}
		}

		curlCmd := fmt.Sprintf("curl --fail --silent http://localhost:%v/configuration/backends", nginx.StatusPort)
		status, err := f.ExecIngressPod(curlCmd)
		assert.Nil(ginkgo.GinkgoT(), err)
		var backends []map[string]interface{}
		json.Unmarshal([]byte(status), &backends)
		gotBackends := 0
		for _, bck := range backends {
			if strings.Contains(bck["name"].(string), "topology") {
				gotBackends = len(bck["endpoints"].([]interface{}))
			}
		}

		if gotHints {
			//we have 2 replics, if there is just one backend it means that we are routing according slices hints to same zone as controller is
			assert.Equal(ginkgo.GinkgoT(), 1, gotBackends)
		} else {
			// two replicas should have two endpoints without topology hints
			assert.Equal(ginkgo.GinkgoT(), 2, gotBackends)
		}
	})
})
