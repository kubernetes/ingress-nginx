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

package annotations

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribeSerial("annotation validations", func() {
	f := framework.NewDefaultFramework("validations")
	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should allow ingress based on their risk on webhooks", func() {
		f.SetNginxConfigMapData(map[string]string{
			"allow-snippet-annotations": "true",
		})
		defer func() {
			f.SetNginxConfigMapData(map[string]string{
				"allow-snippet-annotations": "false",
			})
		}()

		host := "annotation-validations"

		// Low and Medium Risk annotations should be allowed, the rest should be denied
		f.UpdateNginxConfigMapData("annotations-risk-level", "Medium")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/default-backend":       "default/bla", // low risk
			"nginx.ingress.kubernetes.io/denylist-source-range": "1.1.1.1/32",  // medium risk
		}

		ginkgo.By("allow ingress with low/medium risk annotations")
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), ing, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress with allowed annotations should not trigger an error")

		ginkgo.By("block ingress with risky annotations")
		annotations["nginx.ingress.kubernetes.io/modsecurity-transaction-id"] = "bla123"      // High should be blocked
		annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = "some random stuff;" // High should be blocked
		ing = framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating ingress with risky annotations should trigger an error")
	})
	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should allow ingress based on their risk on webhooks", func() {
		f.SetNginxConfigMapData(map[string]string{
			"allow-snippet-annotations": "true",
		})
		defer func() {
			f.SetNginxConfigMapData(map[string]string{
				"allow-snippet-annotations": "false",
			})
		}()
		host := "annotation-validations"

		// Low and Medium Risk annotations should be allowed, the rest should be denied
		f.UpdateNginxConfigMapData("annotations-risk-level", "Medium")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/default-backend":       "default/bla", // low risk
			"nginx.ingress.kubernetes.io/denylist-source-range": "1.1.1.1/32",  // medium risk
		}

		ginkgo.By("allow ingress with low/medium risk annotations")
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), ing, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress with allowed annotations should not trigger an error")

		ginkgo.By("block ingress with risky annotations")
		annotations["nginx.ingress.kubernetes.io/modsecurity-transaction-id"] = "bla123"      // High should be blocked
		annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = "some random stuff;" // High should be blocked
		ing = framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating ingress with risky annotations should trigger an error")
	})
})
