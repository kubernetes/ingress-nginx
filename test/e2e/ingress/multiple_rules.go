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

package ingress

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("single ingress - multiple hosts", func() {
	f := framework.NewDefaultFramework("simh")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeploymentWithNameAndReplicas("first-service", 1)
		f.NewEchoDeploymentWithNameAndReplicas("second-service", 1)
	})

	ginkgo.It("should set the correct $service_name NGINX variable", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_input_headers "service-name: $service_name";`,
		}

		ing := framework.NewSingleIngress("simh", "/", "first.host", f.Namespace, "first-service", 80, annotations)

		ing.Spec.Rules = append(ing.Spec.Rules, networkingv1beta1.IngressRule{
			Host: "second.host",
			IngressRuleValue: networkingv1beta1.IngressRuleValue{
				HTTP: &networkingv1beta1.HTTPIngressRuleValue{
					Paths: []networkingv1beta1.HTTPIngressPath{
						{
							Path: "/",
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: "second-service",
								ServicePort: intstr.FromInt(80),
							},
						},
					},
				},
			},
		})

		f.EnsureIngress(ing)

		f.WaitForNginxServer("first.host",
			func(server string) bool {
				return strings.Contains(server, "first.host")
			})

		f.WaitForNginxServer("second.host",
			func(server string) bool {
				return strings.Contains(server, "second.host")
			})

		body := f.HTTPTestClient().
			GET("/exact").
			WithHeader("Host", "first.host").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "service-name=first-service")
		assert.NotContains(ginkgo.GinkgoT(), body, "service-name=second-service")

		body = f.HTTPTestClient().
			GET("/exact").
			WithHeader("Host", "second.host").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.NotContains(ginkgo.GinkgoT(), body, "service-name=first-service")
		assert.Contains(ginkgo.GinkgoT(), body, "service-name=second-service")
	})
})
