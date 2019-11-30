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

package settings

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("ssl-passthrough", func() {
	f := framework.NewDefaultFramework("ssl-passthrough")

	BeforeEach(func() {
		framework.UpdateDeployment(f.KubeClientSet, f.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, "--enable-ssl-passthrough")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(deployment)

				return err
			})
	})

	It("Non ssl-passthrough backend still work", func() {
		f.NewEchoDeployment()

		host := "foo.bar.com"

		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring(host))
			})

		resp, _, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Retry(10, 1*time.Second, http.StatusNotFound).
			RedirectPolicy(noRedirectPolicyFunc).
			Set("Host", host).
			End()

		Expect(errs).Should(BeEmpty())
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	})

	// It("connect to ssl-passthrough backend", func() {
	// })
})
