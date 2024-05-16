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

var _ = framework.IngressNginxDescribe("[Default Backend] custom-error-pages", func() {
	f := framework.NewDefaultFramework("custom-error-pages")

	ginkgo.It("should export /metrics and /debug/vars by default", func() {
		tt := []struct {
			Name   string
			Path   string
			Status int
		}{
			{"request to /metrics should return HTTP 200", "/metrics", http.StatusOK},
			{"request to /debug/vars should return HTTP 200", "/debug/vars", http.StatusOK},
		}

		setupIngressControllerWithCustomErrorPages(f, nil)

		for _, t := range tt {
			ginkgo.By(t.Name)
			f.HTTPTestClient().
				GET(t.Path).
				Expect().
				Status(t.Status)
		}
	})

	ginkgo.It("shouldn't export /metrics and /debug/vars when IS_METRICS_EXPORT is set to false", func() {
		tt := []struct {
			Name   string
			Path   string
			Status int
		}{
			{"request to /metrics should return HTTP 404", "/metrics", http.StatusNotFound},
			{"request to /debug/vars should return HTTP 404", "/debug/vars", http.StatusNotFound},
		}

		setupIngressControllerWithCustomErrorPages(f, map[string]string{
			"IS_METRICS_EXPORT": "false",
		})

		for _, t := range tt {
			ginkgo.By(t.Name)
			f.HTTPTestClient().
				GET(t.Path).
				Expect().
				Status(t.Status)
		}
	})

	ginkgo.It("shouldn't export /metrics and /debug/vars when METRICS_PORT is set to a different port", func() {
		tt := []struct {
			Name   string
			Path   string
			Status int
		}{
			{"request to /metrics should return HTTP 404", "/metrics", http.StatusNotFound},
			{"request to /debug/vars should return HTTP 404", "/debug/vars", http.StatusNotFound},
		}

		setupIngressControllerWithCustomErrorPages(f, map[string]string{
			"IS_METRICS_EXPORT": "true",
			"METRICS_PORT":      "8081",
		})

		for _, t := range tt {
			ginkgo.By(t.Name)
			f.HTTPTestClient().
				GET(t.Path).
				Expect().
				Status(t.Status)
		}
	})
})

func setupIngressControllerWithCustomErrorPages(f *framework.Framework, envVars map[string]string) {
	f.NewCustomErrorPagesDeployment(framework.WithEnvVars(envVars))

	err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
		args := deployment.Spec.Template.Spec.Containers[0].Args
		args = append(args, fmt.Sprintf("--default-backend-service=%v/%v", f.Namespace, framework.CustomErrorPagesService))
		deployment.Spec.Template.Spec.Containers[0].Args = args
		_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return err
	})
	assert.Nil(ginkgo.GinkgoT(), err, "updating deployment")

	f.WaitForNginxServer("_",
		func(server string) bool {
			return strings.Contains(server, `set $proxy_upstream_name "upstream-default-backend"`)
		})
}
