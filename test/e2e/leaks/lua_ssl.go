/*
Copyright 2019 The Kubernetes Authors.

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

package leaks

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"
	pool "gopkg.in/go-playground/pool.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("DynamicCertificates", func() {
	f := framework.NewDefaultFramework("lua-dynamic-certificates")

	BeforeEach(func() {
		f.NewEchoDeployment()
	})

	AfterEach(func() {
	})

	framework.MemoryLeakIt("should not leak memory from ingress SSL certificates or configuration updates", func() {
		hostCount := 1000
		iterations := 10

		By("Waiting a minute before starting the test")
		time.Sleep(1 * time.Minute)

		for iteration := 1; iteration <= iterations; iteration++ {
			By(fmt.Sprintf("Running iteration %v", iteration))

			p := pool.NewLimited(200)

			batch := p.Batch()

			for index := 1; index <= hostCount; index++ {
				host := fmt.Sprintf("hostname-%v", index)
				batch.Queue(run(host, f))
			}

			batch.QueueComplete()
			batch.WaitAll()

			p.Close()

			By("waiting one minute before next iteration")
			time.Sleep(1 * time.Minute)
		}
	})
})

func privisionIngress(hostname string, f *framework.Framework) {
	ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(hostname, "/", hostname, []string{hostname}, f.Namespace, "http-svc", 80, nil))
	_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
		ing.Spec.TLS[0].Hosts,
		ing.Spec.TLS[0].SecretName,
		ing.Namespace)
	Expect(err).NotTo(HaveOccurred())

	f.WaitForNginxServer(hostname,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %v", hostname)) &&
				strings.Contains(server, "listen 443")
		})
}

func checkIngress(hostname string, f *framework.Framework) {
	req := gorequest.New()
	resp, _, errs := req.
		Get(f.GetURL(framework.HTTPS)).
		TLSClientConfig(&tls.Config{ServerName: hostname, InsecureSkipVerify: true}).
		Set("Host", hostname).
		End()
	Expect(errs).Should(BeEmpty())
	Expect(resp.StatusCode).Should(Equal(http.StatusOK))

	// check the returned secret is not the fake one
	cert := resp.TLS.PeerCertificates[0]
	Expect(cert.DNSNames[0]).Should(Equal(hostname))
}

func deleteIngress(hostname string, f *framework.Framework) {
	err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace).Delete(hostname, &metav1.DeleteOptions{})
	Expect(err).NotTo(HaveOccurred(), "unexpected error deleting ingress")
}

func run(host string, f *framework.Framework) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		By(fmt.Sprintf("\tcreating ingress for host %v", host))
		privisionIngress(host, f)

		time.Sleep(100 * time.Millisecond)

		By(fmt.Sprintf("\tchecking ingress for host %v", host))
		checkIngress(host, f)

		By(fmt.Sprintf("\tdestroying ingress for host %v", host))
		deleteIngress(host, f)

		return true, nil
	}
}
