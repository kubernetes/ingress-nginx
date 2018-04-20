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

package defaultbackend

import (
	"crypto/tls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Default backend - SSL", func() {
	f := framework.NewDefaultFramework("default-backend")

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should return a self generated SSL certificate", func() {
		By("checking SSL Certificate using the NGINX IP address")
		resp, _, errs := gorequest.New().
			Post(f.IngressController.HTTPSURL).
			TLSClientConfig(&tls.Config{
				// the default backend uses a self generated certificate
				InsecureSkipVerify: true,
			}).End()

		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(len(resp.TLS.PeerCertificates)).Should(BeNumerically("==", 1))

		for _, pc := range resp.TLS.PeerCertificates {
			Expect(pc.Issuer.CommonName).Should(Equal("Kubernetes Ingress Controller Fake Certificate"))
		}

		By("checking SSL Certificate using the NGINX catch all server")
		resp, _, errs = gorequest.New().
			Post(f.IngressController.HTTPSURL).
			TLSClientConfig(&tls.Config{
				// the default backend uses a self generated certificate
				InsecureSkipVerify: true,
			}).
			Set("Host", "foo.bar.com").End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(len(resp.TLS.PeerCertificates)).Should(BeNumerically("==", 1))
		for _, pc := range resp.TLS.PeerCertificates {
			Expect(pc.Issuer.CommonName).Should(Equal("Kubernetes Ingress Controller Fake Certificate"))
		}
	})
})
