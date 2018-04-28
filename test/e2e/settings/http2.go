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
	"crypto/tls"
	"golang.org/x/net/http2"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Settings - HTTP2)", func() {

	f := framework.NewDefaultFramework("settings-http2")
	host := "settings-http2"
	blacklistKey := "http2-host-blacklist"

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should allow HTTP2 by default", func() {
		tlsConfig, err := tlsEndpoint(f, host)
		Expect(err).NotTo(HaveOccurred())

		err = framework.WaitForTLS(f.IngressController.HTTPSURL, tlsConfig)
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxServer(host,
			func(servercfg string) bool {
				return strings.Contains(servercfg, "http2;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := http2Request(f, tlsConfig, host)
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.ProtoMajor).Should(BeNumerically("==", 2))
	})

	It("should disable HTTP2 for blacklisted hosts", func() {
		tlsConfig, err := tlsEndpoint(f, host)
		Expect(err).NotTo(HaveOccurred())

		err = framework.WaitForTLS(f.IngressController.HTTPSURL, tlsConfig)
		Expect(err).NotTo(HaveOccurred())

		By("setting blacklist")

		err = f.UpdateNginxConfigMapData(blacklistKey, host)
		Expect(err).NotTo(HaveOccurred())

		err = f.WaitForNginxServer(host,
			func(servercfg string) bool {
				return !strings.Contains(servercfg, "http2;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := http2Request(f, tlsConfig, host)
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(resp.ProtoMajor).Should(BeNumerically("==", 1))
	})
})

func http2Request(f *framework.Framework, tlsConfig *tls.Config, host string) (*http.Response, string, []error) {
	request := gorequest.New().
		Get(f.IngressController.HTTPSURL).
		TLSClientConfig(tlsConfig).
		Set("Host", host)
	err := http2.ConfigureTransport(request.Transport)
	Expect(err).NotTo(HaveOccurred())
	return request.End()
}
