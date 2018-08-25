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
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	logDynamicConfigSuccess = "Dynamic reconfiguration succeeded"
	logDynamicConfigFailure = "Dynamic reconfiguration failed"
	logRequireBackendReload = "Configuration changes detected, backend reload required"
	logBackendReloadSuccess = "Backend successfully reloaded"
	logSkipBackendReload    = "Changes handled by the dynamic configuration, skipping backend reload"
	logInitialConfigSync    = "Initial synchronization of the NGINX configuration"
	waitForLuaSync          = 5 * time.Second
)

var _ = framework.IngressNginxDescribe("Dynamic Configuration", func() {
	f := framework.NewDefaultFramework("dynamic-configuration")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(1)
		Expect(err).NotTo(HaveOccurred())

		ensureIngress(f, "foo.com")
	})

	It("should set nameservers for Lua", func() {
		err := f.WaitForNginxConfiguration(func(cfg string) bool {
			r := regexp.MustCompile(`configuration.nameservers = { [".,0-9a-zA-Z]+ }`)
			return r.MatchString(cfg)
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when only backends change", func() {
		It("handles endpoints only changes", func() {
			var nginxConfig string
			err := f.WaitForNginxConfiguration(func(cfg string) bool {
				nginxConfig = cfg
				return true
			})

			replicas := 2
			err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", replicas, nil)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(waitForLuaSync)

			ensureRequest(f, "foo.com")

			var newNginxConfig string
			err = f.WaitForNginxConfiguration(func(cfg string) bool {
				newNginxConfig = cfg
				return true
			})

			Expect(nginxConfig).Should(Equal(newNginxConfig))
		})

		It("handles an annotation change", func() {
			var nginxConfig string
			err := f.WaitForNginxConfiguration(func(cfg string) bool {
				nginxConfig = cfg
				return true
			})

			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/load-balance"] = "round_robin"
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(waitForLuaSync)

			ensureRequest(f, "foo.com")

			var newNginxConfig string
			err = f.WaitForNginxConfiguration(func(cfg string) bool {
				newNginxConfig = cfg
				return true
			})

			Expect(nginxConfig).Should(Equal(newNginxConfig))
		})
	})

	It("handles a non backend update", func() {
		var nginxConfig string
		err := f.WaitForNginxConfiguration(func(cfg string) bool {
			nginxConfig = cfg
			return true
		})

		ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		ingress.Spec.TLS = []extensions.IngressTLS{
			{
				Hosts:      []string{"foo.com"},
				SecretName: "foo.com",
			},
		}
		_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ingress.Spec.TLS[0].Hosts,
			ingress.Spec.TLS[0].SecretName,
			ingress.Namespace)
		Expect(err).ToNot(HaveOccurred())
		_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
		Expect(err).ToNot(HaveOccurred())

		var newNginxConfig string
		err = f.WaitForNginxConfiguration(func(cfg string) bool {
			newNginxConfig = cfg
			return true
		})

		Expect(nginxConfig).ShouldNot(Equal(newNginxConfig))
	})
})

func ensureIngress(f *framework.Framework, host string) *extensions.Ingress {
	ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80,
		&map[string]string{"nginx.ingress.kubernetes.io/load-balance": "ewma"}))
	Expect(err).NotTo(HaveOccurred())
	Expect(ing).NotTo(BeNil())
	err = f.WaitForNginxServer(host,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %s ;", host)) &&
				strings.Contains(server, "proxy_pass http://upstream_balancer;")
		})
	time.Sleep(waitForLuaSync)
	ensureRequest(f, host)

	return ing
}

func ensureRequest(f *framework.Framework, host string) {
	resp, _, errs := gorequest.New().
		Get(f.IngressController.HTTPURL).
		Set("Host", host).
		End()
	Expect(errs).Should(BeEmpty())
	Expect(resp.StatusCode).Should(Equal(http.StatusOK))
}

func ensureHTTPSRequest(url string, host string, expectedDNSName string) {
	resp, _, errs := gorequest.New().
		Get(url).
		Set("Host", host).
		TLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}).
		End()
	Expect(errs).Should(BeEmpty())
	Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	Expect(len(resp.TLS.PeerCertificates)).Should(BeNumerically("==", 1))
	Expect(resp.TLS.PeerCertificates[0].DNSNames[0]).Should(Equal(expectedDNSName))
}

func getCookie(name string, cookies []*http.Cookie) (*http.Cookie, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return &http.Cookie{}, fmt.Errorf("Cookie does not exist")
}
