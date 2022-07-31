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
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"net/http"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

const testdataURL = "https://github.com/maxmind/MaxMind-DB/blob/5a0be1c0320490b8e4379dbd5295a18a648ff156/test-data/GeoLite2-Country-Test.mmdb?raw=true"

var _ = framework.DescribeSetting("Geoip2", func() {
	f := framework.NewDefaultFramework("geoip2")

	host := "geoip2"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should include geoip2 line in config when enabled and db file exists", func() {
		edition := "GeoLite2-Country"

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--maxmind-edition-ids="+edition)
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		filename := fmt.Sprintf("/etc/nginx/geoip/%s.mmdb", edition)
		exec, err := f.ExecIngressPod(fmt.Sprintf(`sh -c "mkdir -p '%s' && wget -O '%s' '%s' 2>&1"`, filepath.Dir(filename), filename, testdataURL))
		framework.Logf(exec)
		assert.Nil(ginkgo.GinkgoT(), err, fmt.Sprintln("error downloading test geoip2 db", filename))

		f.UpdateNginxConfigMapData("use-geoip2", "true")
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf("geoip2 %s", filename))
			})
	})

	ginkgo.It("should only allow requests from specific countries", func() {
		ginkgo.Skip("GeoIP test are temporarily disabled")

		f.UpdateNginxConfigMapData("use-geoip2", "true")

		httpSnippetAllowingOnlyAustralia :=
			`map $geoip2_city_country_code $blocked_country {
  default 1;
  AU 0;
}`
		f.UpdateNginxConfigMapData("http-snippet", httpSnippetAllowingOnlyAustralia)
		f.UpdateNginxConfigMapData("use-forwarded-headers", "true")

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "map $geoip2_city_country_code $blocked_country")
			})

		configSnippet :=
			`if ($blocked_country) {
  return 403;
}`

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": configSnippet,
		}

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "if ($blocked_country)")
			})

		// Should be blocked
		usIP := "8.8.8.8"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", usIP).
			Expect().
			Status(http.StatusForbidden)

		// Shouldn't be blocked
		australianIP := "1.1.1.1"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", australianIP).
			Expect().
			Status(http.StatusOK)
	})
})
