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

package settings

import (
	"context"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Flag] ingress-class", func() {
	f := framework.NewDefaultFramework("ingress-class")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	ginkgo.Context("Without a specific ingress-class", func() {

		ginkgo.It("should ignore Ingress with class", func() {
			invalidHost := "foo"
			annotations := map[string]string{
				class.IngressKey: "testclass",
			}
			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, annotations)
			f.EnsureIngress(ing)

			validHost := "bar"
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo") &&
					strings.Contains(cfg, "server_name bar")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", invalidHost).
				Expect().
				Status(http.StatusNotFound)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHost).
				Expect().
				Status(http.StatusOK)
		})
	})

	ginkgo.Context("With a specific ingress-class", func() {
		ginkgo.BeforeEach(func() {
			err := framework.UpdateDeployment(f.KubeClientSet, f.Namespace, "nginx-ingress-controller", 1,
				func(deployment *appsv1.Deployment) error {
					args := deployment.Spec.Template.Spec.Containers[0].Args
					args = append(args, "--ingress-class=testclass")
					deployment.Spec.Template.Spec.Containers[0].Args = args
					_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

					return err
				})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
		})

		ginkgo.It("should ignore Ingress with no class", func() {
			invalidHost := "bar"

			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			validHost := "foo"
			annotations := map[string]string{
				class.IngressKey: "testclass",
			}
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(validHost, func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo")
			})

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name bar")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHost).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", invalidHost).
				Expect().
				Status(http.StatusNotFound)
		})

		ginkgo.It("should delete Ingress when class is removed", func() {
			host := "foo"
			annotations := map[string]string{
				class.IngressKey: "testclass",
			}
			ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host, func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)

			ing, err := f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Get(context.TODO(), host, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			delete(ing.Annotations, class.IngressKey)
			_, err = f.KubeClientSet.NetworkingV1beta1().Ingresses(ing.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusNotFound)
		})
	})
})
