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

package servicebackend

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Service] Nil Service Backend", func() {
	f := framework.NewDefaultFramework("service-nil-backend")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should return 404 when backend service is nil", func() {
		ginkgo.By("setting an ingress with a nil backend")
		validHost := "valid.svc.com"
		invalidHost := "nilbackend.svc.com"

		ing := framework.NewSingleIngress(validHost, "/", validHost, f.Namespace,
			framework.EchoService, 80, nil)

		bi := buildIngressWithNonServiceBackend(invalidHost, f.Namespace, "/")

		f.EnsureIngress(bi)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "server_name nilbackend.svc.com") &&
				strings.Contains(cfg, "server_name valid.svc.com")
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

func buildIngressWithNonServiceBackend(host, namespace, path string) *networking.Ingress {
	apiGroup := "otherobj.testingress.com"
	obj := corev1.TypedLocalObjectReference{
		Kind:     "Anything",
		Name:     "mytest",
		APIGroup: &apiGroup,
	}

	return &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: namespace,
		},
		Spec: networking.IngressSpec{
			IngressClassName: framework.GetIngressClassName(namespace),
			Rules: []networking.IngressRule{
				{
					Host: host,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path:     path,
									PathType: &pathtype,
									Backend: networking.IngressBackend{
										Resource: &obj,
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
