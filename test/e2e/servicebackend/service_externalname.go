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

package servicebackend

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"

	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Service Type ExternalName", func() {
	f := framework.NewDefaultFramework("type-externalname")

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should return 200 for service type=ExternalName without a port defined", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "httpbin",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "httpbin.org",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "httpbin", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)+"/get").
			Set("Host", host).
			End()
		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(200))
	})

	It("should return 200 for service type=ExternalName with a port defined", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "httpbin",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "httpbin.org",
				Type:         corev1.ServiceTypeExternalName,
				Ports: []corev1.ServicePort{
					{
						Name:       host,
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "httpbin", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)+"/get").
			Set("Host", host).
			End()
		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(200))
	})

	It("should return status 503 for service type=ExternalName with an invalid host", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "httpbin",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "invalid.hostname",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "httpbin", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)+"/get").
			Set("Host", host).
			End()
		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(503))
	})
})
