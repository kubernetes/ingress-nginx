/*
Copyright 2018 The Kubernetes Authors.

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

package lua

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Dynamic Configuration", func() {
	f := framework.NewDefaultFramework("dynamic-configuration")

	BeforeEach(func() {
		err := enableDynamicConfiguration(f.KubeClientSet)
		Expect(err).NotTo(HaveOccurred())

		err = f.NewEchoDeploymentWithReplicas(1)
		Expect(err).NotTo(HaveOccurred())

		host := "foo.com"
		ing, err := ensureIngress(f, host)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		time.Sleep(5 * time.Second)

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.NginxHTTPURL).
			Set("Host", host).
			End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(ContainSubstring("could not dynamically reconfigure"))
		Expect(log).To(ContainSubstring("first sync of Nginx configuration"))
	})

	AfterEach(func() {
		err := disableDynamicConfiguration(f.KubeClientSet)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when only backends change", func() {
		It("should handle endpoints only changes", func() {
			resp, _, errs := gorequest.New().
				Get(fmt.Sprintf("%s?id=endpoints_only_changes", f.NginxHTTPURL)).
				Set("Host", "foo.com").
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			replicas := 2
			err := framework.UpdateDeployment(f.KubeClientSet, f.Namespace.Name, "http-svc", replicas,
				func(deployment *appsv1beta1.Deployment) error {
					deployment.Spec.Replicas = framework.NewInt32(int32(replicas))
					_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.Namespace.Name).Update(deployment)
					return err
				})
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)
			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "id=endpoints_only_changes")
			restOfLogs := log[index:]

			By("POSTing new backends to Lua endpoint")
			Expect(restOfLogs).To(ContainSubstring("dynamic reconfiguration succeeded"))
			Expect(restOfLogs).ToNot(ContainSubstring("could not dynamically reconfigure"))

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring("backend reload required"))
			Expect(restOfLogs).ToNot(ContainSubstring("ingress backend successfully reloaded"))
			Expect(restOfLogs).To(ContainSubstring("skipping reload"))
			Expect(restOfLogs).ToNot(ContainSubstring("first sync of Nginx configuration"))
		})

		It("should handle annotation changes", func() {
			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace.Name).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/load-balance"] = "round_robin"
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace.Name).Update(ingress)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(5 * time.Second)
			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, fmt.Sprintf("reason: 'UPDATE' Ingress %s/foo.com", f.Namespace.Name))
			restOfLogs := log[index:]

			By("POSTing new backends to Lua endpoint")
			Expect(restOfLogs).To(ContainSubstring("dynamic reconfiguration succeeded"))
			Expect(restOfLogs).ToNot(ContainSubstring("could not dynamically reconfigure"))

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring("backend reload required"))
			Expect(restOfLogs).ToNot(ContainSubstring("ingress backend successfully reloaded"))
			Expect(restOfLogs).To(ContainSubstring("skipping reload"))
			Expect(restOfLogs).ToNot(ContainSubstring("first sync of Nginx configuration"))
		})
	})

	It("should handle a non backend update", func() {
		ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace.Name).Get("foo.com", metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{"foo.com"},
				SecretName: "foo.com",
			},
		}

		_, _, _, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ingress.Spec.TLS[0].Hosts,
			ingress.Spec.TLS[0].SecretName,
			ingress.Namespace)
		Expect(err).ToNot(HaveOccurred())

		_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace.Name).Update(ingress)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)
		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		By("reloading Nginx")
		Expect(log).To(ContainSubstring("ingress backend successfully reloaded"))

		By("POSTing new backends to Lua endpoint")
		Expect(log).To(ContainSubstring("dynamic reconfiguration succeeded"))

		By("still be proxying requests through Lua balancer")
		err = f.WaitForNginxServer("foo.com",
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
		Expect(err).NotTo(HaveOccurred())

		By("generating the respective ssl listen directive")
		err = f.WaitForNginxServer("foo.com",
			func(server string) bool {
				return strings.Contains(server, "server_name foo.com") &&
					strings.Contains(server, "listen 443")
			})
		Expect(err).ToNot(HaveOccurred())
	})
})

func enableDynamicConfiguration(kubeClientSet kubernetes.Interface) error {
	return framework.UpdateDeployment(kubeClientSet, "ingress-nginx", "nginx-ingress-controller", 1,
		func(deployment *appsv1beta1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--enable-dynamic-configuration")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := kubeClientSet.AppsV1beta1().Deployments("ingress-nginx").Update(deployment)
			return err
		})
}

func disableDynamicConfiguration(kubeClientSet kubernetes.Interface) error {
	return framework.UpdateDeployment(kubeClientSet, "ingress-nginx", "nginx-ingress-controller", 1,
		func(deployment *appsv1beta1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			var newArgs []string
			for _, arg := range args {
				if arg != "--enable-dynamic-configuration" {
					newArgs = append(newArgs, arg)
				}
			}
			deployment.Spec.Template.Spec.Containers[0].Args = newArgs
			_, err := kubeClientSet.AppsV1beta1().Deployments("ingress-nginx").Update(deployment)
			return err
		})
}

func ensureIngress(f *framework.Framework, host string) (*extensions.Ingress, error) {
	return f.EnsureIngress(&v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: f.Namespace.Name,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/load-balance": "ewma",
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: "http-svc",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
