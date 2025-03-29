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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const proxyProtocol = "proxy-protocol"

var _ = framework.DescribeSetting("use-proxy-protocol", func() {
	f := framework.NewDefaultFramework(proxyProtocol)

	setting := "use-proxy-protocol"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.UpdateNginxConfigMapData(setting, "false")
	})
	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should respect port passed by the PROXY Protocol", func() {
		host := proxyProtocol

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
		_, err = conn.Write([]byte(header))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error writing header")

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error writing request")

		data, err := io.ReadAll(conn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		body := string(data)
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=%v", proxyProtocol))
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-port=1234")
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-proto=http")
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-for=192.168.0.1")
	})

	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should respect proto passed by the PROXY Protocol server port", func() {
		host := proxyProtocol

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
		_, err = conn.Write([]byte(header))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error writing header")

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error writing request")

		data, err := io.ReadAll(conn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		body := string(data)
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=%v", proxyProtocol))
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-port=443")
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-proto=https")
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-for=192.168.0.1")
	})

	ginkgo.It("should enable PROXY Protocol for HTTPS", func() {
		host := proxyProtocol

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

		data, err := io.ReadAll(tlsConn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		body := string(data)
		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("host=%v", proxyProtocol))
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-port=1234")
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-proto=https")
		assert.Contains(ginkgo.GinkgoT(), body, "x-scheme=https")
		assert.Contains(ginkgo.GinkgoT(), body, "x-forwarded-for=192.168.0.1")
	})

	ginkgo.It("should enable PROXY Protocol for TCP", func() {
		if framework.IsCrossplane() {
			return
		}
		cmapData := map[string]string{}
		cmapData[setting] = "true"
		cmapData["enable-real-ip"] = "true"
		f.SetNginxConfigMapData(cmapData)

		config, err := f.KubeClientSet.
			CoreV1().
			ConfigMaps(f.Namespace).
			Get(context.TODO(), "tcp-services", metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining tcp-services configmap")
		assert.NotNil(ginkgo.GinkgoT(), config, "expected a configmap but none returned")

		if config.Data == nil {
			config.Data = map[string]string{}
		}

		config.Data["8080"] = fmt.Sprintf("%v/%v:80:PROXY", f.Namespace, framework.EchoService)

		_, err = f.KubeClientSet.
			CoreV1().
			ConfigMaps(f.Namespace).
			Update(context.TODO(), config, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating configmap")

		svc, err := f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Get(context.TODO(), "nginx-ingress-controller", metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining ingress-nginx service")
		assert.NotNil(ginkgo.GinkgoT(), svc, "expected a service but none returned")

		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       framework.EchoService,
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
		})
		_, err = f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Update(context.TODO(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		// wait for update and nginx reload and new endpoint is available
		framework.Sleep()

		ip := f.GetNginxIP()

		conn, err := net.Dial("tcp", net.JoinHostPort(ip, "8080"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error creating connection to %s:8080", ip)
		defer conn.Close()

		header := "PROXY TCP4 192.168.0.1 192.168.0.11 56324 8080\r\n"
		_, err = conn.Write([]byte(header))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error writing header")

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error writing request")

		_, err = io.ReadAll(conn)
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error reading connection data")

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, `192.168.0.1`)
	})
})
