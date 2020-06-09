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

package settings

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("use-proxy-protocol", func() {
	f := framework.NewDefaultFramework("proxy-protocol")

	setting := "use-proxy-protocol"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.UpdateNginxConfigMapData(setting, "false")
	})

	ginkgo.It("should respect port passed by the PROXY Protocol", func() {
		host := "proxy-protocol"

		f.UpdateNginxConfigMapData(setting, "true")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name proxy-protocol") &&
					strings.Contains(server, "listen 80 proxy_protocol")
			})

		ip := f.GetNginxIP()

		conn, err := net.Dial("tcp", net.JoinHostPort(ip, "80"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error creating connection to %s:80", ip)
		defer conn.Close()

		header := "PROXY TCP4 192.168.0.1 192.168.0.11 56324 1234\r\n"
		conn.Write([]byte(header))
		conn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))

		data, err := ioutil.ReadAll(conn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		body := string(data)
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=%v", "proxy-protocol"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-port=1234"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-proto=http"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-for=192.168.0.1"))
	})

	ginkgo.It("should respect proto passed by the PROXY Protocol server port", func() {
		host := "proxy-protocol"

		f.UpdateNginxConfigMapData(setting, "true")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name proxy-protocol") &&
					strings.Contains(server, "listen 80 proxy_protocol")
			})

		ip := f.GetNginxIP()

		conn, err := net.Dial("tcp", net.JoinHostPort(ip, "80"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error creating connection to %s:80", ip)
		defer conn.Close()

		header := "PROXY TCP4 192.168.0.1 192.168.0.11 56324 443\r\n"
		conn.Write([]byte(header))
		conn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))

		data, err := ioutil.ReadAll(conn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		body := string(data)
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=%v", "proxy-protocol"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-port=443"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-proto=https"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-for=192.168.0.1"))
	})

	ginkgo.It("should enable PROXY Protocol for HTTPS", func() {
		host := "proxy-protocol"

		f.UpdateNginxConfigMapData(setting, "true")

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil))
		tlsConfig, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "443 proxy_protocol")
			})

		ip := f.GetNginxIP()

		conn, err := net.Dial("tcp", net.JoinHostPort(ip, "443"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error connecting to %v:443", ip)
		defer conn.Close()

		_, err = fmt.Fprintf(conn, "PROXY TCP4 192.168.0.1 192.168.0.11 56324 1234\r\n")
		assert.Nil(ginkgo.GinkgoT(), err, "writing proxy protocol")

		tlsConn := tls.Client(conn, tlsConfig)
		defer tlsConn.Close()

		_, err = tlsConn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))
		assert.Nil(ginkgo.GinkgoT(), err, "writing HTTP request")

		data, err := ioutil.ReadAll(tlsConn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		body := string(data)
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=%v", "proxy-protocol"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-port=1234"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-proto=https"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-scheme=https"))
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("x-forwarded-for=192.168.0.1"))
	})
})
