/*
Copyright 2021 The Kubernetes Authors.

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
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("configmap stream-snippet", func() {
	f := framework.NewDefaultFramework("cm-stream-snippet")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should add value of stream-snippet via config map to nginx config", func() {
		host := "foo.com"
		snippet := `server {listen 8000; proxy_pass 127.0.0.1:80;}`

		f.SetNginxConfigMapData(map[string]string{
			"stream-snippet": snippet,
		})

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		svc, err := f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Get(context.TODO(), "nginx-ingress-controller", metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining ingress-nginx service")
		assert.NotNil(ginkgo.GinkgoT(), svc, "expected a service but none returned")

		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       framework.EchoService,
			Port:       8000,
			TargetPort: intstr.FromInt(8000),
		})

		_, err = f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Update(context.TODO(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, snippet)
			})

		f.HTTPTestClient().
			GET("/healthz").
			WithURL(fmt.Sprintf("http://%v:8000/healthz", f.GetNginxIP())).
			Expect().
			Status(http.StatusOK)
	})
})
