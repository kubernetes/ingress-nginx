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

package admission

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Serial] admission controller", func() {
	f := framework.NewDefaultFramework("admission")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("should not allow overlaps of host and paths without canary annotations", func() {
		host := "admission-test"

		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err := f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		secondIngress := framework.NewSingleIngress("second-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err = f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Create(context.TODO(), secondIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with the same host and path should return an error")

		err = uninstallChart(f)
		assert.Nil(ginkgo.GinkgoT(), err, "uninstalling helm chart")
	})

	ginkgo.It("should allow overlaps of host and paths with canary annotation", func() {
		host := "admission-test"

		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err := f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		canaryAnnotations := map[string]string{
			"nginx.ingress.kubernetes.io/canary":           "true",
			"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
		}
		secondIngress := framework.NewSingleIngress("second-ingress", "/", host, f.Namespace, framework.SlowEchoService, 80, canaryAnnotations)
		_, err = f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Create(context.TODO(), secondIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress with the same host and path should not return an error using a canary annotation")

		err = uninstallChart(f)
		assert.Nil(ginkgo.GinkgoT(), err, "uninstalling helm chart")
	})

	ginkgo.It("should return an error if there is an error validating the ingress definition", func() {
		host := "admission-test"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": "something invalid",
		}
		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err := f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid configuration should return an error")

		err = uninstallChart(f)
		assert.Nil(ginkgo.GinkgoT(), err, "uninstalling helm chart")
	})

})

func uninstallChart(f *framework.Framework) error {
	cmd := exec.Command("helm", "uninstall", "--namespace", f.Namespace, "nginx-ingress")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error uninstalling ingress-nginx release: %v", err)
	}

	return nil
}
