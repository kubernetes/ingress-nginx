/*
Copyright 2018 The Kubernetes Authors.

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

package annotations

import (
	"strings"

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("client-body-buffer-size", func() {
	f := framework.NewDefaultFramework("clientbodybuffersize")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should set client_body_buffer_size to 1000", func() {
		host := "proxy.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/client-body-buffer-size": "1000",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "client_body_buffer_size 1000;")
			})
	})

	ginkgo.It("should set client_body_buffer_size to 1K", func() {
		host := "proxy.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/client-body-buffer-size": "1K",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "client_body_buffer_size 1K;")
			})
	})

	ginkgo.It("should set client_body_buffer_size to 1k", func() {
		host := "proxy.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/client-body-buffer-size": "1k",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "client_body_buffer_size 1k;")
			})
	})

	ginkgo.It("should set client_body_buffer_size to 1m", func() {
		host := "proxy.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/client-body-buffer-size": "1m",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "client_body_buffer_size 1m;")
			})
	})

	ginkgo.It("should set client_body_buffer_size to 1M", func() {
		host := "proxy.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/client-body-buffer-size": "1M",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "client_body_buffer_size 1M;")
			})
	})

	ginkgo.It("should not set client_body_buffer_size to invalid 1b", func() {
		host := "proxy.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/client-body-buffer-size": "1b",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "client_body_buffer_size 1b;")
			})
	})
})
