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
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[SSL] [Flag] default-ssl-certificate", func() {
	f := framework.NewDefaultFramework("default-ssl-certificate")
	var tlsConfig *tls.Config
	secretName := "my-custom-cert"
	service := framework.EchoService
	port := 80

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(1))

		var err error
		tlsConfig, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			[]string{"*"},
			secretName,
			f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		err = f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--default-ssl-certificate=$(POD_NAMESPACE)/"+secretName)
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})

			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		// this asserts that it configures default custom ssl certificate without an ingress at all
		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)
	})

	ginkgo.It("uses default ssl certificate for catch-all ingress", func() {
		ing := framework.NewSingleCatchAllIngress("catch-all", f.Namespace, service, port, nil)
		f.EnsureIngress(ing)

		ginkgo.By("making sure new ingress is deployed")
		expectedConfig := fmt.Sprintf(`set $proxy_upstream_name "%v-%v-%v";`, f.Namespace, service, port)
		f.WaitForNginxServer("_", func(cfg string) bool {
			return strings.Contains(cfg, expectedConfig)
		})

		ginkgo.By("making sure new ingress is responding")

		ginkgo.By("making sure the configured default ssl certificate is being used")
		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)
	})

	ginkgo.It("uses default ssl certificate for host based ingress when configured certificate does not match host", func() {
		host := "foo"

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, service, port, nil))
		_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			[]string{"not.foo"},
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		ginkgo.By("making sure new ingress is deployed")
		expectedConfig := fmt.Sprintf(`set $proxy_upstream_name "%v-%v-%v";`, f.Namespace, service, port)
		f.WaitForNginxServer(host, func(cfg string) bool {
			return strings.Contains(cfg, expectedConfig)
		})

		ginkgo.By("making sure the configured default ssl certificate is being used")
		framework.WaitForTLS(f.GetURL(framework.HTTPS), tlsConfig)
	})
})
