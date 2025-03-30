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
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("single ingress - multiple hosts", func() {
	f := framework.NewDefaultFramework("simh")
	pathprefix := networking.PathTypePrefix
	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentName("first-service"))
		f.NewEchoDeployment(framework.WithDeploymentName("second-service"))
	})

	ginkgo.It("should set the correct $service_name NGINX variable", func() {
		customHeader := "Service-Name"
		customHeaderValue := "$service_name"

		h := make(map[string]string)
		h[customHeader] = customHeaderValue

		cfgMap := "custom-headers"

		f.CreateConfigMap(cfgMap, h)

		f.UpdateNginxConfigMapData("add-headers", fmt.Sprintf("%v/%v", f.Namespace, cfgMap))

		ing := framework.NewSingleIngress("simh", "/", "first.host", f.Namespace, "first-service", 80, nil)

		ing.Spec.Rules = append(ing.Spec.Rules, networking.IngressRule{
			Host: "second.host",
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: []networking.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathprefix,
							Backend: networking.IngressBackend{
								Service: &networking.IngressServiceBackend{
									Name: "second-service",
									Port: networking.ServiceBackendPort{
										Number: int32(80),
									},
								},
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

		f.HTTPTestClient().
			GET("/exact").
			WithHeader("Host", "first.host").
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Service-Name", []string{"first-service"})

		f.HTTPTestClient().
			GET("/exact").
			WithHeader("Host", "second.host").
			Expect().
			Status(http.StatusOK).
			Headers().ValueEqual("Service-Name", []string{"second-service"})

	})
})
