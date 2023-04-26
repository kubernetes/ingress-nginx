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

	"k8s.io/ingress-nginx/test/e2e/framework/httpexpect"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

func buildHTTPBinExternalNameService(f *framework.Framework, portName string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "neverssl",
			Namespace: f.Namespace,
		},
		Spec: corev1.ServiceSpec{
			ExternalName: "neverssl.com",
			Type:         corev1.ServiceTypeExternalName,
			Ports: []corev1.ServicePort{
				{
					Name:       portName,
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   "TCP",
				},
			},
		},
	}
}

var _ = framework.IngressNginxDescribe("[Service] type=ExternalName", func() {
	f := framework.NewDefaultFramework("type-externalname")

	ginkgo.It("works with ExternalName set to incomplete fqdn", func() {
		f.NewEchoDeployment()

		host := "echo"

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neverssl",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: framework.EchoService,
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, nil)
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

	ginkgo.It("should return 200 for service ExternalName without a port defined", func() {
		host := "echo"

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neverssl",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "neverssl.com",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "neverssl.com",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return 200 for service ExternalName with a port defined", func() {
		host := "echo"

		svc := buildHTTPBinExternalNameService(f, host)
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "neverssl.com",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return status 502 for service ExternalName with an invalid host", func() {
		host := "echo"

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neverssl",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "invalid.hostname",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			StatusRange(httpexpect.Status5xx)
	})

	ginkgo.It("should return 200 for service ExternalName using a port name", func() {
		host := "echo"

		svc := buildHTTPBinExternalNameService(f, host)
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "neverssl.com",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, annotations)
		namedBackend := networking.IngressBackend{
			Service: &networking.IngressServiceBackend{
				Name: "neverssl",
				Port: networking.ServiceBackendPort{
					Name: host,
				},
			},
		}
		ing.Spec.Rules[0].HTTP.Paths[0].Backend = namedBackend
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return 200 for service ExternalName using FQDN with trailing dot", func() {
		host := "echo"

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neverssl",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "github.com.",
				Type:         corev1.ServiceTypeExternalName,
			},
		}

		f.EnsureService(svc)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, nil)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusMovedPermanently)
	})

	ginkgo.It("should update the ExternalName after a service update", func() {
		host := "echo"

		svc := buildHTTPBinExternalNameService(f, host)
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-vhost": "neverssl.com",
		}
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, annotations)
		namedBackend := networking.IngressBackend{
			Service: &networking.IngressServiceBackend{
				Name: "neverssl",
				Port: networking.ServiceBackendPort{
					Name: host,
				},
			},
		}
		ing.Spec.Rules[0].HTTP.Paths[0].Backend = namedBackend
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, `neverssl.com`)

		svc, err := f.KubeClientSet.CoreV1().Services(f.Namespace).Get(context.TODO(), "neverssl", metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining service")

		svc.Spec.ExternalName = "google.com"

		_, err = f.KubeClientSet.CoreV1().Services(f.Namespace).Update(context.Background(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		framework.Sleep(10)

		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, `neverssl.com`)

		ginkgo.By("checking the service is updated to use google.com")
		curlCmd := fmt.Sprintf("curl --fail --silent http://localhost:%v/configuration/backends", nginx.StatusPort)
		output, err := f.ExecIngressPod(curlCmd)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Contains(ginkgo.GinkgoT(), output, `{"address":"google.com"`)
	})

	ginkgo.It("should sync ingress on ExternalName service addition/deletion", func() {
		host := "echo"

		// Create the Ingress first
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "neverssl", 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})

		// Nginx should return 503 without the underlying service being available
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusServiceUnavailable)

		// Now create the service
		svc := buildHTTPBinExternalNameService(f, host)
		f.EnsureService(svc)

		framework.Sleep()

		// 503 should change to 200 OK
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		// And back to 503 after deleting the service
		err := f.KubeClientSet.CoreV1().Services(f.Namespace).Delete(context.TODO(), "neverssl", metav1.DeleteOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error deleting service")

		framework.Sleep()

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusServiceUnavailable)
	})
})
