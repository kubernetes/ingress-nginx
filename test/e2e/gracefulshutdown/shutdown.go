/*
Copyright 2019 The Kubernetes Authors.

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
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Shutdown] ingress controller", func() {
	f := framework.NewDefaultFramework("shutdown-ingress-controller")

	host := "shutdown"

	ginkgo.BeforeEach(func() {
		f.UpdateNginxConfigMapData("worker-shutdown-timeout", "600s")
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("should shutdown in less than 60 secons without pending connections", func() {
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name shutdown")
			})

		f.HTTPTestClient().
			GET("/sleep/1").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		startTime := time.Now()

		f.ScaleDeploymentToZero("nginx-ingress-controller")

		assert.LessOrEqual(ginkgo.GinkgoT(), int(time.Since(startTime).Seconds()), 60, "waiting shutdown")
	})
	/* @rikatz - Removing this tests as they are failing in GH Actions but not locally.
	ginkgo.It("should shutdown after waiting 60 seconds for pending connections to be closed", func(done ginkgo.Done) {
		defer close(done)

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			grace := int64(3600)
			deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &grace
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment")

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-send-timeout": "600",
			"nginx.ingress.kubernetes.io/proxy-read-timeout": "600",
		}
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, annotations))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name shutdown")
			})

		startTime := time.Now()

		result := make(chan int)
		go func() {
			defer ginkgo.GinkgoRecover()
			resp := f.HTTPTestClient().
				GET("/sleep/70").
				WithHeader("Host", host).
				Expect().
				Raw()

			result <- resp.StatusCode
		}()

		framework.Sleep(1 * time.Second)

		f.ScaleDeploymentToZero("nginx-ingress-controller")

		statusCode := <-result
		assert.Equal(ginkgo.GinkgoT(), http.StatusOK, statusCode, "expecting a valid response from HTTP request")
		assert.GreaterOrEqual(ginkgo.GinkgoT(), int(time.Since(startTime).Seconds()), 60, "waiting shutdown")
	}, 100)

	ginkgo.It("should shutdown after waiting 150 seconds for pending connections to be closed", func(done ginkgo.Done) {
		defer close(done)

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			grace := int64(3600)
			deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &grace
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-send-timeout": "600",
			"nginx.ingress.kubernetes.io/proxy-read-timeout": "600",
		}
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, annotations))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name shutdown")
			})

		startTime := time.Now()

		result := make(chan int)
		go func() {
			defer ginkgo.GinkgoRecover()
			resp := f.HTTPTestClient().
				GET("/sleep/150").
				WithHeader("Host", host).
				Expect().
				Raw()

			result <- resp.StatusCode
		}()

		framework.Sleep(1 * time.Second)

		f.ScaleDeploymentToZero("nginx-ingress-controller")

		statusCode := <-result
		assert.Equal(ginkgo.GinkgoT(), http.StatusOK, statusCode, "expecting a valid response from HTTP request")
		assert.GreaterOrEqual(ginkgo.GinkgoT(), int(time.Since(startTime).Seconds()), 150, "waiting shutdown")
	}, 200) */
})
