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
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.NotContains(ginkgo.GinkgoT(), body, "forwarded=1.2.3.4")
	})

	ginkgo.It("should trust Forwarded header when striping-incoming is false", func() {
		host := "forwarded-rfc7239"

		f.UpdateNginxConfigMapData(enableForwardedRFC7239, "true")
		f.UpdateNginxConfigMapData(forwardedRFC7239StripIncomming, "false")
		f.UpdateNginxConfigMapData(forwardedRFC7239For, "ip")

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		ginkgo.By("ensuring valid headers are passed through correctly")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Forwarded", "for=1.1.1.1;secret=_5ecREy").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "for=1.1.1.1;secret=_5e(REy, for=1.2.3.4")

		ginkgo.By("ensuring invalid headers are striped")
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Forwarded", "for=1.1.1.1;secret=:x:"). // colon should be quoted
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "for=1.2.3.4")
	})

	ginkgo.It("should contain parameters in order as setting forwarded-rfc7239 specified", func() {
		host := "forwarded-rfc7239"

		f.UpdateNginxConfigMapData(enableForwardedRFC7239, "true")
		f.UpdateNginxConfigMapData(forwardedRFC7239StripIncomming, "false")
		f.UpdateNginxConfigMapData(forwardedRFC7239For, "ip")
		f.UpdateNginxConfigMapData(forwardedRFC7239By, "ip")

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		ginkgo.By("ensuring singly pass through incoming header when empty parameter list")
		f.UpdateNginxConfigMapData(forwardedRFC7239, "")
		body := f.HTTPTestClient().
			GET("/").
			WithHeader("Forwarded", "for=1.1.1.1").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "for=1.1.1.1")

		ginkgo.By("ensuring any parameter combinations work")
		f.UpdateNginxConfigMapData(forwardedRFC7239, "for,by,proto,host")
		body = f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, `for=1.2.3.4;by="1.2.3.4:80";proto=http;host=forwarded-rfc7239`)
	})

	ginkgo.It("should config for and by parameters as constants", func() {
		host := "forwarded-rfc7239"

		f.UpdateNginxConfigMapData(enableForwardedRFC7239, "true")
		f.UpdateNginxConfigMapData(forwardedRFC7239, "for,by")
		f.UpdateNginxConfigMapData(forwardedRFC7239For, "ip")
		f.UpdateNginxConfigMapData(forwardedRFC7239By, "host2")

		f.UpdateNginxConfigMapData(forwardedRFC7239, "")

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name forwarded-rfc7239") &&
					strings.Contains(server, "proxy_set_header Forwarded $proxy_add_forwarded_rfc2379;")
			})

		ginkgo.By("ensuring by parameter is specified value")
		body := f.HTTPTestClient().
			GET("/").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, "for=1.2.3.4;by=host2")

		ginkgo.By("ensuring for parameter is specified value")
		f.UpdateNginxConfigMapData(forwardedRFC7239For, "a@cY")
		f.UpdateNginxConfigMapData(forwardedRFC7239By, "ip")
		body = f.HTTPTestClient().
			GET("/").
			Expect().
			Status(http.StatusOK).
			Body().
			Raw()

		assert.Contains(ginkgo.GinkgoT(), body, `for="a@cY";by="1.2.3.4:80"`)
	})
})
