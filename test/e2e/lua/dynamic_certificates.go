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

package lua

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Dynamic Certificate", func() {
	f := framework.NewDefaultFramework("dynamic-certificate")
	host := "foo.com"

	BeforeEach(func() {
		err := framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, "--enable-dynamic-certificates")
				args = append(args, "--enable-ssl-chain-completion=false")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.IngressController.Namespace).Update(deployment)

				return err
			})
		Expect(err).NotTo(HaveOccurred())

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "ok, res = pcall(require, \"certificate\")")
			})

		f.NewEchoDeploymentWithReplicas(1)
	})

	It("picks up the certificate when we add TLS spec to existing ingress", func() {
		ensureIngress(f, host, "http-svc")

		ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		ing.Spec.TLS = []extensions.IngressTLS{
			{
				Hosts:      []string{host},
				SecretName: host,
			},
		}
		_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).ToNot(HaveOccurred())
		_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ing)
		Expect(err).ToNot(HaveOccurred())
		time.Sleep(waitForLuaSync)

		ensureHTTPSRequest(f.IngressController.HTTPSURL, host, host)
	})

	It("picks up the previously missing secret for a given ingress without reloading", func() {
		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.IngressController.Namespace, "http-svc", 80, nil)
		f.EnsureIngress(ing)

		time.Sleep(waitForLuaSync)

		ensureHTTPSRequest(fmt.Sprintf("%s?id=dummy_log_splitter_foo_bar", f.IngressController.HTTPSURL), host, "ingress.local")

		_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		Expect(err).ToNot(HaveOccurred())

		By("configuring certificate_by_lua and skipping Nginx configuration of the new certificate")
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "ssl_certificate_by_lua_block") &&
					!strings.Contains(server, fmt.Sprintf("ssl_certificate /etc/ingress-controller/ssl/%s-%s.pem;", ing.Namespace, host)) &&
					!strings.Contains(server, fmt.Sprintf("ssl_certificate_key /etc/ingress-controller/ssl/%s-%s.pem;", ing.Namespace, host)) &&
					strings.Contains(server, "listen 443")
			})

		time.Sleep(waitForLuaSync)

		By("serving the configured certificate on HTTPS endpoint")
		ensureHTTPSRequest(f.IngressController.HTTPSURL, host, host)

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())
		index := strings.Index(log, "id=dummy_log_splitter_foo_bar")
		restOfLogs := log[index:]

		By("skipping Nginx reload")
		Expect(restOfLogs).ToNot(ContainSubstring(logRequireBackendReload))
		Expect(restOfLogs).ToNot(ContainSubstring(logBackendReloadSuccess))
	})

	Context("given an ingress with TLS correctly configured", func() {
		BeforeEach(func() {
			ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.IngressController.Namespace, "http-svc", 80, nil))

			time.Sleep(waitForLuaSync)

			ensureHTTPSRequest(f.IngressController.HTTPSURL, host, "ingress.local")

			_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(waitForLuaSync)

			By("configuring certificate_by_lua and skipping Nginx configuration of the new certificate")
			f.WaitForNginxServer(ing.Spec.TLS[0].Hosts[0],
				func(server string) bool {
					return strings.Contains(server, "ssl_certificate_by_lua_block") &&
						!strings.Contains(server, fmt.Sprintf("ssl_certificate /etc/ingress-controller/ssl/%s-%s.pem;", ing.Namespace, host)) &&
						!strings.Contains(server, fmt.Sprintf("ssl_certificate_key /etc/ingress-controller/ssl/%s-%s.pem;", ing.Namespace, host)) &&
						strings.Contains(server, "listen 443")
				})

			time.Sleep(waitForLuaSync)

			By("serving the configured certificate on HTTPS endpoint")
			ensureHTTPSRequest(f.IngressController.HTTPSURL, host, host)
		})

		It("picks up the updated certificate without reloading", func() {
			ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			ensureHTTPSRequest(fmt.Sprintf("%s?id=dummy_log_splitter_foo_bar", f.IngressController.HTTPSURL), host, host)

			_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(waitForLuaSync)

			By("configuring certificate_by_lua and skipping Nginx configuration of the new certificate")
			f.WaitForNginxServer(ing.Spec.TLS[0].Hosts[0],
				func(server string) bool {
					return strings.Contains(server, "ssl_certificate_by_lua_block") &&
						!strings.Contains(server, fmt.Sprintf("ssl_certificate /etc/ingress-controller/ssl/%s-%s.pem;", ing.Namespace, host)) &&
						!strings.Contains(server, fmt.Sprintf("ssl_certificate_key /etc/ingress-controller/ssl/%s-%s.pem;", ing.Namespace, host)) &&
						strings.Contains(server, "listen 443")
				})

			time.Sleep(waitForLuaSync)

			By("serving the configured certificate on HTTPS endpoint")
			ensureHTTPSRequest(f.IngressController.HTTPSURL, host, host)

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "id=dummy_log_splitter_foo_bar")
			restOfLogs := log[index:]

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring(logRequireBackendReload))
			Expect(restOfLogs).ToNot(ContainSubstring(logBackendReloadSuccess))
		})

		It("falls back to using default certificate when secret gets deleted without reloading", func() {
			ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})

			ensureHTTPSRequest(fmt.Sprintf("%s?id=dummy_log_splitter_foo_bar", f.IngressController.HTTPSURL), host, host)

			f.KubeClientSet.CoreV1().Secrets(ing.Namespace).Delete(ing.Spec.TLS[0].SecretName, nil)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(waitForLuaSync)

			By("configuring certificate_by_lua and skipping Nginx configuration of the new certificate")
			f.WaitForNginxServer(ing.Spec.TLS[0].Hosts[0],
				func(server string) bool {
					return strings.Contains(server, "ssl_certificate_by_lua_block") &&
						strings.Contains(server, "ssl_certificate /etc/ingress-controller/ssl/default-fake-certificate.pem;") &&
						strings.Contains(server, "ssl_certificate_key /etc/ingress-controller/ssl/default-fake-certificate.pem;") &&
						strings.Contains(server, "listen 443")
				})

			time.Sleep(waitForLuaSync)

			By("serving the default certificate on HTTPS endpoint")
			ensureHTTPSRequest(f.IngressController.HTTPSURL, host, "ingress.local")

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "id=dummy_log_splitter_foo_bar")
			restOfLogs := log[index:]

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring(logRequireBackendReload))
			Expect(restOfLogs).ToNot(ContainSubstring(logBackendReloadSuccess))
		})

		It("picks up a non-certificate only change", func() {
			newHost := "foo2.com"
			ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			ing.Spec.Rules[0].Host = newHost
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ing)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(waitForLuaSync)

			By("serving the configured certificate on HTTPS endpoint")
			ensureHTTPSRequest(f.IngressController.HTTPSURL, newHost, "ingress.local")
		})

		It("removes HTTPS configuration when we delete TLS spec", func() {
			ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			ing.Spec.TLS = []extensions.IngressTLS{}
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ing)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(waitForLuaSync)

			ensureRequest(f, host)
		})
	})
})
