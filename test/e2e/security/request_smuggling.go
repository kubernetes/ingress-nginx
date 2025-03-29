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

package security

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Security] request smuggling", func() {
	f := framework.NewDefaultFramework("request-smuggling")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should not return body content from error_page", func() {
		if framework.IsCrossplane() {
			ginkgo.Skip("Crossplane does not support snippets") // TODO: Re-add this test when we enable admin defined snippets
		}
		host := "foo.bar.com"

		snippet := `
server {
	listen 80;
	server_name notlocalhost;
	location /_hidden/index.html {
	return 200 'This should be hidden!';
	}
}`

		f.UpdateNginxConfigMapData("http-snippet", snippet)

		// TODO: currently using a self hosted HTTPBun instance results in a 499, we
		// should move away from using httpbun.com once we have the httpbun
		// deployment as part of the framework
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, map[string]string{
			"nginx.ingress.kubernetes.io/auth-signin": "https://httpbun.com/bearer/d4bcba7a-0def-4a31-91a7-47e420adf44b",
			"nginx.ingress.kubernetes.io/auth-url":    "https://httpbun.com/basic-auth/user/passwd",
		})
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		out, err := smugglingRequest(host, f.GetNginxIP(), 80)
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining response of request smuggling check")
		assert.NotContains(ginkgo.GinkgoT(), out, "This should be hidden!")
	})
})

func smugglingRequest(host, addr string, port int) (string, error) {
	hostPort := net.JoinHostPort(addr, fmt.Sprintf("%v", port))
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(conn, "GET /echo HTTP/1.1\r\nHost: %v\r\nContent-Length: 56\r\n\r\nGET /_hidden/index.html HTTP/1.1\r\nHost: notlocalhost\r\n\r\n", host)
	if err != nil {
		return "", err
	}

	// wait for /_hidden/index.html response
	framework.Sleep()

	buf := make([]byte, 1024)
	r := bufio.NewReader(conn)
	_, err = r.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
