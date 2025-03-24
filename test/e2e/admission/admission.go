/*
Copyright 2020 The Kubernetes Authors.

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

package admission

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"

	networking "k8s.io/api/networking/v1"
)

const admissionTestHost = "admission-test"

var _ = framework.IngressNginxDescribeSerial("[Admission] admission controller", func() {
	f := framework.NewDefaultFramework("admission")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("should not allow overlaps of host and paths without canary annotations", func() {
		host := admissionTestHost

		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		secondIngress := framework.NewSingleIngress("second-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), secondIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with the same host and path should return an error")
	})

	ginkgo.It("should allow path overlaps if disable-path-overlap-validation is enabled", func() {
		host := admissionTestHost

		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--disable-path-overlap-validation")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating deployment")
		defer func() {
			err = f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
				args := []string{}

				for _, v := range deployment.Spec.Template.Spec.Containers[0].Args {
					if strings.Contains(v, "--disable-path-overlap-validation") {
						continue
					}

					args = append(args, v)
				}
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
				return err
			})
			assert.Nil(ginkgo.GinkgoT(), err, "restoring deployment")
		}()
		oneIngress := framework.NewSingleIngress("a-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), oneIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		anotherIngress := framework.NewSingleIngress("another-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), anotherIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress with the same host and path should  be allowed when disable-path-overlap-validation is true ")
	})

	ginkgo.It("should allow overlaps of host and paths with canary annotation", func() {
		host := admissionTestHost

		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, nil)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		canaryAnnotations := map[string]string{
			"nginx.ingress.kubernetes.io/canary":           "true",
			"nginx.ingress.kubernetes.io/canary-by-header": "CanaryByHeader",
		}
		secondIngress := framework.NewSingleIngress("second-ingress", "/", host, f.Namespace, framework.SlowEchoService, 80, canaryAnnotations)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), secondIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress with the same host and path should not return an error using a canary annotation")
	})

	ginkgo.It("should block ingress with invalid path", func() {
		host := "invalid-path"

		firstIngress := framework.NewSingleIngress("valid-path", "/mypage", host, f.Namespace, framework.EchoService, 80, nil)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		secondIngress := framework.NewSingleIngress("second-ingress", "/etc/nginx", host, f.Namespace, framework.EchoService, 80, nil)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), secondIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid path should return an error")
	})

	ginkgo.It("should return an error if there is an error validating the ingress definition", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := admissionTestHost

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": "something invalid",
		}
		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid configuration should return an error")
	})

	ginkgo.It("should return an error if there is an invalid value in some annotation", func() {
		host := admissionTestHost

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/connection-proxy-header": "a;}",
		}

		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "}")

		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid annotation value should return an error")
	})

	ginkgo.It("should return an error if there is a forbidden value in some annotation", func() {
		host := admissionTestHost

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/connection-proxy-header": "set_by_lua",
		}

		f.UpdateNginxConfigMapData("annotation-value-word-blocklist", "set_by_lua")

		firstIngress := framework.NewSingleIngress("first-ingress", "/", host, f.Namespace, framework.EchoService, 80, annotations)
		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), firstIngress, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid annotation value should return an error")
	})

	ginkgo.It("should return an error if there is an invalid path and wrong pathType is set", func() {
		host := "path-validation"
		var (
			exactPathType  = networking.PathTypeExact
			prefixPathType = networking.PathTypePrefix
			implSpecific   = networking.PathTypeImplementationSpecific
		)

		f.UpdateNginxConfigMapData("strict-validate-path-type", "true")

		invalidPath := framework.NewSingleIngress("first-ingress", "/foo/bar/[a-z]{3}", host, f.Namespace, framework.EchoService, 80, nil)
		invalidPath.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &exactPathType

		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), invalidPath, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid path value should return an error")

		invalidPath.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &prefixPathType
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), invalidPath, metav1.CreateOptions{})
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid path value should return an error")

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		pathSpecific := framework.NewSingleIngress("pathspec-ingress", "/foo/bar/[a-z]{3}", host, f.Namespace, framework.EchoService, 80, annotations)
		pathSpecific.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].PathType = &implSpecific
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), pathSpecific, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress with arbitrary path and implSpecific value should not return an error")

		validPath := framework.NewSingleIngress("second-ingress", "/bloblo", host, f.Namespace, framework.EchoService, 80, nil)
		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Create(context.TODO(), validPath, metav1.CreateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress with valid path should not return an error")
	})

	ginkgo.It("should not return an error if the Ingress V1 definition is valid with Ingress Class", func() {
		out, err := createIngress(f.Namespace, validV1Ingress)
		assert.Equal(ginkgo.GinkgoT(), "ingress.networking.k8s.io/extensions created\n", out)
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress using kubectl")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "extensions")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", "extensions").
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should not return an error if the Ingress V1 definition is valid with IngressClass annotation", func() {
		out, err := createIngress(f.Namespace, validV1IngressAnnotation)
		assert.Equal(ginkgo.GinkgoT(), "ingress.networking.k8s.io/extensions-class created\n", out)
		assert.Nil(ginkgo.GinkgoT(), err, "creating an ingress using kubectl")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "extensions-class")
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", "extensions-class").
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return an error if the Ingress V1 definition contains invalid annotations", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		out, err := createIngress(f.Namespace, invalidV1Ingress)
		assert.Empty(ginkgo.GinkgoT(), out)
		assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress using kubectl")

		_, err = f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Get(context.TODO(), "extensions-invalid", metav1.GetOptions{})
		if !apierrors.IsNotFound(err) {
			assert.NotNil(ginkgo.GinkgoT(), err, "creating an ingress with invalid configuration should return an error")
		}
	})

	ginkgo.It("should not return an error for an invalid Ingress when it has unknown class", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()
		out, err := createIngress(f.Namespace, invalidV1IngressWithOtherClass)
		assert.Equal(ginkgo.GinkgoT(), "ingress.networking.k8s.io/extensions-invalid-other created\n", out)
		assert.Nil(ginkgo.GinkgoT(), err, "creating an invalid ingress with unknown class using kubectl")
	})
})

const (
	validV1Ingress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: extensions
spec:
  ingressClassName: nginx
  rules:
  - host: extensions
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: echo
            port:
              number: 80

---
`
	validV1IngressAnnotation = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: extensions-class
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: extensions-class
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: echo
            port:
              number: 80

---
`

	invalidV1Ingress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: extensions-invalid
  annotations:
    nginx.ingress.kubernetes.io/configuration-snippet: |
      invalid directive
spec:
  ingressClassName: nginx
  rules:
  - host: extensions-invalid
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: echo
            port:
              number: 80
---
`
	invalidV1IngressWithOtherClass = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: extensions-invalid-other
  annotations:
    nginx.ingress.kubernetes.io/configuration-snippet: |
      invalid directive
spec:
  ingressClassName: nginx-other
  rules:
  - host: extensions-invalid
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: echo
            port:
              number: 80
---
`
)

func createIngress(namespace, ingressDefinition string) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)
	//nolint:gosec // Ignore G204 error
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v --warnings-as-errors=false apply --namespace %s -f -", framework.KubectlPath, namespace))
	cmd.Stdin = strings.NewReader(ingressDefinition)
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		stderr := strings.TrimSpace(execErr.String())
		return "", fmt.Errorf("kubectl error: %v\n%v", err, stderr)
	}

	return execOut.String(), nil
}
