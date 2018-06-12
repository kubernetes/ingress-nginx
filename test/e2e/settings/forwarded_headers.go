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
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("X-Forwarded headers", func() {
	f := framework.NewDefaultFramework("forwarded-headers")

	setting := "use-forwarded-headers"

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())

		err = f.UpdateNginxConfigMapData(setting, "false")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should trust X-Forwarded headers when setting is true", func() {
		host := "forwarded-headers"

		err := f.UpdateNginxConfigMapData(setting, "true")
		Expect(err).NotTo(HaveOccurred())

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-headers")
			})
		Expect(err).NotTo(HaveOccurred())

		ip, err := f.GetNginxIP()
		Expect(err).NotTo(HaveOccurred())
		port, err := f.GetNginxPort("http")
		Expect(err).NotTo(HaveOccurred())

		conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", ip, port))
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		conn.Write([]byte("GET / HTTP/1.1\r\nHost: forwarded-headers\r\nX-Forwarded-Port: 1234\r\nX-Forwarded-Proto: myproto\r\nX-Forwarded-For: 1.2.3.4\r\nX-Forwarded-Host: myhost\r\n\r\n"))

		data, err := ioutil.ReadAll(conn)
		Expect(err).NotTo(HaveOccurred())
		body := string(data)
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=myhost")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-host=myhost")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-proto=myproto")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-port=1234")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-for=1.2.3.4")))
	})
	It("should not trust X-Forwarded headers when setting is false", func() {
		host := "forwarded-headers"

		err := f.UpdateNginxConfigMapData(setting, "false")
		Expect(err).NotTo(HaveOccurred())

		ing, err := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-headers")
			})
		Expect(err).NotTo(HaveOccurred())

		ip, err := f.GetNginxIP()
		Expect(err).NotTo(HaveOccurred())
		port, err := f.GetNginxPort("http")
		Expect(err).NotTo(HaveOccurred())

		conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", ip, port))
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		conn.Write([]byte("GET / HTTP/1.1\r\nHost: forwarded-headers\r\nX-Forwarded-Port: 1234\r\nX-Forwarded-Proto: myproto\r\nX-Forwarded-For: 1.2.3.4\r\nX-Forwarded-Host: myhost\r\n\r\n"))

		data, err := ioutil.ReadAll(conn)
		Expect(err).NotTo(HaveOccurred())
		body := string(data)
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=forwarded-headers")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-port=80")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-forwarded-proto=http")))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("x-original-forwarded-for=1.2.3.4")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("host=myhost")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-host=myhost")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-proto=myproto")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-port=1234")))
		Expect(body).ShouldNot(ContainSubstring(fmt.Sprintf("x-forwarded-for=1.2.3.4")))
	})
})
