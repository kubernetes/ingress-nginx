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

package annotations

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Proxy", func() {
	f := framework.NewDefaultFramework("proxy")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(2)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should set proxy_redirect to off", func() {
		host := "proxy.foo.com"

		ing, err := f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.IngressController.Namespace,
				Annotations: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-redirect-from": "off",
					"nginx.ingress.kubernetes.io/proxy-redirect-to":   "goodbye.com",
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
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_redirect off;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should set proxy_redirect to default", func() {
		host := "proxy.foo.com"

		ing, err := f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.IngressController.Namespace,
				Annotations: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-redirect-from": "default",
					"nginx.ingress.kubernetes.io/proxy-redirect-to":   "goodbye.com",
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
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_redirect default;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should set proxy_redirect to hello.com goodbye.com", func() {
		host := "proxy.foo.com"

		ing, err := f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.IngressController.Namespace,
				Annotations: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-redirect-from": "hello.com",
					"nginx.ingress.kubernetes.io/proxy-redirect-to":   "goodbye.com",
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
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("proxy_redirect hello.com goodbye.com;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})
})
