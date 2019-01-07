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
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	appsv1beta1 "k8s.io/api/apps/v1beta1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Ingress class", func() {
	f := framework.NewDefaultFramework("ingress-class")

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	AfterEach(func() {
	})

	Context("Without a specific ingress-class", func() {

		It("should ignore Ingress with class", func() {
			invalidHost := "foo"
			annotations := map[string]string{
				"kubernetes.io/ingress.class": "testclass",
			}
			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			validHost := "bar"
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.IngressController.Namespace, "http-svc", 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo") &&
					strings.Contains(cfg, "server_name bar")
			})

			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", invalidHost).
				End()
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))

			resp, _, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", validHost).
				End()
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
	})

	Context("With a specific ingress-class", func() {
		BeforeEach(func() {
			framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "nginx-ingress-controller", 1,
				func(deployment *appsv1beta1.Deployment) error {
					args := deployment.Spec.Template.Spec.Containers[0].Args
					args = append(args, "--ingress-class=testclass")
					deployment.Spec.Template.Spec.Containers[0].Args = args
					_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.IngressController.Namespace).Update(deployment)

					return err
				})
		})

		It("should ignore Ingress with no class", func() {
			invalidHost := "bar"

			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.IngressController.Namespace, "http-svc", 80, nil)
			f.EnsureIngress(ing)

			validHost := "foo"
			annotations := map[string]string{
				"kubernetes.io/ingress.class": "testclass",
			}
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.IngressController.Namespace, "http-svc", 80, &annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(validHost, func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo")
			})

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name bar")
			})

			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", validHost).
				End()
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			resp, _, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", invalidHost).
				End()
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
		})

		It("should delete Ingress when class is removed", func() {
			host := "foo"
			annotations := map[string]string{
				"kubernetes.io/ingress.class": "testclass",
			}
			ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
			ing = f.EnsureIngress(ing)

			f.WaitForNginxServer(host, func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo")
			})

			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			delete(ing.Annotations, "kubernetes.io/ingress.class")
			ing = f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo")
			})

			resp, _, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
		})
	})

})
