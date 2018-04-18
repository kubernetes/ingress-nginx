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
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewEchoDeployment creates a new single replica deployment of the echoserver image in a particular namespace
func (f *Framework) NewEchoDeployment() error {
	return f.NewEchoDeploymentWithReplicas(1)
}

// NewEchoDeploymentWithReplicas creates a new deployment of the echoserver image in a particular namespace. Number of
// replicas is configurable
func (f *Framework) NewEchoDeploymentWithReplicas(replicas int32) error {
	deployment := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-svc",
			Namespace: f.IngressController.Namespace,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: NewInt32(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "http-svc",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "http-svc",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(0),
					Containers: []corev1.Container{
						{
							Name:  "http-svc",
							Image: "gcr.io/google_containers/echoserver:1.8",
							Env:   []corev1.EnvVar{},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}

	d, err := f.EnsureDeployment(deployment)
	if err != nil {
		return err
	}

	if d == nil {
		return fmt.Errorf("unexpected error creating deployement for echoserver")
	}

	err = WaitForPodsReady(f.KubeClientSet, 5*time.Minute, int(replicas), f.IngressController.Namespace, metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(d.Spec.Template.ObjectMeta.Labels)).String(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to wait for to become ready")
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-svc",
			Namespace: f.IngressController.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   "TCP",
				},
			},
			Selector: map[string]string{
				"app": "http-svc",
			},
		},
	}

	s, err := f.EnsureService(service)
	if err != nil {
		return err
	}

	if s == nil {
		return fmt.Errorf("unexpected error creating service for echoserver deployment")
	}

	return nil
}
