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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Affinity", func() {
	f := framework.NewDefaultFramework("affinity")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(1)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should set sticky cookie SERVERID", func() {
		host := "sticky.foo.com"

		bi := buildIngress(host, f.Namespace.Name)
		bi.Annotations["ingress.kubernetes.io/affinity"] = "cookie"
		bi.Annotations["ingress.kubernetes.io/session-cookie-name"] = "SERVERID"

		ing, err := f.EnsureIngress(bi)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		stickyUpstreamName := "sticky-"+f.Namespace.Name+"-http-svc-80"
		upstreamPrefix := fmt.Sprintf(`
    upstream %v {
        sticky hash=md5 name=SERVERID  httponly;`, stickyUpstreamName)
		err = f.WaitForNginxServer(host,
			func(cfg string) bool {
				return strings.Contains(cfg, "proxy_pass http://"+stickyUpstreamName+";") &&
					strings.Contains(cfg, upstreamPrefix)
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.NginxHTTPURL).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("SERVERID="))
	})

	XIt("should redirect to '/something' with enabled affinity", func() {
		host := "example.com"

		ing, err := f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.Namespace.Name,
				Annotations: map[string]string{
					"ingress.kubernetes.io/affinity": "cookie",
					"ingress.kubernetes.io/session-cookie-name": "SERVERID",
					"ingress.kubernetes.io/rewrite-target": "/something",
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
											ServicePort: intstr.FromInt(8080),
										},
									},
								},
							},
						},
					},
				},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		stickyUpstreamName := "sticky-"+f.Namespace.Name+"-http-svc-8080;"
		proxyPass := fmt.Sprintf(`
	rewrite /(.*) /something/$1 break;
	proxy_pass http://%v;`,stickyUpstreamName)
		upstreamPrefix := fmt.Sprintf(`
    upstream %v {
        sticky hash=sha1 name=SERVERID  httponly;`, stickyUpstreamName)
		err = f.WaitForNginxServer(host,
			func(cfg string) bool {
				return strings.Contains(cfg, proxyPass) &&
					strings.Contains(cfg, upstreamPrefix)
			})
		Expect(err).NotTo(HaveOccurred())

		resp, body, errs := gorequest.New().
			Get(f.NginxHTTPURL).
			Set("Host", host).
			End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring("real path=/something"))
		Expect(resp.Header.Get("Set-Cookie")).Should(ContainSubstring("SERVERID="))
	})
})
