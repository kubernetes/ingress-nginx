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
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gavv/httpexpect/v2"
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Service] Type ExternalName", func() {
	f := framework.NewDefaultFramework("type-externalname")

	ginkgo.It("works with external name set to incomplete fdqn", func() {
		f.NewEchoDeployment()

		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      framework.HTTPBinService,
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: framework.EchoService,
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return 200 for service type=ExternalName without a port defined", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      framework.HTTPBinService,
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "httpbin.org",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "httpbin.org",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return 200 for service type=ExternalName with a port defined", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      framework.HTTPBinService,
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

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "httpbin.org",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return status 502 for service type=ExternalName with an invalid host", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      framework.HTTPBinService,
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "invalid.hostname",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			StatusRange(httpexpect.Status5xx)
	})

	ginkgo.It("should return 200 for service type=ExternalName using a port name", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      framework.HTTPBinService,
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

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "httpbin.org",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, annotations)
		ing.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort = intstr.FromString(host)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should update the external name after a service update", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      framework.HTTPBinService,
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

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "httpbin.org",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, annotations)
		ing.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort = intstr.FromString(host)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		body := f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, `"X-Forwarded-Host": "echo"`)

		svc, err := f.KubeClientSet.CoreV1().Services(f.Namespace).Get(context.TODO(), framework.HTTPBinService, metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining httpbin service")

		svc.Spec.ExternalName = "eu.httpbin.org"

		_, err = f.KubeClientSet.CoreV1().Services(f.Namespace).Update(context.Background(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating httpbin service")

		framework.Sleep()

		body = f.HTTPTestClient().
			GET("/get").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, `"X-Forwarded-Host": "echo"`)

		ginkgo.By("checking the service is updated to use eu.httpbin.org")
		curlCmd := fmt.Sprintf("curl --fail --silent http://localhost:%v/configuration/backends", nginx.StatusPort)
		output, err := f.ExecIngressPod(curlCmd)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Contains(ginkgo.GinkgoT(), output, `{"address":"eu.httpbin.org"`)
	})
})
