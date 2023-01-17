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
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/controller/ingressclass"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Flag] ingress-class", func() {
	f := framework.NewDefaultFramework("ingress-class")

	var doOnce sync.Once

	otherIngressClassName := "test-new-ingress-class"
	otherController := "k8s.io/other-class"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(1))

		doOnce.Do(func() {
			_, err := f.KubeClientSet.NetworkingV1().IngressClasses().
				Create(context.TODO(), &networkingv1.IngressClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: otherIngressClassName,
					},
					Spec: networkingv1.IngressClassSpec{
						Controller: otherController,
					},
				}, metav1.CreateOptions{})

			if !apierrors.IsAlreadyExists(err) {
				assert.Nil(ginkgo.GinkgoT(), err, "creating IngressClass")
			}
		})
	})

	ginkgo.Context("With default ingress class config", func() {
		ginkgo.It("should ignore Ingress with a different class annotation", func() {
			invalidHost := "foo"
			annotations := map[string]string{
				ingressclass.IngressKey: "testclass",
			}
			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, annotations)
			// We should drop the ingressClassName here as we just want to rely on the annotation in this test
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			validHost := "bar"
			annotationClass := map[string]string{
				ingressclass.IngressKey: ingressclass.DefaultAnnotationValue,
			}
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, annotationClass)
			ing.Spec.IngressClassName = nil
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

		ginkgo.It("should ignore Ingress with different controller class", func() {
			invalidHost := "foo-1"
			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = &otherIngressClassName
			f.EnsureIngress(ing)

			validHost := "bar-1"
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo-1") &&
					strings.Contains(cfg, "server_name bar-1")
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

		ginkgo.It("should accept both Ingresses with default IngressClassName and IngressClass annotation", func() {
			validHostAnnotation := "foo-ok"
			annotationClass := map[string]string{
				ingressclass.IngressKey: ingressclass.DefaultAnnotationValue,
			}
			ing := framework.NewSingleIngress(validHostAnnotation, "/", validHostAnnotation, f.Namespace, framework.EchoService, 80, annotationClass)
			// We need to drop the Class here as we just want the annotation
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			validHostClass := "bar-ok"
			ing = framework.NewSingleIngress(validHostClass, "/", validHostClass, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo-ok") &&
					strings.Contains(cfg, "server_name bar-ok")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHostAnnotation).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHostClass).
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It("should ignore Ingress without IngressClass configuration", func() {
			invalidHost := "foo-invalid"
			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			validHostClass := "bar-valid"
			ing = framework.NewSingleIngress(validHostClass, "/", validHostClass, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo-invalid") &&
					strings.Contains(cfg, "server_name bar-valid")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", invalidHost).
				Expect().
				Status(http.StatusNotFound)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHostClass).
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It("should delete Ingress when class is removed", func() {
			hostAnnotation := "foo-annotation"

			annotations := map[string]string{
				ingressclass.IngressKey: ingressclass.DefaultAnnotationValue,
			}
			ing := framework.NewSingleIngress(hostAnnotation, "/", hostAnnotation, f.Namespace, framework.EchoService, 80, annotations)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			hostClass := "foo-class"
			ing = framework.NewSingleIngress(hostClass, "/", hostClass, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo-annotation") &&
					strings.Contains(cfg, "server_name foo-class")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostAnnotation).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostClass).
				Expect().
				Status(http.StatusOK)

			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), hostAnnotation, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			delete(ing.Annotations, ingressclass.IngressKey)
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(ing.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ingWithClass, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), hostClass, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ingWithClass.Spec.IngressClassName = nil
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(ingWithClass.Namespace).Update(context.TODO(), ingWithClass, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.Sleep()

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo-annotation") &&
					!strings.Contains(cfg, "server_name foo-class")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostAnnotation).
				Expect().
				Status(http.StatusNotFound)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostClass).
				Expect().
				Status(http.StatusNotFound)
		})

		ginkgo.It("should serve Ingress when class is added", func() {
			hostNoAnnotation := "foo-no-annotation"

			ing := framework.NewSingleIngress(hostNoAnnotation, "/", hostNoAnnotation, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			hostNoClass := "foo-no-class"
			ing = framework.NewSingleIngress(hostNoClass, "/", hostNoClass, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name foo-no-nnotation") &&
					!strings.Contains(cfg, "server_name foo-no-class")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostNoAnnotation).
				Expect().
				Status(http.StatusNotFound)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostNoClass).
				Expect().
				Status(http.StatusNotFound)

			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), hostNoAnnotation, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			annotation := map[string]string{
				ingressclass.IngressKey: ingressclass.DefaultAnnotationValue,
			}
			ing.Annotations = annotation
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(ing.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ingWithClass, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), hostNoClass, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ingWithClass.Spec.IngressClassName = framework.GetIngressClassName(f.Namespace)
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(ingWithClass.Namespace).Update(context.TODO(), ingWithClass, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.Sleep()

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo-no-annotation") &&
					strings.Contains(cfg, "server_name foo-no-class")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostNoAnnotation).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostNoClass).
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It("should serve Ingress when class is updated between annotation and ingressClassName", func() {
			hostAnnotation2class := "foo-annotation2class"
			annotationClass := map[string]string{
				ingressclass.IngressKey: ingressclass.DefaultAnnotationValue,
			}
			ing := framework.NewSingleIngress(hostAnnotation2class, "/", hostAnnotation2class, f.Namespace, framework.EchoService, 80, annotationClass)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			hostClass2Annotation := "foo-class2annotation"
			ing = framework.NewSingleIngress(hostClass2Annotation, "/", hostClass2Annotation, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo-annotation2class") &&
					strings.Contains(cfg, "server_name foo-class2annotation")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostAnnotation2class).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostClass2Annotation).
				Expect().
				Status(http.StatusOK)

			ingAnnotation2Class, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), hostAnnotation2class, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			delete(ingAnnotation2Class.Annotations, ingressclass.IngressKey)
			ingAnnotation2Class.Spec.IngressClassName = framework.GetIngressClassName(ingAnnotation2Class.Namespace)
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(ingAnnotation2Class.Namespace).Update(context.TODO(), ingAnnotation2Class, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ingClass2Annotation, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), hostClass2Annotation, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ingClass2Annotation.Spec.IngressClassName = nil
			ingClass2Annotation.Annotations = annotationClass
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(ingClass2Annotation.Namespace).Update(context.TODO(), ingClass2Annotation, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			framework.Sleep()

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo-annotation2class") &&
					strings.Contains(cfg, "server_name foo-class2annotation")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostAnnotation2class).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", hostClass2Annotation).
				Expect().
				Status(http.StatusOK)
		})

	})

	ginkgo.Context("With specific ingress-class flags", func() {
		ginkgo.BeforeEach(func() {
			err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
				args := []string{}
				for _, v := range deployment.Spec.Template.Spec.Containers[0].Args {
					if strings.Contains(v, "--ingress-class") && strings.Contains(v, "--controller-class") {
						continue
					}

					args = append(args, v)
				}

				args = append(args, "--ingress-class=testclass")
				args = append(args, "--controller-class=k8s.io/other-class")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

				return err
			})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
		})

		ginkgo.It("should ignore Ingress with no class and accept the correctly configured Ingresses", func() {
			invalidHost := "bar"

			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			validHost := "foo"
			annotations := map[string]string{
				ingressclass.IngressKey: "testclass",
			}
			ing = framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, annotations)
			// Delete the IngressClass as we want just the annotation here
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			validHostClass := "foobar123"
			ing = framework.NewSingleIngress(validHostClass, "/", validHostClass, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = &otherIngressClassName
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name bar") &&
					strings.Contains(cfg, "server_name foo") &&
					strings.Contains(cfg, "server_name foobar123")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHost).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHostClass).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", invalidHost).
				Expect().
				Status(http.StatusNotFound)
		})

	})

	ginkgo.Context("With watch-ingress-without-class flag", func() {
		ginkgo.BeforeEach(func() {
			err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
				args := []string{}
				for _, v := range deployment.Spec.Template.Spec.Containers[0].Args {
					if strings.Contains(v, "--watch-ingress-without-class") && strings.Contains(v, "--controller-class") {
						continue
					}

					args = append(args, v)
				}

				args = append(args, "--watch-ingress-without-class")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

				return err
			})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
		})

		ginkgo.It("should watch Ingress with no class and ignore ingress with a different class", func() {
			validHost := "bar"

			ing := framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			invalidHost := "foo"
			annotations := map[string]string{
				ingressclass.IngressKey: "testclass123",
			}
			ing = framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, annotations)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name bar") &&
					!strings.Contains(cfg, "server_name foo")
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

	})

	ginkgo.Context("With ingress-class-by-name flag", func() {
		ginkgo.BeforeEach(func() {
			err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
				args := []string{}
				for _, v := range deployment.Spec.Template.Spec.Containers[0].Args {
					if strings.Contains(v, "--ingress-class-by-name") &&
						strings.Contains(v, "--ingress-class=test-new-ingress-class") {
						continue
					}

					args = append(args, v)
				}
				args = append(args, "--ingress-class=test-new-ingress-class")
				args = append(args, "--ingress-class-by-name")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

				return err
			})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
		})

		ginkgo.It("should watch Ingress that uses the class name even if spec is different", func() {
			validHostClassName := "validhostclassname"

			ing := framework.NewSingleIngress(validHostClassName, "/", validHostClassName, f.Namespace, framework.EchoService, 80, nil)
			ing.Spec.IngressClassName = &otherIngressClassName
			f.EnsureIngress(ing)

			validHostClass := "validhostclassspec"
			ing = framework.NewSingleIngress(validHostClass, "/", validHostClass, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			invalidHost := "invalidannotation"
			annotations := map[string]string{
				ingressclass.IngressKey: "testclass123",
			}
			ing = framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, annotations)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name validhostclassname") &&
					strings.Contains(cfg, "server_name validhostclassspec") &&
					!strings.Contains(cfg, "server_name invalidannotation")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHostClass).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHostClassName).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", invalidHost).
				Expect().
				Status(http.StatusNotFound)
		})

	})

	ginkgo.Context("Without IngressClass Cluster scoped Permission", func() {

		ginkgo.BeforeEach(func() {
			icname := fmt.Sprintf("ic-%s", f.Namespace)

			newRole := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: icname,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     icname,
				},
				Subjects: []rbacv1.Subject{
					{
						APIGroup:  "",
						Kind:      "ServiceAccount",
						Namespace: f.Namespace,
						Name:      "blablabla",
					},
				},
			}
			_, err := f.KubeClientSet.RbacV1().ClusterRoleBindings().Update(context.TODO(), newRole, metav1.UpdateOptions{})

			assert.Nil(ginkgo.GinkgoT(), err, "Updating IngressClass ClusterRoleBinding")

			// Force the correct annotation value just for the re-deployment
			err = f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
				args := []string{}
				for _, v := range deployment.Spec.Template.Spec.Containers[0].Args {
					if strings.Contains(v, "--ingress-class=testclass") {
						continue
					}

					args = append(args, v)
				}
				args = append(args, "--ingress-class=testclass")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

				return err
			})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
		})

		ginkgo.It("should watch Ingress with correct annotation", func() {

			validHost := "foo"
			annotations := map[string]string{
				ingressclass.IngressKey: "testclass",
			}
			ing := framework.NewSingleIngress(validHost, "/", validHost, f.Namespace, framework.EchoService, 80, annotations)
			ing.Spec.IngressClassName = nil
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name foo")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", validHost).
				Expect().
				Status(http.StatusOK)
		})

		ginkgo.It("should ignore Ingress with only IngressClassName", func() {

			invalidHost := "noclassforyou"

			ing := framework.NewSingleIngress(invalidHost, "/", invalidHost, f.Namespace, framework.EchoService, 80, nil)
			f.EnsureIngress(ing)

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name noclassforyou")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", invalidHost).
				Expect().
				Status(http.StatusNotFound)
		})

	})
})
