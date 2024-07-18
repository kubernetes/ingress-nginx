/*
Copyright 2023 The Kubernetes Authors.

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

package defaultbackend

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

const waitForMetrics = 2 * time.Second

var _ = framework.IngressNginxDescribe("[metrics] exported prometheus metrics", func() {
	f := framework.NewDefaultFramework("metrics")
	host := "foo.com"
	wildcardHost := "wildcard." + host

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return (strings.Contains(server, fmt.Sprintf("server_name %s;", host)) ||
					strings.Contains(server, fmt.Sprintf("server_name %s ;", host))) &&
					strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
	})

	ginkgo.It("exclude socket request metrics are absent", func() {
		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--exclude-socket-metrics=nginx_ingress_controller_request_size,nginx_ingress_controller_header_duration_seconds")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating deployment")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
		time.Sleep(waitForMetrics)

		ip := f.GetNginxPodIP()
		mf, err := f.GetMetric("nginx_ingress_controller_request_size", ip)
		assert.ErrorContains(ginkgo.GinkgoT(), err, "nginx_ingress_controller_request_size")
		assert.Nil(ginkgo.GinkgoT(), mf)
	})
	ginkgo.It("exclude socket request metrics are present", func() {
		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--exclude-socket-metrics=non_existing_metric_does_not_affect_existing_metrics")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating deployment")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
		time.Sleep(waitForMetrics)

		ip := f.GetNginxPodIP()
		mf, err := f.GetMetric("nginx_ingress_controller_request_size", ip)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.NotNil(ginkgo.GinkgoT(), mf)
	})
	ginkgo.It("request metrics per undefined host are present when flag is set", func() {
		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--metrics-per-undefined-host=true")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating deployment")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", wildcardHost).
			Expect().
			Status(http.StatusNotFound)
		time.Sleep(waitForMetrics)

		ip := f.GetNginxPodIP()
		reqMetrics, err := f.GetMetric("nginx_ingress_controller_requests", ip)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.NotNil(ginkgo.GinkgoT(), reqMetrics.Metric)
		assert.Len(ginkgo.GinkgoT(), reqMetrics.Metric, 1)

		containedLabel := false
		for _, label := range reqMetrics.Metric[0].Label {
			if *label.Name == "host" && *label.Value == wildcardHost {
				containedLabel = true
				break
			}
		}

		assert.Truef(ginkgo.GinkgoT(), containedLabel, "expected reqMetrics to contain label with \"name\"=\"host\" \"value\"=%q, but it did not: %s", wildcardHost, reqMetrics.String())
	})
	ginkgo.It("request metrics per undefined host are not present when flag is not set", func() {
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", wildcardHost).
			Expect().
			Status(http.StatusNotFound)
		time.Sleep(waitForMetrics)

		ip := f.GetNginxPodIP()
		reqMetrics, err := f.GetMetric("nginx_ingress_controller_requests", ip)
		assert.EqualError(ginkgo.GinkgoT(), err, "there is no metric with name nginx_ingress_controller_requests")
		assert.Nil(ginkgo.GinkgoT(), reqMetrics)
	})
})
