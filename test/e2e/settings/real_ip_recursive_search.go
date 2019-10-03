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

package settings

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Real IP Recursive Search", func() {
	f := framework.NewDefaultFramework("real-ip-recursive-search")

	setting := "real-ip-recursive-search"

	BeforeEach(func() {
		f.NewEchoDeployment()
		f.UpdateNginxConfigMapData("use-forwarded-headers", "true")
		f.UpdateNginxConfigMapData(setting, "false")
	})

	AfterEach(func() {
	})

	It("should include real_ip_recursive=off if turned off", func() {
		host := "real-ip-recursive-search"
		expectedConfig := "real_ip_recursive off;"

		f.UpdateNginxConfigMapData(setting, "false")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedConfig)
			})

		By("ensuring x-forwarded-for values are parsed correctly")
		resp, body, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Set("Host", host).
			Set("X-Forwarded-For", "10.0.0.1,1.2.3.4").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%s", host)))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-for=1.2.3.4")))
	})

	It("should include real_ip_recursive=on if turned on", func() {
		host := "real-ip-recursive-search"
		expectedConfig := "real_ip_recursive on;"

		f.UpdateNginxConfigMapData(setting, "true")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedConfig)
			})

		By("ensuring x-forwarded-for values are parsed correctly")
		resp, body, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Set("Host", host).
			Set("X-Forwarded-For", "10.0.0.1,1.2.3.4").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%s", host)))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-for=10.0.0.1")))
	})

	It("should include real_ip_recursive=on if proxy-protocol is on", func() {
		host := "real-ip-recursive-search"
		expectedConfig := "real_ip_recursive on;"

		f.UpdateNginxConfigMapData("proxy-protocol", "true")

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, expectedConfig)
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name real-ip-recursive-search") &&
					strings.Contains(server, "listen 80 proxy_protocol")
			})

		ip := f.GetNginxIP()

		conn, err := net.Dial("tcp", net.JoinHostPort(ip, "80"))
		Expect(err).NotTo(HaveOccurred(), "unexpected error creating connection to %s:80", ip)
		defer conn.Close()

		header := "PROXY TCP4 192.168.0.1 192.168.0.11 56324 1234\r\n"
		conn.Write([]byte(header))
		conn.Write([]byte("GET / HTTP/1.1\r\nHost: real-ip-recursive-search\r\nX-Forwarded-For: 192.168.0.1,10.0.0.1\r\n\r\n"))

		data, err := ioutil.ReadAll(conn)
		Expect(err).NotTo(HaveOccurred(), "unexpected error reading connection data")
		body := string(data)
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%s", host)))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-port=80")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-for=192.168.0.1")))
	})
})
