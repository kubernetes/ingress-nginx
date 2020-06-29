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
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Default Backend] SSL", func() {
	f := framework.NewDefaultFramework("default-backend")

	ginkgo.It("should return a self generated SSL certificate", func() {
		ginkgo.By("checking SSL Certificate using the NGINX IP address")
		framework.Sleep()

		resp := f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			Expect().
			Raw()

		assert.Equal(ginkgo.GinkgoT(), len(resp.TLS.PeerCertificates), 1)

		for _, pc := range resp.TLS.PeerCertificates {
			assert.Equal(ginkgo.GinkgoT(), pc.Issuer.CommonName, "Kubernetes Ingress Controller Fake Certificate")
		}

		ginkgo.By("checking SSL Certificate using the NGINX catch all server")
		resp = f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", "foo.bar.com").
			Expect().
			Raw()

		assert.Equal(ginkgo.GinkgoT(), len(resp.TLS.PeerCertificates), 1)
		for _, pc := range resp.TLS.PeerCertificates {
			assert.Equal(ginkgo.GinkgoT(), pc.Issuer.CommonName, "Kubernetes Ingress Controller Fake Certificate")
		}
	})
})
