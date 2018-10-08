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

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - HealthCheck", func() {
	f := framework.NewDefaultFramework("healthcheck")

	BeforeEach(func() {
		err := f.DisableDynamicConfiguration()
		Expect(err).NotTo(HaveOccurred())

		err = f.NewEchoDeploymentWithReplicas(2)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should set upstream-max-fails to 11", func() {
		host := "healthcheck.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-max-fails": "11",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxConfiguration(
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("max_fails=11 fail_timeout=0;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not set upstream-max-fails to 11s", func() {
		host := "healthcheck.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-max-fails": "11s",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxConfiguration(
			func(server string) bool {
				return Expect(server).ShouldNot(ContainSubstring("max_fails=11s fail_timeout=0;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should set upstream-fail-timeout to 15", func() {
		host := "healthcheck.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-fail-timeout": "15",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxConfiguration(
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("max_fails=0 fail_timeout=15;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not set upstream-fail-timeout to 15s", func() {
		host := "healthcheck.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-fail-timeout": "15s",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxConfiguration(
			func(server string) bool {
				return Expect(server).ShouldNot(ContainSubstring("max_fails=0 fail_timeout=15s;"))
			})
		Expect(err).NotTo(HaveOccurred())
	})
})
