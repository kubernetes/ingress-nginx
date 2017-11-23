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

package ssl

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("SSL", func() {
	f := framework.NewDefaultFramework("ssl")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should not appear references to secret updates not used in ingress rules", func() {
		host := "ssl-update"

		dummySecret, err := f.EnsureSecret(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: f.Namespace.Name,
			},
			Data: map[string][]byte{
				"key": []byte("value"),
			},
		})
		Expect(err).NotTo(HaveOccurred())

		ing, err := f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.Namespace.Name,
			},
			Spec: v1beta1.IngressSpec{
				TLS: []v1beta1.IngressTLS{
					{
						Hosts:      []string{host},
						SecretName: host,
					},
				},
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
								},
							},
						},
					},
				},
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(ing).ToNot(BeNil())

		_, _, _, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).ToNot(HaveOccurred())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name ssl-update") &&
					strings.Contains(server, "listen 443")
			})
		Expect(err).ToNot(HaveOccurred())

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("starting syncing of secret %v/dummy", f.Namespace.Name)))
		time.Sleep(30 * time.Second)
		dummySecret.Data["some-key"] = []byte("some value")
		f.KubeClientSet.CoreV1().Secrets(f.Namespace.Name).Update(dummySecret)
		time.Sleep(30 * time.Second)
		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("starting syncing of secret %v/dummy", f.Namespace.Name)))
		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("error obtaining PEM from secret %v/dummy", f.Namespace.Name)))
	})
})
