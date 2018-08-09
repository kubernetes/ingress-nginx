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

package multipleingress

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Multiple Ingress - Same Service", func() {
	f := framework.NewDefaultFramework("multiple-ingress")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should add server entry for both the ingress", func() {
		ingress1spec := buildIngress("ingress-1.example.com", f.IngressController.Namespace, "/", "http-svc", 80)
		ingress1, err := f.EnsureIngress(ingress1spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(ingress1).NotTo(BeNil())

		//add permanent redirect annotation to 2nd ingress
		redirectPath := "/something"
		redirectURL := "http://redirect.example.com"
		ingress2spec := buildIngress("ingress-2.example.com", f.IngressController.Namespace, redirectPath, "http-svc", 80)
		ingress2spec.Annotations = map[string]string{
			"nginx.ingress.kubernetes.io/permanent-redirect": redirectURL,
		}

		ingress2, err := f.EnsureIngress(ingress2spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(ingress2).NotTo(BeNil())

		err = f.WaitForNginxServer("ingress-1.example.com",
			func(server string) bool {
				fmt.Println(server)
				fmt.Println("")
				return strings.Contains(server, "proxy_pass http://upstream_balancer")
			})
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxServer("ingress-2.example.com",
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("return 301 %s;", redirectURL))
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", "ingress-1.example.com").
			End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(200))

		resp, _, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+redirectPath).
			Set("Host", "ingress-2.example.com").
			RedirectPolicy(noRedirectPolicyFunc).
			End()

		Expect(errs).To(BeNil())
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(BeNumerically("==", http.StatusMovedPermanently))
		Expect(resp.Header.Get("Location")).Should(Equal(redirectURL))
	})
})

func noRedirectPolicyFunc(gorequest.Request, []gorequest.Request) error {
	return http.ErrUseLastResponse
}

func buildIngress(host, namespace, path, backendService string, port int) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: path,
									Backend: v1beta1.IngressBackend{
										ServiceName: backendService,
										ServicePort: intstr.FromInt(port),
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
