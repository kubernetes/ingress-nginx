/*
Copyright 2020 The Kubernetes Authors.

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
	"strings"

	. "github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Settings - hash size", func() {
	f := framework.NewDefaultFramework("hash-size")

	host := "hash-size"

	BeforeEach(func() {
		f.NewEchoDeployment()
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)
	})
	Context("Check server names hash size", func() {
		It("should set server_names_hash_bucket_size", func() {
			f.UpdateNginxConfigMapData("server-name-hash-bucket-size", "512")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_names_hash_bucket_size 512;`))
			})
		})

		It("should set server_names_hash_max_size", func() {
			f.UpdateNginxConfigMapData("server-name-hash-max-size", "4096")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_names_hash_max_size 4096;`))
			})
		})
	})

	Context("Check proxy header hash size", func() {
		It("should set proxy-headers-hash-bucket-size", func() {
			f.UpdateNginxConfigMapData("proxy-headers-hash-bucket-size", "512")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`proxy_headers_hash_bucket_size 512;`))
			})
		})
		It("should set proxy-headers-hash-max-size", func() {
			f.UpdateNginxConfigMapData("proxy-headers-hash-max-size", "4096")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`proxy_headers_hash_max_size 4096;`))
			})
		})
	})
	Context("Check the variable hash size", func() {
		It("should set variables-hash-bucket-size", func() {
			f.UpdateNginxConfigMapData("variables-hash-bucket-size", "512")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`variables_hash_bucket_size 512;`))
			})
		})
		It("should set variables-hash-max-size", func() {
			f.UpdateNginxConfigMapData("variables-hash-max-size", "512")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`variables_hash_max_size 512;`))
			})
		})
	})

	Context("Check the map hash size", func() {
		It("should set vmap-hash-bucket-size", func() {
			f.UpdateNginxConfigMapData("map-hash-bucket-size", "512")

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`map_hash_bucket_size 512;`))
			})
		})
	})

})
