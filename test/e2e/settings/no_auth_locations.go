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

package setting

import (
	"fmt"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("No Auth locations", func() {
	f := framework.NewDefaultFramework("no-auth-locations")

	setting := "no-auth-locations"
	username := "foo"
	password := "bar"
	secretName := "test-secret"
	host := "no-auth-locations"
	noAuthPath := "/noauth"

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())

		s, err := f.EnsureSecret(buildSecret(username, password, secretName, f.IngressController.Namespace))
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())
		Expect(s.ObjectMeta).NotTo(BeNil())

		err = f.UpdateNginxConfigMapData(setting, noAuthPath)
		Expect(err).NotTo(HaveOccurred())

		bi := buildBasicAuthIngressWithSecondPath(host, f.IngressController.Namespace, s.Name, noAuthPath)
		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())
	})

	AfterEach(func() {
	})

	It("should return status code 401 when accessing '/' unauthentication", func() {
		err := f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("test auth"))
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

	It("should return status code 200 when accessing '/'  authentication", func() {
		err := f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("test auth"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			SetBasicAuth(username, password).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})

	It("should return status code 200 when accessing '/noauth' unauthenticated", func() {
		err := f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("test auth"))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(fmt.Sprintf("%s/noauth", f.IngressController.HTTPURL)).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})
})

func buildBasicAuthIngressWithSecondPath(host, namespace, secretName, pathName string) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: namespace,
			Annotations: map[string]string{"nginx.ingress.kubernetes.io/auth-type": "basic",
				"nginx.ingress.kubernetes.io/auth-secret": secretName,
				"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: "http-svc",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: pathName,
									Backend: v1beta1.IngressBackend{
										ServiceName: "http-svc",
										ServicePort: intstr.FromInt(80),
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
