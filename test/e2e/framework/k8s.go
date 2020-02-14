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

	appsv1 "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// EnsureSecret creates a Secret object or returns it if it already exists.
func (f *Framework) EnsureSecret(secret *api.Secret) *api.Secret {
	err := createSecretWithRetries(f.KubeClientSet, f.Namespace, secret)
	Expect(err).To(BeNil(), "unexpected error creating secret")

	s, err := f.KubeClientSet.CoreV1().Secrets(secret.Namespace).Get(secret.Name, metav1.GetOptions{})
	Expect(s).NotTo(BeNil())
	Expect(s.ObjectMeta).NotTo(BeNil())

	return s
}

// EnsureConfigMap creates a ConfigMap object or returns it if it already exists.
func (f *Framework) EnsureConfigMap(configMap *api.ConfigMap) (*api.ConfigMap, error) {
	cm, err := f.KubeClientSet.CoreV1().ConfigMaps(f.Namespace).Create(configMap)
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return f.KubeClientSet.CoreV1().ConfigMaps(f.Namespace).Update(configMap)
		}
		return nil, err
	}

	return cm, nil
}

// EnsureIngress creates an Ingress object or returns it if it already exists.
func (f *Framework) EnsureIngress(ingress *networking.Ingress) *networking.Ingress {
	ing, err := f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Create(ingress)
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				var err error
				ing, err = f.KubeClientSet.NetworkingV1beta1().Ingresses(f.Namespace).Update(ingress)
				if err != nil {
					return err
				}

				return nil
			})

			Expect(err).NotTo(HaveOccurred())
		}
	}

	Expect(ing).NotTo(BeNil(), "expected an ingress but none returned")

	if ing.Annotations == nil {
		ing.Annotations = make(map[string]string)
	}

	time.Sleep(5 * time.Second)

	return ing
}

// EnsureService creates a Service object or returns it if it already exists.
func (f *Framework) EnsureService(service *core.Service) *core.Service {
	err := createServiceWithRetries(f.KubeClientSet, f.Namespace, service)
	Expect(err).To(BeNil(), "unexpected error creating service")

	s, err := f.KubeClientSet.CoreV1().Services(f.Namespace).Get(service.Name, metav1.GetOptions{})
	Expect(err).To(BeNil(), "unexpected error searching service")
	Expect(s).NotTo(BeNil())
	Expect(s.ObjectMeta).NotTo(BeNil())

	return s
}

// EnsureDeployment creates a Deployment object or returns it if it already exists.
func (f *Framework) EnsureDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	err := createDeploymentWithRetries(f.KubeClientSet, f.Namespace, deployment)
	Expect(err).To(BeNil(), "unexpected error creating deployment")

	s, err := f.KubeClientSet.AppsV1().Deployments(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{})
	Expect(err).To(BeNil(), "unexpected error searching deployment")
	Expect(s).NotTo(BeNil())
	Expect(s.ObjectMeta).NotTo(BeNil())

	return s
}

// WaitForPodsReady waits for a given amount of time until a group of Pods is running in the given namespace.
func WaitForPodsReady(kubeClientSet kubernetes.Interface, timeout time.Duration, expectedReplicas int, namespace string, opts metav1.ListOptions) error {
	return wait.Poll(Poll, timeout, func() (bool, error) {
		pl, err := kubeClientSet.CoreV1().Pods(namespace).List(opts)
		if err != nil {
			return false, nil
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

// WaitForPodsDeleted waits for a given amount of time until a group of Pods are deleted in the given namespace.
func WaitForPodsDeleted(kubeClientSet kubernetes.Interface, timeout time.Duration, namespace string, opts metav1.ListOptions) error {
	return wait.Poll(Poll, timeout, func() (bool, error) {
		pl, err := kubeClientSet.CoreV1().Pods(namespace).List(opts)
		if err != nil {
			return false, nil
		}

		if len(pl.Items) == 0 {
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

	return wait.Poll(Poll, timeout, func() (bool, error) {
		endpoint, err := kubeClientSet.CoreV1().Endpoints(ns).Get(name, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) {
			return false, nil
		}

		Expect(err).NotTo(HaveOccurred())

		if len(endpoint.Subsets) == 0 || len(endpoint.Subsets[0].Addresses) == 0 {
			return false, nil
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
		return nil, nil
	}

	if len(l.Items) == 0 {
		return nil, fmt.Errorf("there is no ingress-nginx pods running in namespace %v", ns)
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
		return nil, fmt.Errorf("there is no ingress-nginx pods running in namespace %v", ns)
	}

	return pod, nil
}

func createDeploymentWithRetries(c kubernetes.Interface, namespace string, obj *appsv1.Deployment) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().Deployments(namespace).Create(obj)
		if err == nil || k8sErrors.IsAlreadyExists(err) {
			return true, nil
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(createFunc)
}

func createSecretWithRetries(c kubernetes.Interface, namespace string, obj *v1.Secret) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.CoreV1().Secrets(namespace).Create(obj)
		if err == nil || k8sErrors.IsAlreadyExists(err) {
			return true, nil
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}
	return retryWithExponentialBackOff(createFunc)
}

func createServiceWithRetries(c kubernetes.Interface, namespace string, obj *v1.Service) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.CoreV1().Services(namespace).Create(obj)
		if err == nil || k8sErrors.IsAlreadyExists(err) {
			return true, nil
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(createFunc)
}

const (
	// Parameters for retrying with exponential backoff.
	retryBackoffInitialDuration = 100 * time.Millisecond
	retryBackoffFactor          = 3
	retryBackoffJitter          = 0
	retryBackoffSteps           = 6
)

// Utility for retrying the given function with exponential backoff.
func retryWithExponentialBackOff(fn wait.ConditionFunc) error {
	backoff := wait.Backoff{
		Duration: retryBackoffInitialDuration,
		Factor:   retryBackoffFactor,
		Jitter:   retryBackoffJitter,
		Steps:    retryBackoffSteps,
	}
	return wait.ExponentialBackoff(backoff, fn)
}

func isRetryableAPIError(err error) bool {
	// These errors may indicate a transient error that we can retry in tests.
	if k8sErrors.IsInternalError(err) || k8sErrors.IsTimeout(err) || k8sErrors.IsServerTimeout(err) ||
		k8sErrors.IsTooManyRequests(err) || utilnet.IsProbableEOF(err) || utilnet.IsConnectionReset(err) {
		return true
	}
	// If the error sends the Retry-After header, we respect it as an explicit confirmation we should retry.
	if _, shouldRetry := k8sErrors.SuggestsClientDelay(err); shouldRetry {
		return true
	}
	return false
}
