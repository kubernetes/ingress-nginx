/*
Copyright 2017 The Kubernetes Authors.

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

package framework

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	api "k8s.io/api/core/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// EnsureSecret creates a Secret object or returns it if it already exists.
func (f *Framework) EnsureSecret(secret *api.Secret) *api.Secret {
	s, err := f.KubeClientSet.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			s, err := f.KubeClientSet.CoreV1().Secrets(secret.Namespace).Update(secret)
			Expect(err).NotTo(HaveOccurred(), "unexpected error updating secret")

			return s
		}

		Expect(err).NotTo(HaveOccurred(), "unexpected error creating secret")
	}

	Expect(s).NotTo(BeNil())
	Expect(s.ObjectMeta).NotTo(BeNil())

	return s
}

// EnsureConfigMap creates a ConfigMap object or returns it if it already exists.
func (f *Framework) EnsureConfigMap(configMap *api.ConfigMap) (*api.ConfigMap, error) {
	cm, err := f.KubeClientSet.CoreV1().ConfigMaps(configMap.Namespace).Create(configMap)
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return f.KubeClientSet.CoreV1().ConfigMaps(configMap.Namespace).Update(configMap)
		}
		return nil, err
	}

	return cm, nil
}

// EnsureIngress creates an Ingress object or returns it if it already exists.
func (f *Framework) EnsureIngress(ingress *extensions.Ingress) *extensions.Ingress {
	ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(ingress)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			ing, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Create(ingress)
			Expect(err).NotTo(HaveOccurred(), "unexpected error creating ingress")
			return ing
		}

		Expect(err).NotTo(HaveOccurred())
	}

	Expect(ing).NotTo(BeNil())

	if ing.Annotations == nil {
		ing.Annotations = make(map[string]string)
	}

	return ing
}

// EnsureService creates a Service object or returns it if it already exists.
func (f *Framework) EnsureService(service *core.Service) *core.Service {
	s, err := f.KubeClientSet.CoreV1().Services(service.Namespace).Update(service)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			s, err := f.KubeClientSet.CoreV1().Services(service.Namespace).Create(service)
			Expect(err).NotTo(HaveOccurred(), "unexpected error creating service")
			return s

		}

		Expect(err).NotTo(HaveOccurred())
	}

	Expect(s).NotTo(BeNil(), "expected a service but none returned")

	return s
}

// EnsureDeployment creates a Deployment object or returns it if it already exists.
func (f *Framework) EnsureDeployment(deployment *extensions.Deployment) (*extensions.Deployment, error) {
	d, err := f.KubeClientSet.Extensions().Deployments(deployment.Namespace).Update(deployment)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return f.KubeClientSet.Extensions().Deployments(deployment.Namespace).Create(deployment)
		}
		return nil, err
	}
	return d, nil
}

// WaitForPodsReady waits for a given amount of time until a group of Pods is running in the given namespace.
func WaitForPodsReady(kubeClientSet kubernetes.Interface, timeout time.Duration, expectedReplicas int, namespace string, opts metav1.ListOptions) error {
	return wait.Poll(2*time.Second, timeout, func() (bool, error) {
		pl, err := kubeClientSet.CoreV1().Pods(namespace).List(opts)
		if err != nil {
			return false, err
		}

		r := 0
		for _, p := range pl.Items {
			if isRunning, _ := podRunningReady(&p); isRunning {
				r++
			}
		}

		if r == expectedReplicas {
			return true, nil
		}

		return false, nil
	})
}

// WaitForEndpoints waits for a given amount of time until an endpoint contains.
func WaitForEndpoints(kubeClientSet kubernetes.Interface, timeout time.Duration, name, ns string, expectedEndpoints int) error {
	if expectedEndpoints == 0 {
		return nil
	}
	return wait.Poll(2*time.Second, timeout, func() (bool, error) {
		endpoint, err := kubeClientSet.CoreV1().Endpoints(ns).Get(name, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) {
			return false, err
		}
		Expect(err).NotTo(HaveOccurred())
		if len(endpoint.Subsets) == 0 || len(endpoint.Subsets[0].Addresses) == 0 {
			return false, err
		}

		r := 0
		for _, es := range endpoint.Subsets {
			r += len(es.Addresses)
		}

		if r == expectedEndpoints {
			return true, nil
		}

		return false, nil
	})
}

// podRunningReady checks whether pod p's phase is running and it has a ready
// condition of status true.
func podRunningReady(p *core.Pod) (bool, error) {
	// Check the phase is running.
	if p.Status.Phase != core.PodRunning {
		return false, fmt.Errorf("want pod '%s' on '%s' to be '%v' but was '%v'",
			p.ObjectMeta.Name, p.Spec.NodeName, core.PodRunning, p.Status.Phase)
	}
	// Check the ready condition is true.
	if !podutil.IsPodReady(p) {
		return false, fmt.Errorf("pod '%s' on '%s' didn't have condition {%v %v}; conditions: %v",
			p.ObjectMeta.Name, p.Spec.NodeName, core.PodReady, core.ConditionTrue, p.Status.Conditions)
	}
	return true, nil
}

func getIngressNGINXPod(ns string, kubeClientSet kubernetes.Interface) (*core.Pod, error) {
	l, err := kubeClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ingress-nginx",
	})
	if err != nil {
		return nil, err
	}

	if len(l.Items) == 0 {
		return nil, fmt.Errorf("There is no ingress-nginx pods running in namespace %v", ns)
	}

	var pod *core.Pod

	for _, p := range l.Items {
		if strings.HasPrefix(p.GetName(), "nginx-ingress-controller") {
			if isRunning, err := podRunningReady(&p); err == nil && isRunning {
				pod = &p
				break
			}
		}
	}

	if pod == nil {
		return nil, fmt.Errorf("There is no ingress-nginx pods running in namespace %v", ns)
	}

	return pod, nil
}
