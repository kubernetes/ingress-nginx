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
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewGRPCFortuneTellerDeployment creates a new single replica
// deployment of the fortune teller image in a particular namespace
func (f *Framework) NewGRPCFortuneTellerDeployment() {
	f.NewNewGRPCFortuneTellerDeploymentWithReplicas(1)
}

// NewNewGRPCFortuneTellerDeploymentWithReplicas creates a new deployment of the
// fortune teller image in a particular namespace. Number of replicas is configurable
func (f *Framework) NewNewGRPCFortuneTellerDeploymentWithReplicas(replicas int32) {
	deployment := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fortune-teller",
			Namespace: f.IngressController.Namespace,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: NewInt32(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "fortune-teller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "fortune-teller",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(0),
					Containers: []corev1.Container{
						{
							Name:  "fortune-teller",
							Image: "quay.io/kubernetes-ingress-controller/grpc-fortune-teller:0.1",
							Env:   []corev1.EnvVar{},
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 50051,
								},
							},
						},
					},
				},
			},
		},
	}

	d, err := f.EnsureDeployment(deployment)
	Expect(err).NotTo(HaveOccurred())
	Expect(d).NotTo(BeNil(), "expected a fortune-teller deployment")

	err = WaitForPodsReady(f.KubeClientSet, DefaultTimeout, int(replicas), f.IngressController.Namespace, metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(d.Spec.Template.ObjectMeta.Labels)).String(),
	})
	Expect(err).NotTo(HaveOccurred(), "failed to wait for to become ready")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fortune-teller",
			Namespace: f.IngressController.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "grpc",
					Port:       50051,
					TargetPort: intstr.FromInt(50051),
					Protocol:   "TCP",
				},
			},
			Selector: map[string]string{
				"app": "fortune-teller",
			},
		},
	}

	f.EnsureService(service)
}
