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

package annotations

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

func startIngress(f *framework.Framework, annotations *map[string]string) map[string]bool {
	host := "upstream-hash-by.foo.com"

	ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, annotations)
	f.EnsureIngress(ing)
	f.WaitForNginxServer(host,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
		})

	err := wait.PollImmediate(framework.Poll, framework.DefaultTimeout, func() (bool, error) {
		resp, _, _ := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()
		if resp.StatusCode == http.StatusOK {
			return true, nil
		}
		return false, nil
	})
	Expect(err).Should(BeNil())

	re, _ := regexp.Compile(`Hostname: http-svc.*`)
	podMap := map[string]bool{}

	for i := 0; i < 100; i++ {
		_, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())

		podName := re.FindString(body)
		Expect(podName).ShouldNot(Equal(""))
		podMap[podName] = true

	}

	return podMap
}

var _ = framework.IngressNginxDescribe("Annotations - UpstreamHashBy", func() {
	f := framework.NewDefaultFramework("upstream-hash-by")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(6)
	})

	AfterEach(func() {
	})

	It("should connect to the same pod", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-hash-by": "$request_uri",
		}

		podMap := startIngress(f, &annotations)
		Expect(len(podMap)).Should(Equal(1))

	})

	It("should connect to the same subset of pods", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-hash-by":             "$request_uri",
			"nginx.ingress.kubernetes.io/upstream-hash-by-subset":      "true",
			"nginx.ingress.kubernetes.io/upstream-hash-by-subset-size": "3",
		}

		podMap := startIngress(f, &annotations)
		Expect(len(podMap)).Should(Equal(3))

	})
})
