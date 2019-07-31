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

package annotations

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - FastCGI", func() {
	f := framework.NewDefaultFramework("fastcgi")

	BeforeEach(func() {
		f.NewFastCGIHelloServerDeployment()
	})

	It("should use fastcgi_pass in the configuration file", func() {
		host := "fastcgi"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
		}

		ing := framework.NewSingleIngress(host, "/hello", host, f.Namespace, "fastcgi-helloserver", 9000, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("include /etc/nginx/fastcgi_params;")) &&
					Expect(server).Should(ContainSubstring("fastcgi_pass"))
			})
	})

	It("should add fastcgi_index in the configuration file", func() {
		host := "fastcgi-index"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
			"nginx.ingress.kubernetes.io/fastcgi-index":    "index.php",
		}

		ing := framework.NewSingleIngress(host, "/hello", host, f.Namespace, "fastcgi-helloserver", 9000, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("fastcgi_index \"index.php\";"))
			})
	})

	It("should add fastcgi_param in the configuration file", func() {
		configuration := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fastcgi-configmap",
				Namespace: f.Namespace,
			},
			Data: map[string]string{
				"SCRIPT_FILENAME": "/home/www/scripts/php$fastcgi_script_name",
				"REDIRECT_STATUS": "200",
			},
		}

		cm, err := f.EnsureConfigMap(configuration)
		Expect(err).NotTo(HaveOccurred(), "failed to create an the configmap")
		Expect(cm).NotTo(BeNil(), "expected a configmap but none returned")

		host := "fastcgi-params-configmap"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol":         "FCGI",
			"nginx.ingress.kubernetes.io/fastcgi-params-configmap": "fastcgi-configmap",
		}

		ing := framework.NewSingleIngress(host, "/hello", host, f.Namespace, "fastcgi-helloserver", 9000, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("fastcgi_param SCRIPT_FILENAME \"/home/www/scripts/php$fastcgi_script_name\";")) &&
					Expect(server).Should(ContainSubstring("fastcgi_param REDIRECT_STATUS \"200\";"))
			})
	})

	It("should return OK for service with backend protocol FastCGI", func() {
		host := "fastcgi-helloserver"
		path := "/hello"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
		}

		ing := framework.NewSingleIngress(host, path, host, f.Namespace, "fastcgi-helloserver", 9000, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("fastcgi_pass"))
			})

		resp, body, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)+path).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring("Hello world!"))
	})
})
