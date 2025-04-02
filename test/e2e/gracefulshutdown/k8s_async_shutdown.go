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

package gracefulshutdown

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Shutdown] Asynchronous shutdown", func() {
	f := framework.NewDefaultFramework("k8s-async-shutdown", func(f *framework.Framework) {
		f.Namespace = "k8s-async-shutdown"
	})

	host := "async-shutdown"

	ginkgo.BeforeEach(func() {
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("should not shut down while still receiving traffic", func() {
		defer ginkgo.GinkgoRecover()

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			// Note: e2e's default terminationGracePeriodSeconds is 1 for some reason, so extend it
			grace := int64(300)
			deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &grace
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name "+host)
			})

		// We need to get pod IP first because after the pod becomes terminating,
		// it is removed from Service endpoints, and becomes unable to be discovered by "f.HTTPTestClient()".
		ip := f.GetNginxPodIP()

		// Assume that the upstream takes 30 seconds to update its endpoints,
		// therefore we are still receiving traffic while shutting down
		go func() {
			defer ginkgo.GinkgoRecover()
			for i := 0; i < 30; i++ {
				f.HTTPDumbTestClient().
					GET("/").
					WithURL(fmt.Sprintf("http://%s/", ip)).
					WithHeader("Host", host).
					Expect().
					Status(http.StatusOK)

				framework.Sleep(time.Second)
			}
		}()

		start := time.Now()
		f.ScaleDeploymentToZero("nginx-ingress-controller")
		assert.GreaterOrEqualf(ginkgo.GinkgoT(), int(time.Since(start).Seconds()), 35,
			"should take more than 30 + 5 seconds for graceful shutdown")
	})
})
