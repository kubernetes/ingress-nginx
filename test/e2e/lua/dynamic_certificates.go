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
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Lua] dynamic certificates", func() {
	f := framework.NewDefaultFramework("dynamic-certificate")
	host := "foo.com"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("picks up the certificate when we add TLS spec to existing ingress", func() {
		ensureIngress(f, host, framework.EchoService)

		ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), host, metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err)
		ing.Spec.TLS = []networking.IngressTLS{
			{
				Hosts:      []string{host},
				SecretName: host,
			},
		}
		_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err)

		time.Sleep(waitForLuaSync)

		ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host, host)
	})

	ginkgo.It("picks up the previously missing secret for a given ingress without reloading", func() {
		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		time.Sleep(waitForLuaSync)

		ip := f.GetNginxPodIP()
		mf, err := f.GetMetric("nginx_ingress_controller_success", ip)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.NotNil(ginkgo.GinkgoT(), mf)

		rc0, err := extractReloadCount(mf)
		assert.Nil(ginkgo.GinkgoT(), err)

		ensureHTTPSRequest(f, fmt.Sprintf("%s?id=dummy_log_splitter_foo_bar", f.GetURL(framework.HTTPS)), host, "ingress.local")

		_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		time.Sleep(waitForLuaSync)

		ginkgo.By("serving the configured certificate on HTTPS endpoint")
		ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host, host)

		log, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.NotEmpty(ginkgo.GinkgoT(), log)

		ginkgo.By("skipping Nginx reload")
		mf, err = f.GetMetric("nginx_ingress_controller_success", ip)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.NotNil(ginkgo.GinkgoT(), mf)

		rc1, err := extractReloadCount(mf)
		assert.Nil(ginkgo.GinkgoT(), err)

		assert.Equal(ginkgo.GinkgoT(), rc0, rc1)
	})

	ginkgo.Context("given an ingress with TLS correctly configured", func() {
		ginkgo.BeforeEach(func() {
			ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))

			time.Sleep(waitForLuaSync)

			ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host, "ingress.local")

			_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			time.Sleep(waitForLuaSync)

			ginkgo.By("configuring certificate_by_lua and skipping Nginx configuration of the new certificate")
			f.WaitForNginxServer(ing.Spec.TLS[0].Hosts[0],
				func(server string) bool {
					return strings.Contains(server, "listen 443")
				})

			time.Sleep(waitForLuaSync)

			ginkgo.By("serving the configured certificate on HTTPS endpoint")
			ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host, host)
		})

		/*
			TODO(elvinefendi): this test currently does not work as expected
			because Go transport code strips (https://github.com/golang/go/blob/431b5c69ca214ce4291f008c1ce2a50b22bc2d2d/src/crypto/tls/handshake_messages.go#L424)
			trailing dot from SNI as suggest by the standard (https://tools.ietf.org/html/rfc6066#section-3).
		*/
		ginkgo.It("supports requests with domain with trailing dot", func() {
			ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host+".", host)
		})

		ginkgo.It("picks up the updated certificate without reloading", func() {
			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), host, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ensureHTTPSRequest(f, fmt.Sprintf("%s?id=dummy_log_splitter_foo_bar", f.GetURL(framework.HTTPS)), host, host)

			_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			assert.Nil(ginkgo.GinkgoT(), err)

			time.Sleep(waitForLuaSync)

			ginkgo.By("configuring certificate_by_lua and skipping Nginx configuration of the new certificate")
			f.WaitForNginxServer(ing.Spec.TLS[0].Hosts[0],
				func(server string) bool {
					return strings.Contains(server, "listen 443")
				})

			ginkgo.By("serving the configured certificate on HTTPS endpoint")
			ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host, host)

			log, err := f.NginxLogs()
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.NotEmpty(ginkgo.GinkgoT(), log)

			index := strings.Index(log, "id=dummy_log_splitter_foo_bar")
			assert.GreaterOrEqual(ginkgo.GinkgoT(), index, 0, "log does not contains id=dummy_log_splitter_foo_bar")
			restOfLogs := log[index:]

			ginkgo.By("skipping Nginx reload")
			assert.NotContains(ginkgo.GinkgoT(), restOfLogs, logRequireBackendReload)
			assert.NotContains(ginkgo.GinkgoT(), restOfLogs, logBackendReloadSuccess)
		})

		ginkgo.It("falls back to using default certificate when secret gets deleted without reloading", func() {
			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), host, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ensureHTTPSRequest(f, fmt.Sprintf("%s?id=dummy_log_splitter_foo_bar", f.GetURL(framework.HTTPS)), host, host)

			ip := f.GetNginxPodIP()
			mf, err := f.GetMetric("nginx_ingress_controller_success", ip)
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.NotNil(ginkgo.GinkgoT(), mf)

			rc0, err := extractReloadCount(mf)
			assert.Nil(ginkgo.GinkgoT(), err)

			err = f.KubeClientSet.CoreV1().Secrets(ing.Namespace).Delete(context.TODO(), ing.Spec.TLS[0].SecretName, metav1.DeleteOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			time.Sleep(waitForLuaSync)

			ginkgo.By("serving the default certificate on HTTPS endpoint")
			ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), host, "ingress.local")

			mf, err = f.GetMetric("nginx_ingress_controller_success", ip)
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.NotNil(ginkgo.GinkgoT(), mf)

			rc1, err := extractReloadCount(mf)
			assert.Nil(ginkgo.GinkgoT(), err)

			ginkgo.By("skipping Nginx reload")
			assert.Equal(ginkgo.GinkgoT(), rc0, rc1)
		})

		ginkgo.It("picks up a non-certificate only change", func() {
			newHost := "foo2.com"
			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), host, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ing.Spec.Rules[0].Host = newHost
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			time.Sleep(waitForLuaSync)

			ginkgo.By("serving the configured certificate on HTTPS endpoint")
			ensureHTTPSRequest(f, f.GetURL(framework.HTTPS), newHost, "ingress.local")
		})

		ginkgo.It("removes HTTPS configuration when we delete TLS spec", func() {
			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), host, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			ing.Spec.TLS = []networking.IngressTLS{}
			_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			time.Sleep(waitForLuaSync)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)

		})
	})
})

func extractReloadCount(mf *dto.MetricFamily) (float64, error) {
	vec, err := expfmt.ExtractSamples(&expfmt.DecodeOptions{
		Timestamp: model.Now(),
	}, mf)

	if err != nil {
		return 0, err
	}

	return float64(vec[0].Value), nil
}
