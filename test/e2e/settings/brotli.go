/*
Copyright 2021 The Kubernetes Authors.

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
	"net/http"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("brotli", func() {
	f := framework.NewDefaultFramework("brotli")

	host := "brotli"

	ginkgo.BeforeEach(func() {
		f.NewHttpbinDeployment()
	})

	ginkgo.It("should only compress responses that meet the `brotli-min-length` condition", func() {
		brotliMinLength := 24
		contentEncoding := "application/octet-stream"
		f.UpdateNginxConfigMapData("enable-brotli", "true")
		f.UpdateNginxConfigMapData("brotli-types", contentEncoding)
		f.UpdateNginxConfigMapData("brotli-min-length", strconv.Itoa(brotliMinLength))

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBinService, 80, nil))

		f.WaitForNginxConfiguration(
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host)) &&
					strings.Contains(server, "brotli on") &&
					strings.Contains(server, fmt.Sprintf("brotli_types %v", contentEncoding)) &&
					strings.Contains(server, fmt.Sprintf("brotli_min_length %d", brotliMinLength))
			})

		f.HTTPTestClient().
			GET(fmt.Sprintf("/bytes/%d", brotliMinLength)).
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "br").
			Expect().
			Status(http.StatusOK).
			ContentType(contentEncoding).
			ContentEncoding("br")

		f.HTTPTestClient().
			GET(fmt.Sprintf("/bytes/%d", brotliMinLength-1)).
			WithHeader("Host", host).
			WithHeader("Accept-Encoding", "br").
			Expect().
			Status(http.StatusOK).
			ContentType(contentEncoding).
			ContentEncoding()
	})
})
