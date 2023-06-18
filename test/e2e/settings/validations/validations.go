/*
Copyright 2023 The Kubernetes Authors.

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
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

func buildSecret(username, password, name, namespace string) *corev1.Secret {
	//out, err := exec.Command("openssl", "passwd", "-crypt", password).CombinedOutput()
	out, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	encpass := fmt.Sprintf("%v:%s\n", username, out)
	assert.Nil(ginkgo.GinkgoT(), err)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       name,
			Namespace:                  namespace,
			DeletionGracePeriodSeconds: framework.NewInt64(1),
		},
		Data: map[string][]byte{
			"auth": []byte(encpass),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

var _ = framework.DescribeAnnotation("annotation validations", func() {
	f := framework.NewDefaultFramework("annotations-validations")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		otherns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "otherns",
			},
		}
		_, err := f.KubeClientSet.CoreV1().Namespaces().Create(context.Background(), otherns, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating namespace")
	})

	ginkgo.AfterEach(func() {
		err := f.KubeClientSet.CoreV1().Namespaces().Delete(context.Background(), "otherns", metav1.DeleteOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "deleting namespace")
	})

	ginkgo.It("should return status code 401 when authentication is configured but Authorization header is not configured", func() {
		host := "annotation-validations"
		// Allow cross namespace consumption
		f.UpdateNginxConfigMapData("allow-cross-namespace-resources", "true")
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		s := f.EnsureSecret(buildSecret("foo", "bar", "test", "otherns"))

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": fmt.Sprintf("%s/%s", s.Namespace, s.Name),
			"nginx.ingress.kubernetes.io/auth-realm":  "test auth",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name annotation-validations")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusUnauthorized).
			Body().Contains("401 Authorization Required")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithBasicAuth("foo", "bar").
			Expect().
			Status(http.StatusOK)
	})
})
