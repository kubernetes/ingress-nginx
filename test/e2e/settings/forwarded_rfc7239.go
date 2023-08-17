/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	enableForwardedRFC7239         = "enable-forwarded-rfc7239"
	forwardedRFC7239StripIncomming = "forwarded-rfc7239-strip-incomming"
	forwardedRFC7239               = "forwarded-rfc7239"
	forwardedRFC7239By             = "forwarded-rfc7239-by"
	forwardedRFC7239For            = "forwarded-rfc7239-for"
)

var _ = framework.DescribeSetting("Configure Forwarded RFC7239", func() {
	f := framework.NewDefaultFramework("enable-forwarded-rfc7239")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.AfterEach(func() {
	})

	ginkgo.It("should not send forwarded header when disabled", func() {
		host := "forwarded-rfc7239"

		f.UpdateNginxConfigMapData(enableForwardedRFC7239, "false")

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					!strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.NotContains(ginkgo.GinkgoT(), body, "forwarded=")
	})

	ginkgo.It("should trust Forwarded header when striping-incoming is false", func() {
		host := "forwarded-rfc7239"

		config := map[string]string{}
		config[enableForwardedRFC7239] = "true"
		config[forwardedRFC7239StripIncomming] = "false"
		config[forwardedRFC7239] = "for"
		config[forwardedRFC7239For] = "ip"
		f.SetNginxConfigMapData(config)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		serverIP := f.GetNginxPodIP()
		clientIP := getClientIP(serverIP)

		ginkgo.By("ensuring valid headers are passed through correctly")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Forwarded", "for=1.1.1.1;secret=_5ecREy, for=\"[2001:4860:4860::8888]\"").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("forwarded=for=1.1.1.1;secret=_5ecREy, for=\"[2001:4860:4860::8888]\", for=%s", clientIP))

		ginkgo.By("ensuring invalid headers are striped")
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Forwarded", "for=2001:4860:4860::8888"). // invalid header, ipv6 should be bracked and quoted
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("forwarded=for=%s", clientIP))
	})

	ginkgo.It("should contain parameters in order as setting forwarded-rfc7239 specified", func() {
		host := "forwarded-rfc7239"

		config := map[string]string{}
		config[enableForwardedRFC7239] = "true"
		config[forwardedRFC7239StripIncomming] = "false"
		config[forwardedRFC7239] = "for"
		config[forwardedRFC7239For] = "ip"
		config[forwardedRFC7239By] = "ip"
		f.SetNginxConfigMapData(config)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		serverIP := f.GetNginxPodIP()
		clientIP := getClientIP(serverIP)

		ginkgo.By("ensuring singly pass through incoming header when empty parameter list")
		f.UpdateNginxConfigMapData(forwardedRFC7239, "")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Forwarded", "for=1.1.1.1").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "forwarded=for=1.1.1.1")

		ginkgo.By("ensuring any parameter combinations work")
		f.UpdateNginxConfigMapData(forwardedRFC7239, "for,by,proto,host")
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf(`forwarded=for=%s;by="%s:80";proto=http;host=forwarded-rfc7239`, clientIP, serverIP))
	})

	ginkgo.It("should config \"for\" and \"by\" parameters as static obfuscated strings", func() {
		host := "forwarded-rfc7239"

		config := map[string]string{}
		config[enableForwardedRFC7239] = "true"
		config[forwardedRFC7239StripIncomming] = "false"
		config[forwardedRFC7239] = "for,by"
		config[forwardedRFC7239For] = "ip"
		config[forwardedRFC7239By] = "_SERVER1"
		f.SetNginxConfigMapData(config)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		serverIP := f.GetNginxPodIP()
		clientIP := getClientIP(serverIP)

		ginkgo.By("ensuring \"by\" parameter is a static obfuscated string")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf("forwarded=for=%s;by=_SERVER1", clientIP))

		ginkgo.By("ensuring \"for\" parameter is a static obfuscated string")
		config[forwardedRFC7239For] = "_HOST1"
		config[forwardedRFC7239By] = "ip"
		f.SetNginxConfigMapData(config)
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf(`forwarded=for=_HOST1;by="%s:80"`, serverIP))

		ginkgo.By("ensuring invalid static obfuscated strings are ingored")
		config[forwardedRFC7239For] = "_HOST1"
		config[forwardedRFC7239By] = "_%"
		f.SetNginxConfigMapData(config)
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, fmt.Sprintf(`forwarded=for=_HOST1;by="%s:80"`, serverIP))
	})
})

func getClientIP(serverIP string) net.IP {
	conn, err := net.Dial("tcp", serverIP+":80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	return localAddr.IP
}
