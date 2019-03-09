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

package settings

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Global use-service-upstream", func() {
	f := framework.NewDefaultFramework("use-service-upstream")

	BeforeEach(func() {
		f.UpdateNginxConfigMapData("use-service-upstream", "true")

		f.NewEchoDeploymentWithReplicas(1)

		framework.UpdateDeployment(f.KubeClientSet, f.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, fmt.Sprintf("--default-backend-service=%v/http-svc", f.Namespace))
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.Namespace).Update(deployment)

				return err
			})
	})

	AfterEach(func() {
	})

	It("should use default backend to access a service without endpoints", func() {
		host := "foo"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, nil)
		ing.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort = intstr.FromInt(81)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name foo")
			})

		_, body, errs := gorequest.New().
			Get(f.GetURL(framework.HTTP)).
			Set("Host", host).
			End()
		Expect(errs).To(BeNil())
		//Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(fmt.Sprintf("host=http-svc.%v", f.Namespace)))
	})
})
