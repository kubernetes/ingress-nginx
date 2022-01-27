/*
Copyright 2021 The Kubernetes Authors.

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
	"net/http"
	"strings"

	"github.com/gavv/httpexpect/v2"
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Flag] disable-service-external-name", func() {
	f := framework.NewDefaultFramework("disabled-service-external-name")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(2))

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--disable-svc-external-name=true")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
	})

	ginkgo.It("should ignore services of external-name type", func() {

		nonexternalhost := "echo-svc.com"

		externalhost := "echo-external-svc.com"
		svcexternal := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "external",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "httpbin.org",
				Type:         corev1.ServiceTypeExternalName,
			},
		}
		f.EnsureService(svcexternal)

		ingexternal := framework.NewSingleIngress(externalhost, "/", externalhost, f.Namespace, "external", 80, nil)
		f.EnsureIngress(ingexternal)

		ing := framework.NewSingleIngress(nonexternalhost, "/", nonexternalhost, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(nonexternalhost, func(cfg string) bool {
			return strings.Contains(cfg, "server_name echo-svc.com")
		})

		f.WaitForNginxServer(externalhost, func(cfg string) bool {
			return strings.Contains(cfg, "server_name echo-external-svc.com")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", nonexternalhost).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", externalhost).
			Expect().
			StatusRange(httpexpect.Status5xx)

	})
})
