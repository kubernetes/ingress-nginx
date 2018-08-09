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

package multipleingress

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Service backend - 503", func() {
	f := framework.NewDefaultFramework("service-backend")

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should return 503 when backend service does not exist", func() {
		//create a new service
		service := buildService("some-service-name", "some-namespace", 80, 443)
		svc, err := f.EnsureService(service)
		Expect(err).NotTo(HaveOccurred())
		Expect(svc).NotTo(BeNil())

		ingress1spec := buildIngress("ingress-1.example.com", "some-namespace", "/", "some-service-name", 80)
		ingress1, err := f.EnsureIngress(ingress1spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(ingress1).NotTo(BeNil())

		ingress2spec := buildIngress("ingress-2.example.com", "some-namespace", "/", "some-service-name", 443)
		//add secure-backend annotation to 2nd ingress
		ingress2spec.Annotations["nginx.ingress.kubernetes.io/secure-backends"] = "true"

		ingress2, err := f.EnsureIngress(ingress2spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(ingress2).NotTo(BeNil())

		err = f.WaitForNginxServer("ingress-1.example.com",
			func(server string) bool {
				return strings.Contains(server, "return 503;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", "ingress-1.example.com").
			End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(503))
	})
})

func buildService(name, namespace string, port1, port2 int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{
			{
				Name:       fmt.Sprintf("%s", port1),
				Port:       port1,
				TargetPort: intstr.FromInt(int(port1)),
				Protocol:   "TCP",
			},
			{
				Name:       fmt.Sprintf("%s", port2),
				Port:       port2,
				TargetPort: intstr.FromInt(int(port2)),
				Protocol:   "TCP",
			},
		},
			Selector: map[string]string{
				"app": name,
			},
		},
	}
}

func buildIngress(host, namespace, path, backendService string, port int) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: path,
									Backend: v1beta1.IngressBackend{
										ServiceName: backendService,
										ServicePort: intstr.FromInt(port),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
