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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[SSL] secret update", func() {
	f := framework.NewDefaultFramework("ssl")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should not appear references to secret updates not used in ingress rules", func() {
		host := "ssl-update"

		dummySecret := f.EnsureSecret(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: f.Namespace,
			},
			Data: map[string][]byte{
				"key": []byte("value"),
			},
		})

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))
		_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		framework.Sleep()

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name ssl-update") &&
					strings.Contains(server, "listen 443")
			})

		log, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.NotContains(ginkgo.GinkgoT(), log, fmt.Sprintf("starting syncing of secret %v/dummy", f.Namespace))

		framework.Sleep()

		dummySecret.Data["some-key"] = []byte("some value")

		f.KubeClientSet.CoreV1().Secrets(f.Namespace).Update(context.TODO(), dummySecret, metav1.UpdateOptions{})

		assert.NotContains(ginkgo.GinkgoT(), log, fmt.Sprintf("starting syncing of secret %v/dummy", f.Namespace))
		assert.NotContains(ginkgo.GinkgoT(), log, fmt.Sprintf("error obtaining PEM from secret %v/dummy", f.Namespace))
	})

	ginkgo.It("should return the fake SSL certificate if the secret is invalid", func() {
		host := "invalid-ssl"

		// create a secret without cert or key
		f.EnsureSecret(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.Namespace,
			},
		})

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name invalid-ssl") &&
					strings.Contains(server, "listen 443")
			})

		resp := f.HTTPTestClientWithTLSConfig(&tls.Config{ServerName: host, InsecureSkipVerify: true}).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Raw()

		// check the returned secret is the fake one
		cert := resp.TLS.PeerCertificates[0]

		assert.Equal(ginkgo.GinkgoT(), len(resp.TLS.PeerCertificates), 1)
		for _, pc := range resp.TLS.PeerCertificates {
			assert.Equal(ginkgo.GinkgoT(), pc.Issuer.CommonName, "Kubernetes Ingress Controller Fake Certificate")
		}

		assert.Equal(ginkgo.GinkgoT(), cert.DNSNames[0], "ingress.local")
		assert.Equal(ginkgo.GinkgoT(), cert.Subject.Organization[0], "Acme Co")
		assert.Equal(ginkgo.GinkgoT(), cert.Subject.CommonName, "Kubernetes Ingress Controller Fake Certificate")

		// verify the log contains a warning about invalid certificate
		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, fmt.Sprintf("%v/invalid-ssl\" contains no keypair or CA certificate", f.Namespace))
	})
})
