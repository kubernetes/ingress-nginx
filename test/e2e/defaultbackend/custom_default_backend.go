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

package defaultbackend

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Default Backend] custom service", func() {
	f := framework.NewDefaultFramework("custom-default-backend")

	ginkgo.It("uses custom default backend that returns 200 as status code", func() {
		f.NewEchoDeployment()

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, fmt.Sprintf("--default-backend-service=%v/%v", f.Namespace, framework.EchoService))
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating deployment")

		f.WaitForNginxServer("_",
			func(server string) bool {
				return strings.Contains(server, `set $proxy_upstream_name "upstream-default-backend"`) ||
					strings.Contains(server, `set $proxy_upstream_name upstream-default-backend`)
			})

		f.HTTPTestClient().
			GET("/").
			Expect().
			Status(http.StatusOK)
	})
})
