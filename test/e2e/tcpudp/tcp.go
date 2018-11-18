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

package settings

import (
	"fmt"
	"strings"

	"github.com/parnurzeal/gorequest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("TCP Feature", func() {
	f := framework.NewDefaultFramework("tcp")

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should expose a TCP service", func() {
		f.NewEchoDeploymentWithReplicas(1)

		config, err := f.KubeClientSet.
			CoreV1().
			ConfigMaps(f.IngressController.Namespace).
			Get("tcp-services", metav1.GetOptions{})
		Expect(err).To(BeNil(), "unexpected error obtaining tcp-services configmap")
		Expect(config).NotTo(BeNil(), "expected a configmap but none returned")

		if config.Data == nil {
			config.Data = map[string]string{}
		}

		config.Data["8080"] = fmt.Sprintf("%v/http-svc:80", f.IngressController.Namespace)
		_, err = f.KubeClientSet.
			CoreV1().
			ConfigMaps(f.IngressController.Namespace).
			Update(config)
		Expect(err).NotTo(HaveOccurred(), "unexpected error updating configmap")

		svc, err := f.KubeClientSet.
			CoreV1().
			Services(f.IngressController.Namespace).
			Get("ingress-nginx", metav1.GetOptions{})
		Expect(err).To(BeNil(), "unexpected error obtaining ingress-nginx service")
		Expect(svc).NotTo(BeNil(), "expected a service but none returned")

		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "http-svc",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
		})
		_, err = f.KubeClientSet.
			CoreV1().
			Services(f.IngressController.Namespace).
			Update(svc)
		Expect(err).NotTo(HaveOccurred(), "unexpected error updating service")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf(`ngx.var.proxy_upstream_name="tcp-%v-http-svc-80"`, f.IngressController.Namespace))
			})

		ip := f.GetNginxIP()
		port, err := f.GetNginxPort("http-svc")
		Expect(err).NotTo(HaveOccurred(), "unexpected error obtaning service port")

		resp, _, errs := gorequest.New().
			Get(fmt.Sprintf("http://%v:%v", ip, port)).
			End()
		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(200))
	})
})
