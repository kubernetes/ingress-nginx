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

package settings

import (
	"context"
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Flag] disable-sync-events", func() {
	f := framework.NewDefaultFramework("disable-sync-events")

	ginkgo.It("should create sync events (default)", func() {
		host := "sync-events-default"
		f.NewEchoDeployment(framework.WithDeploymentReplicas(1))

		ing := framework.NewSingleIngressWithIngressClass(host, "/", host, f.Namespace, framework.EchoService, f.IngressClass, 80, nil)
		ing = f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		events, err := f.KubeClientSet.CoreV1().Events(ing.Namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: "reason=Sync,involvedObject.name=" + host})
		assert.Nil(ginkgo.GinkgoT(), err, "listing events")

		assert.NotEmpty(ginkgo.GinkgoT(), events.Items, "got events")
	})

	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should create sync events", func() {
		host := "disable-sync-events-false"
		f.NewEchoDeployment(framework.WithDeploymentReplicas(1))

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--disable-sync-events=false")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		ing := framework.NewSingleIngressWithIngressClass(host, "/", host, f.Namespace, framework.EchoService, f.IngressClass, 80, nil)
		ing = f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		events, err := f.KubeClientSet.CoreV1().Events(ing.Namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: "reason=Sync,involvedObject.name=" + host})
		assert.Nil(ginkgo.GinkgoT(), err, "listing events")

		assert.NotEmpty(ginkgo.GinkgoT(), events.Items, "got events")
	})

	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should not create sync events", func() {
		host := "disable-sync-events-true"
		f.NewEchoDeployment(framework.WithDeploymentReplicas(1))

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--disable-sync-events=true")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		ing := framework.NewSingleIngressWithIngressClass(host, "/", host, f.Namespace, framework.EchoService, f.IngressClass, 80, nil)
		ing = f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		events, err := f.KubeClientSet.CoreV1().Events(ing.Namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: "reason=Sync,involvedObject.name=" + host})
		assert.Nil(ginkgo.GinkgoT(), err, "listing events")

		assert.Empty(ginkgo.GinkgoT(), events.Items, "got events")
	})
})
