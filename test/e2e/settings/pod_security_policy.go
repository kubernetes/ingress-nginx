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
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	ingressControllerPSP = "ingress-controller-psp"
)

var _ = framework.IngressNginxDescribe("[Serial] Pod Security Policies", func() {
	f := framework.NewDefaultFramework("pod-security-policies")

	BeforeEach(func() {
		psp := createPodSecurityPolicy()
		_, err := f.KubeClientSet.Extensions().PodSecurityPolicies().Create(psp)
		if !k8sErrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred(), "creating Pod Security Policy")
		}

		role, err := f.KubeClientSet.RbacV1().ClusterRoles().Get("nginx-ingress-clusterrole", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "getting ingress controller cluster role")
		Expect(role).NotTo(BeNil())

		role.Rules = append(role.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"policy"},
			Resources:     []string{"podsecuritypolicies"},
			ResourceNames: []string{ingressControllerPSP},
			Verbs:         []string{"use"},
		})

		_, err = f.KubeClientSet.RbacV1().ClusterRoles().Update(role)
		Expect(err).NotTo(HaveOccurred(), "updating ingress controller cluster role to use a pod security policy")

		// update the deployment just to trigger a rolling update and the use of the security policy
		err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, "--v=2")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.IngressController.Namespace).Update(deployment)

				return err
			})
		Expect(err).NotTo(HaveOccurred())

		f.NewEchoDeployment()
	})

	AfterEach(func() {
		role, err := f.KubeClientSet.RbacV1().ClusterRoles().Get("nginx-ingress-clusterrole", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "getting ingress controller cluster role")
		Expect(role).NotTo(BeNil())

		index := -1
		for idx, rule := range role.Rules {
			found := false
			for _, rn := range rule.ResourceNames {
				if rn == ingressControllerPSP {
					found = true
					break
				}
			}
			if found {
				index = idx
			}
		}

		role.Rules = append(role.Rules[:index], role.Rules[index+1:]...)
		_, err = f.KubeClientSet.RbacV1().ClusterRoles().Update(role)
		Expect(err).NotTo(HaveOccurred(), "updating ingress controller cluster role to not use a pod security policy")
	})

	It("should be running with a Pod Security Policy", func() {
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "server_tokens on")
			})

		resp, _, _ := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", "foo.bar.com").
			End()
		Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
	})
})

func createPodSecurityPolicy() *extensions.PodSecurityPolicy {
	trueValue := true
	return &extensions.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingressControllerPSP,
		},
		Spec: extensions.PodSecurityPolicySpec{
			AllowPrivilegeEscalation: &trueValue,
			RequiredDropCapabilities: []corev1.Capability{"All"},
			RunAsUser: extensions.RunAsUserStrategyOptions{
				Rule: "RunAsAny",
			},
			SELinux: extensions.SELinuxStrategyOptions{
				Rule: "RunAsAny",
			},
			FSGroup: extensions.FSGroupStrategyOptions{
				Ranges: []extensions.IDRange{
					{
						Min: 1,
						Max: 65535,
					},
				},
				Rule: "MustRunAs",
			},
			SupplementalGroups: extensions.SupplementalGroupsStrategyOptions{
				Ranges: []extensions.IDRange{
					{
						Min: 1,
						Max: 65535,
					},
				},
				Rule: "MustRunAs",
			},
		},
	}

}
