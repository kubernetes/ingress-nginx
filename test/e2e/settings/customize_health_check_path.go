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
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Customize health check path", func() {
	f := framework.NewDefaultFramework("custom-health-check-path")

	Context("with a plain HTTP ingress", func() {
		It("should return HTTP/1.1 200 OK on custom health check path and port", func() {

			f.WaitForNginxConfiguration(func(server string) bool {
				return strings.Contains(server, "location /not-healthz")
			})

			err := framework.WaitForPodsReady(f.KubeClientSet, framework.DefaultTimeout, 1, f.Namespace, metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/name=ingress-nginx",
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
