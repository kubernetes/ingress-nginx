/*
Copyright 2024 The Kubernetes Authors.

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
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("Geoip2Autoreload", func() {
	f := framework.NewDefaultFramework("geoip2-autoreload")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should up and running nginx controller using autoreload flag", func() {
		edition := "GeoLite2-Country"

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--maxmind-edition-ids="+edition)
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		filename := fmt.Sprintf("/etc/ingress-controller/geoip/%s.mmdb", edition)
		exec, err := f.ExecIngressPod(fmt.Sprintf(`sh -c "mkdir -p '%s' && wget -O '%s' '%s' 2>&1"`, filepath.Dir(filename), filename, testdataURL))
		framework.Logf(exec)
		assert.Nil(ginkgo.GinkgoT(), err, fmt.Sprintln("error downloading test geoip2 db", filename))

		f.SetNginxConfigMapData(map[string]string{
			"use-geoip2":                   "true",
			"geoip2-autoreload-in-minutes": "5",
		})

		// Check Configmap Autoreload Patterns
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf("geoip2 %s", filename)) &&
					strings.Contains(cfg, fmt.Sprintf("auto_reload 5m;"))
			},
		)

		// Check if Nginx could up, running and routing with auto_reload configs
		host := "ping.com"
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, host) &&
					strings.Contains(server, "location /")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})
