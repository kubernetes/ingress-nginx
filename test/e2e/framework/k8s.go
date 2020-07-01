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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
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
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// EnsureSecret creates a Secret object or returns it if it already exists.
func (f *Framework) EnsureSecret(secret *api.Secret) *api.Secret {
	err := createSecretWithRetries(f.KubeClientSet, f.Namespace, secret)
	assert.Nil(ginkgo.GinkgoT(), err, "creating secret")

	s, err := f.KubeClientSet.CoreV1().Secrets(secret.Namespace).Get(context.TODO(), secret.Name, metav1.GetOptions{})
	assert.Nil(ginkgo.GinkgoT(), err, "getting secret")
	assert.NotNil(ginkgo.GinkgoT(), s, "getting secret")

	return s
}

// EnsureConfigMap creates a ConfigMap object or returns it if it already exists.
func (f *Framework) EnsureConfigMap(configMap *api.ConfigMap) (*api.ConfigMap, error) {
	cm, err := f.KubeClientSet.CoreV1().ConfigMaps(f.Namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return f.KubeClientSet.CoreV1().ConfigMaps(f.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		}
		return nil, err
	}

	return cm, nil
}

// GetIngress gets an Ingress object from the given namespace, name and retunrs it, throws error if it does not exists.
func (f *Framework) GetIngress(namespace string, name string) *networking.Ingress {
	ing, err := f.KubeClientSet.NetworkingV1beta1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	assert.Nil(ginkgo.GinkgoT(), err, "getting ingress")
	assert.NotNil(ginkgo.GinkgoT(), ing, "expected an ingress but none returned")
	return ing
}

// EnsureIngress creates an Ingress object and retunrs it, throws error if it already exists.
func (f *Framework) EnsureIngress(ingress *networking.Ingress) *networking.Ingress {
	fn := func() {
		err := createIngressWithRetries(f.KubeClientSet, f.Namespace, ingress)
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")
	}

	f.waitForReload(fn)

	ing := f.GetIngress(f.Namespace, ingress.Name)
	if ing.Annotations == nil {
		ing.Annotations = make(map[string]string)
	}

	return ing
}

// UpdateIngress updates an Ingress object and returns the updated object.
func (f *Framework) UpdateIngress(ingress *networking.Ingress) *networking.Ingress {
	err := updateIngressWithRetries(f.KubeClientSet, f.Namespace, ingress)
	assert.Nil(ginkgo.GinkgoT(), err, "updating ingress")

	ing := f.GetIngress(f.Namespace, ingress.Name)
	if ing.Annotations == nil {
		ing.Annotations = make(map[string]string)
	}

	// updating an ingress requires a reload.
	Sleep(1 * time.Second)

	return ing
}

// EnsureService creates a Service object and retunrs it, throws error if it already exists.
func (f *Framework) EnsureService(service *core.Service) *core.Service {
	err := createServiceWithRetries(f.KubeClientSet, f.Namespace, service)
	assert.Nil(ginkgo.GinkgoT(), err, "creating service")

	s, err := f.KubeClientSet.CoreV1().Services(f.Namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
	assert.Nil(ginkgo.GinkgoT(), err, "getting service")
	assert.NotNil(ginkgo.GinkgoT(), s, "expected a service but none returned")

	return s
}

// EnsureDeployment creates a Deployment object and retunrs it, throws error if it already exists.
func (f *Framework) EnsureDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	err := createDeploymentWithRetries(f.KubeClientSet, f.Namespace, deployment)
	assert.Nil(ginkgo.GinkgoT(), err, "creating deployment")

	d, err := f.KubeClientSet.AppsV1().Deployments(deployment.Namespace).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
	assert.Nil(ginkgo.GinkgoT(), err, "getting deployment")
	assert.NotNil(ginkgo.GinkgoT(), d, "expected a deployment but none returned")

	return d
}

// waitForPodsReady waits for a given amount of time until a group of Pods is running in the given namespace.
func waitForPodsReady(kubeClientSet kubernetes.Interface, timeout time.Duration, expectedReplicas int, namespace string, opts metav1.ListOptions) error {
	return wait.Poll(Poll, timeout, func() (bool, error) {
		pl, err := kubeClientSet.CoreV1().Pods(namespace).List(context.TODO(), opts)
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

// waitForPodsDeleted waits for a given amount of time until a group of Pods are deleted in the given namespace.
func waitForPodsDeleted(kubeClientSet kubernetes.Interface, timeout time.Duration, namespace string, opts metav1.ListOptions) error {
	return wait.Poll(Poll, timeout, func() (bool, error) {
		pl, err := kubeClientSet.CoreV1().Pods(namespace).List(context.TODO(), opts)
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

	return wait.PollImmediate(Poll, timeout, func() (bool, error) {
		endpoint, err := kubeClientSet.CoreV1().Endpoints(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) {
			return false, nil
		}

		assert.Nil(ginkgo.GinkgoT(), err, "getting endpoints")

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

// GetIngressNGINXPod returns the ingress controller running pod
func GetIngressNGINXPod(ns string, kubeClientSet kubernetes.Interface) (*core.Pod, error) {
	var pod *core.Pod
	err := wait.Poll(Poll, DefaultTimeout, func() (bool, error) {
		l, err := kubeClientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=ingress-nginx",
		})
		if err != nil {
			return false, nil
		}

		for _, p := range l.Items {
			if strings.HasPrefix(p.GetName(), "nginx-ingress-controller") {
				isRunning, err := podRunningReady(&p)
				if err != nil {
					continue
				}

				if isRunning {
					pod = &p
					return true, nil
				}
			}
		}

		return false, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return nil, fmt.Errorf("timeout waiting at least one ingress-nginx pod running in namespace %v", ns)
		}

		return nil, err
	}

	return pod, nil
}

func createDeploymentWithRetries(c kubernetes.Interface, namespace string, obj *appsv1.Deployment) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.AppsV1().Deployments(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil {
			return true, nil
		}
		if k8sErrors.IsAlreadyExists(err) {
			return false, err
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
		_, err := c.CoreV1().Secrets(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil {
			return true, nil
		}
		if k8sErrors.IsAlreadyExists(err) {
			return false, err
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
		_, err := c.CoreV1().Services(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil {
			return true, nil
		}
		if k8sErrors.IsAlreadyExists(err) {
			return false, err
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(createFunc)
}

func createIngressWithRetries(c kubernetes.Interface, namespace string, obj *networking.Ingress) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.NetworkingV1beta1().Ingresses(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil {
			return true, nil
		}
		if k8sErrors.IsAlreadyExists(err) {
			return false, err
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to create object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(createFunc)
}

func updateIngressWithRetries(c kubernetes.Interface, namespace string, obj *networking.Ingress) error {
	if obj == nil {
		return fmt.Errorf("Object provided to create is empty")
	}
	updateFunc := func() (bool, error) {
		_, err := c.NetworkingV1beta1().Ingresses(namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err == nil {
			return true, nil
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to update object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(updateFunc)
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
