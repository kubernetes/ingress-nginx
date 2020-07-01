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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	pool "gopkg.in/go-playground/pool.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Memory Leak] Dynamic Certificates", func() {
	f := framework.NewDefaultFramework("lua-dynamic-certificates")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	framework.MemoryLeakIt("should not leak memory from ingress SSL certificates or configuration updates", func() {
		hostCount := 1000
		iterations := 10

		ginkgo.By("Waiting a minute before starting the test")
		framework.Sleep(1 * time.Minute)

		for iteration := 1; iteration <= iterations; iteration++ {
			ginkgo.By(fmt.Sprintf("Running iteration %v", iteration))

			p := pool.NewLimited(200)

			batch := p.Batch()

			for index := 1; index <= hostCount; index++ {
				host := fmt.Sprintf("hostname-%v", index)
				batch.Queue(run(host, f))
			}

			batch.QueueComplete()
			batch.WaitAll()

			p.Close()

			ginkgo.By("waiting one minute before next iteration")
			framework.Sleep(1 * time.Minute)
		}
	})
})

func privisionIngress(hostname string, f *framework.Framework) {
	ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(hostname, "/", hostname, []string{hostname}, f.Namespace, framework.EchoService, 80, nil))
	_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
		ing.Spec.TLS[0].Hosts,
		ing.Spec.TLS[0].SecretName,
		ing.Namespace)
	assert.Nil(ginkgo.GinkgoT(), err)

	f.WaitForNginxServer(hostname,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %v", hostname)) &&
				strings.Contains(server, "listen 443")
		})
}

func checkIngress(hostname string, f *framework.Framework) {
	resp := f.HTTPTestClientWithTLSConfig(&tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true,
	}).
		GET("/").
		WithURL(f.GetURL(framework.HTTPS)).
		WithHeader("Host", hostname).
		Expect().
		Raw()

	assert.Equal(ginkgo.GinkgoT(), resp.StatusCode, http.StatusOK)

	// check the returned secret is not the fake one
	cert := resp.TLS.PeerCertificates[0]
	assert.Equal(ginkgo.GinkgoT(), cert.DNSNames[0], hostname)
}

func deleteIngress(hostname string, f *framework.Framework) {
	err := f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Delete(context.TODO(), hostname, metav1.DeleteOptions{})
	assert.Nil(ginkgo.GinkgoT(), err, "unexpected error deleting ingress")
}

func run(host string, f *framework.Framework) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		ginkgo.By(fmt.Sprintf("\tcreating ingress for host %v", host))
		privisionIngress(host, f)

		framework.Sleep(100 * time.Millisecond)

		ginkgo.By(fmt.Sprintf("\tchecking ingress for host %v", host))
		checkIngress(host, f)

		ginkgo.By(fmt.Sprintf("\tdestroying ingress for host %v", host))
		deleteIngress(host, f)

		return true, nil
	}
}
