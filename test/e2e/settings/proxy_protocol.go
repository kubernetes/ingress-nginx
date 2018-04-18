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

package setting

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Proxy Protocol", func() {
	f := framework.NewDefaultFramework("proxy-protocol")

	setting := "use-proxy-protocol"

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())

		err = f.UpdateNginxConfigMapData(setting, "false")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should respect port passed by the PROXY Protocol", func() {
		host := "proxy-protocol"

		err := f.UpdateNginxConfigMapData(setting, "true")
		Expect(err).NotTo(HaveOccurred())

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name proxy-protocol") &&
					strings.Contains(server, "listen 80 proxy_protocol")
			})
		Expect(err).NotTo(HaveOccurred())

		ip, err := f.GetNginxIP()
		Expect(err).NotTo(HaveOccurred())
		port, err := f.GetNginxPort("http")
		Expect(err).NotTo(HaveOccurred())

		conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", ip, port))
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		header := "PROXY TCP4 192.168.0.1 192.168.0.11 56324 1234\r\n"
		conn.Write([]byte(header))
		conn.Write([]byte("GET / HTTP/1.1\r\nHost: proxy-protocol\r\n\r\n"))

		data, err := ioutil.ReadAll(conn)
		Expect(err).NotTo(HaveOccurred())
		body := string(data)
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=%v", "proxy-protocol")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-port=80")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-for=192.168.0.1")))
	})
})
