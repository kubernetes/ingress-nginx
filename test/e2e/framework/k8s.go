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
	"time"

	api "k8s.io/api/core/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// EnsureSecret creates a Secret object or returns it if it already exists.
func (f *Framework) EnsureSecret(secret *api.Secret) (*api.Secret, error) {
	s, err := f.KubeClientSet.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return f.KubeClientSet.CoreV1().Secrets(secret.Namespace).Update(secret)
		}
		return nil, err
	}
	return s, nil
}

// EnsureIngress creates an Ingress object or returns it if it already exists.
func (f *Framework) EnsureIngress(ingress *extensions.Ingress) (*extensions.Ingress, error) {
	s, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(ingress)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return f.KubeClientSet.ExtensionsV1beta1().Ingresses(ingress.Namespace).Create(ingress)
		}
		return nil, err
	}
	return s, nil
}

// EnsureService creates a Service object or returns it if it already exists.
func (f *Framework) EnsureService(service *core.Service) (*core.Service, error) {
	s, err := f.KubeClientSet.CoreV1().Services(service.Namespace).Update(service)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return f.KubeClientSet.CoreV1().Services(service.Namespace).Create(service)
		}
		return nil, err
	}
	return s, nil
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

// WaitForPodsReady waits for a given amount of time until a group of Pods is running.
func (f *Framework) WaitForPodsReady(timeout time.Duration, expectedReplicas int, opts metav1.ListOptions) error {
	return wait.Poll(time.Second, timeout, func() (bool, error) {
		pl, err := f.KubeClientSet.CoreV1().Pods(f.Namespace.Name).List(opts)
		if err != nil {
			return false, err
		}

		r := 0
		for _, p := range pl.Items {
			if p.Status.Phase != core.PodRunning {
				continue
			}
			r++
		}

		if r == expectedReplicas {
			return true, nil
		}

		return false, nil
	})
}
