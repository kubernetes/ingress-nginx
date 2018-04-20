/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Alias", func() {
	f := framework.NewDefaultFramework("alias")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should return status code 200 when no authentication is configured", func() {
		host := "auth"

		bi, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(bi).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth")) &&
					Expect(server).ShouldNot(ContainSubstring("return 503"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", host)))
	})

	It("should return status code 503 when authentication is configured with an invalid secret", func() {
		host := "auth"

		bi, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(bi).NotTo(BeNil())

		bi.Annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		bi.Annotations["nginx.ingress.kubernetes.io/auth-secret"] = "something"
		bi.Annotations["nginx.ingress.kubernetes.io/auth-realm"] = "test auth"

		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth")) &&
					Expect(server).Should(ContainSubstring("return 503"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusServiceUnavailable))
		Expect(body).Should(ContainSubstring("503 Service Temporarily Unavailable"))
	})

	It("should return status code 401 when authentication is configured but Authorization header is not configured", func() {
		host := "auth"

		s, err := f.EnsureSecret(buildSecret("foo", "bar", "test", f.IngressController.Namespace))
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())
		Expect(s.ObjectMeta).NotTo(BeNil())

		bi, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(bi).NotTo(BeNil())

		bi.Annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		bi.Annotations["nginx.ingress.kubernetes.io/auth-secret"] = s.Name
		bi.Annotations["nginx.ingress.kubernetes.io/auth-realm"] = "test auth"

		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth")) &&
					Expect(server).ShouldNot(ContainSubstring("return 503"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusUnauthorized))
		Expect(body).Should(ContainSubstring("401 Authorization Required"))
	})

	It("should return status code 401 when authentication is configured and Authorization header is sent with invalid credentials", func() {
		host := "auth"

		s, err := f.EnsureSecret(buildSecret("foo", "bar", "test", f.IngressController.Namespace))
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())
		Expect(s.ObjectMeta).NotTo(BeNil())

		bi, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(bi).NotTo(BeNil())

		bi.Annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		bi.Annotations["nginx.ingress.kubernetes.io/auth-secret"] = s.Name
		bi.Annotations["nginx.ingress.kubernetes.io/auth-realm"] = "test auth"

		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth")) &&
					Expect(server).ShouldNot(ContainSubstring("return 503"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			SetBasicAuth("user", "pass").
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusUnauthorized))
		Expect(body).Should(ContainSubstring("401 Authorization Required"))
	})

	It("should return status code 200 when authentication is configured and Authorization header is sent", func() {
		host := "auth"

		s, err := f.EnsureSecret(buildSecret("foo", "bar", "test", f.IngressController.Namespace))
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())
		Expect(s.ObjectMeta).NotTo(BeNil())

		bi, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(bi).NotTo(BeNil())

		bi.Annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		bi.Annotations["nginx.ingress.kubernetes.io/auth-secret"] = s.Name
		bi.Annotations["nginx.ingress.kubernetes.io/auth-realm"] = "test auth"

		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth")) &&
					Expect(server).ShouldNot(ContainSubstring("return 503"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			SetBasicAuth("foo", "bar").
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})

	It("should return status code 500 when authentication is configured with invalid content and Authorization header is sent", func() {
		host := "auth"

		s, err := f.EnsureSecret(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: f.IngressController.Namespace,
				},
				Data: map[string][]byte{
					// invalid content
					"auth": []byte("foo:"),
				},
				Type: corev1.SecretTypeOpaque,
			},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())
		Expect(s.ObjectMeta).NotTo(BeNil())

		bi, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(bi).NotTo(BeNil())

		bi.Annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		bi.Annotations["nginx.ingress.kubernetes.io/auth-secret"] = s.Name
		bi.Annotations["nginx.ingress.kubernetes.io/auth-realm"] = "test auth"

		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("server_name auth")) &&
					Expect(server).ShouldNot(ContainSubstring("return 503"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			SetBasicAuth("foo", "bar").
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusInternalServerError))
	})
})

// TODO: test Digest Auth
//   401
//   Realm name
//   Auth ok
//   Auth error

func buildSecret(username, password, name, namespace string) *corev1.Secret {
	out, err := exec.Command("openssl", "passwd", "-crypt", password).CombinedOutput()
	encpass := fmt.Sprintf("%v:%s\n", username, out)
	Expect(err).NotTo(HaveOccurred())

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       name,
			Namespace:                  namespace,
			DeletionGracePeriodSeconds: framework.NewInt64(1),
		},
		Data: map[string][]byte{
			"auth": []byte(encpass),
		},
		Type: corev1.SecretTypeOpaque,
	}
}
