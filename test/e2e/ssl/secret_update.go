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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("SSL", func() {
	f := framework.NewDefaultFramework("ssl")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should not appear references to secret updates not used in ingress rules", func() {
		host := "ssl-update"

		dummySecret := f.EnsureSecret(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: f.IngressController.Namespace,
			},
			Data: map[string][]byte{
				"key": []byte("value"),
			},
		})

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.IngressController.Namespace, "http-svc", 80, nil))
		_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).ToNot(HaveOccurred())

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name ssl-update") &&
					strings.Contains(server, "listen 443")
			})

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("starting syncing of secret %v/dummy", f.IngressController.Namespace)))
		time.Sleep(5 * time.Second)
		dummySecret.Data["some-key"] = []byte("some value")
		f.KubeClientSet.CoreV1().Secrets(f.IngressController.Namespace).Update(dummySecret)
		time.Sleep(5 * time.Second)
		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("starting syncing of secret %v/dummy", f.IngressController.Namespace)))
		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("error obtaining PEM from secret %v/dummy", f.IngressController.Namespace)))
	})
})
