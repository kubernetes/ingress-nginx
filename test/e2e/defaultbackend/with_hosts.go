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

package defaultbackend

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"strings"

	"github.com/parnurzeal/gorequest"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Default backend with hosts", func() {
	f := framework.NewDefaultFramework("default-backend-hosts")
	host := "foo.com"

	BeforeEach(func() {
		f.NewEchoDeploymentWithReplicas(1)
	})

	AfterEach(func() {
	})

	It("should apply the annotation to the default backend", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-buffer-size": "8k",
		}

		ing := &extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "default-backend-annotations",
				Namespace:   f.Namespace,
				Annotations: annotations,
			},
			Spec: extensions.IngressSpec{
				Backend: &extensions.IngressBackend{
					ServiceName: "http-svc",
					ServicePort: intstr.FromInt(8080),
				},
				Rules: []extensions.IngressRule{
					{
						Host: host,
					},
				},
			},
		}

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_buffer_size 8k;")
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Set("Host", "foo.com").
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})

})
