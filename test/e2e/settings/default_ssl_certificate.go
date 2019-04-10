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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1beta1 "k8s.io/api/apps/v1beta1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Default SSL Certificate", func() {
	f := framework.NewDefaultFramework("default-ssl-certificate")
	secretName := "my-custom-cert"

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)

		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			[]string{"*"},
			secretName,
			f.Namespace)
		Expect(err).NotTo(HaveOccurred())

		framework.UpdateDeployment(f.KubeClientSet, f.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, "--default-ssl-certificate=$(POD_NAMESPACE)/"+secretName)
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.Namespace).Update(deployment)

				return err
			})

		// this asserts that it configures default custom ssl certificate without an ingress at all
		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)
	})

	It("configures ssl certificate for catch-all ingress", func() {
		ing := framework.NewSingleCatchAllIngress("catch-all", f.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		sslCertificate := fmt.Sprintf("ssl_certificate /etc/ingress-controller/ssl/%s-%s.pem;", f.Namespace, secretName)
		sslCertificateKey := fmt.Sprintf("ssl_certificate_key /etc/ingress-controller/ssl/%s-%s.pem;", f.Namespace, secretName)
		f.WaitForNginxServer("_", func(cfg string) bool {
			return strings.Contains(cfg, sslCertificate) && strings.Contains(cfg, sslCertificateKey)
		})
	})

	It("configures ssl certificate for host based ingress with tls spec", func() {
		host := "foo"

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "http-svc", 80, nil))
		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).NotTo(HaveOccurred())

		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)

		sslCertificate := fmt.Sprintf("ssl_certificate /etc/ingress-controller/ssl/%s-%s.pem;", f.Namespace, secretName)
		sslCertificateKey := fmt.Sprintf("ssl_certificate_key /etc/ingress-controller/ssl/%s-%s.pem;", f.Namespace, secretName)
		f.WaitForNginxServer(host, func(cfg string) bool {
			return strings.Contains(cfg, "server_name foo") && strings.Contains(cfg, sslCertificate) && strings.Contains(cfg, sslCertificateKey)
		})
	})
})
