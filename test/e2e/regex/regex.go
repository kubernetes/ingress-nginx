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

package regex

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	host = "regex"
)

type ingressPath struct {
	path    string
	service string
}

var _ = framework.IngressNginxDescribe("Regex - Match", func() {

	var (
		bi                *v1beta1.Ingress
		matchingTestCases = []struct {
			Name    string
			Path    string
			Service string
		}{
			{"exact match", "/users/otherpath/documents/foo", "http-svc-foo"},
			{"exact match", "/users/demo", "http-svc-demo"},
			{"exact match", "/users/demo/documents", "http-svc-demo-documents"},
			{"regex match", "/users/something/documents", "http-svc-documents"},
			{"regex match", "/users/something/documents/something", "http-svc-documents"},
			{"regex match", "/users/something/records", "http-svc-records"},
			{"regex match", "/users/John.Doe/", "http-svc-users"},
			{"regex match", "/users/John.Doe", "http-svc-users"},
			{"regex match", "/users/", "http-svc-users"},
			{"regex match at the end", "/something/users/demo", "http-svc-demo-end"},
		}
		nonMatchingTestCases = []struct {
			Name string
			Path string
		}{
			{"missing trailing slash", "/users"},
			{"unknown url prefix", "/something/users/otherpath/documents/foo"},
			{"unknown url prefix and suffix", "/something/users/otherpath/documents/foo/somethingelse"},
		}
	)

	f := framework.NewDefaultFramework("match")

	BeforeEach(func() {
		newEchoDeploymentWithServices(f, "http-svc-documents", "http-svc-records", "http-svc-foo",
			"http-svc-demo-documents", "http-svc-users", "http-svc-demo", "http-svc-demo-end")

		bi = buildIngressRewrite(host, f.Namespace.Name,
			ingressPath{path: "/.*(demo)$", service: "http-svc-demo-end"},
			ingressPath{path: "/users/.*/documents", service: "http-svc-documents"},
			ingressPath{path: "/users/demo/documents", service: "http-svc-demo-documents"},
			ingressPath{path: "/users/demo", service: "http-svc-demo"},
			ingressPath{path: "/users/otherpath/documents/foo", service: "http-svc-foo"},
			ingressPath{path: "/users/.*/records", service: "http-svc-records"},
			ingressPath{path: "/users/.*", service: "http-svc-users"},
		)
		bi.Annotations["nginx.ingress.kubernetes.io/location-modifier"] = "~*"

	})

	JustBeforeEach(func() {
		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "location")
			})

		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	Context("with 'nginx.ingress.kubernetes.io/rewrite-target = /'", func() {
		BeforeEach(func() {
			bi.Annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/"
		})

		It("should match the exact location before the regex location.", func() {
			for _, test := range matchingTestCases {
				By(test.Name + " " + test.Path + " should return 200")

				resp, body, errs := gorequest.New().
					Get(f.NginxHTTPURL+test.Path).
					Set("Host", host).
					End()

				Expect(len(errs)).Should(BeNumerically("==", 0))
				Expect(resp.StatusCode).Should(Equal(http.StatusOK))
				Expect(body).Should(ContainSubstring(fmt.Sprintf("Hostname: %s", test.Service)))
			}
		})

		It("should return 404 for no match.", func() {
			for _, test := range nonMatchingTestCases {
				By(test.Name + " " + test.Path + " should return 404")

				resp, body, errs := gorequest.New().
					Get(f.NginxHTTPURL+test.Path).
					Set("Host", host).
					End()

				Expect(len(errs)).Should(BeNumerically("==", 0))
				Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
				Expect(body).Should(ContainSubstring("default backend - 404"))
			}
		})
	})

	Context("without 'nginx.ingress.kubernetes.io/rewrite-target", func() {
		It("should match the exact location before the regex location.", func() {
			for _, test := range matchingTestCases {
				By(test.Name + " " + test.Path + " should return 200")

				resp, body, errs := gorequest.New().
					Get(f.NginxHTTPURL+test.Path).
					Set("Host", host).
					End()

				Expect(len(errs)).Should(BeNumerically("==", 0))
				Expect(resp.StatusCode).Should(Equal(http.StatusOK))
				Expect(body).Should(ContainSubstring(fmt.Sprintf("Hostname: %s", test.Service)))
			}
		})

		It("should return 404 for no match.", func() {
			for _, test := range nonMatchingTestCases {
				By(test.Name + " " + test.Path + " should return 404")

				resp, body, errs := gorequest.New().
					Get(f.NginxHTTPURL+test.Path).
					Set("Host", host).
					End()

				Expect(len(errs)).Should(BeNumerically("==", 0))
				Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
				Expect(body).Should(ContainSubstring("default backend - 404"))
			}
		})
	})
})

func newEchoDeploymentWithServices(f *framework.Framework, serviceNames ...string) {
	for _, serviceName := range serviceNames {
		fmt.Printf("Deploy service %s ", serviceName)
		err := f.NewEchoDeploymentWithServiceName(serviceName)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error %s deploy %s", err, serviceName))
	}
	fmt.Println()
}

func buildIngressRewrite(host string, namespace string, ingressPaths ...ingressPath) *v1beta1.Ingress {
	ingPaths := make([]v1beta1.HTTPIngressPath, len(ingressPaths))
	for i, ingressPath := range ingressPaths {
		ingPaths[i] = v1beta1.HTTPIngressPath{
			Path: ingressPath.path,
			Backend: v1beta1.IngressBackend{
				ServiceName: ingressPath.service,
				ServicePort: intstr.FromInt(80),
			},
		}
	}

	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        host,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: ingPaths,
						},
					},
				},
			},
		},
	}
}
