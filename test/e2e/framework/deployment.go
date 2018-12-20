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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewEchoDeployment creates a new single replica deployment of the echoserver image in a particular namespace
func (f *Framework) NewEchoDeployment() {
	f.NewEchoDeploymentWithReplicas(1)
}

// NewEchoDeploymentWithReplicas creates a new deployment of the echoserver image in a particular namespace. Number of
// replicas is configurable
func (f *Framework) NewEchoDeploymentWithReplicas(replicas int32) {
	f.NewEchoDeploymentWithNameAndReplicas("http-svc", replicas)
}

// NewEchoDeploymentWithNameAndReplicas creates a new deployment of the echoserver image in a particular namespace. Number of
// replicas is configurable and
// name is configurable
func (f *Framework) NewEchoDeploymentWithNameAndReplicas(name string, replicas int32) {
	f.NewDeployment(name, "gcr.io/kubernetes-e2e-test-images/echoserver:2.2", 8080, replicas)
}

// NewSlowEchoDeployment creates a new deployment of the slow echo server image in a particular namespace.
func (f *Framework) NewSlowEchoDeployment() {
	f.NewDeployment("slowecho", "breunigs/slowechoserver", 8080, 1)
}

// NewHttpbinDeployment creates a new single replica deployment of the httpbin image in a particular namespace.
func (f *Framework) NewHttpbinDeployment() {
	f.NewDeployment("httpbin", "kennethreitz/httpbin", 80, 1)
}

// NewDeployment creates a new deployment in a particular namespace.
func (f *Framework) NewDeployment(name, image string, port int32, replicas int32) {
	probe := &corev1.Probe{
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromString("http"),
				Path: "/",
			},
		},
	}

	deployment := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.IngressController.Namespace,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: NewInt32(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(0),
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Env:   []corev1.EnvVar{},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: port,
								},
							},
							ReadinessProbe: probe,
							LivenessProbe:  probe,
						},
					},
				},
			},
		},
	}

	d, err := f.EnsureDeployment(deployment)
	Expect(err).NotTo(HaveOccurred(), "failed to create a deployment")
	Expect(d).NotTo(BeNil(), "expected a deployement but none returned")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.IngressController.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(int(port)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": name,
			},
		},
	}

	s := f.EnsureService(service)
	Expect(s).NotTo(BeNil(), "expected a service but none returned")

	err = WaitForEndpoints(f.KubeClientSet, DefaultTimeout, name, f.IngressController.Namespace, int(replicas))
	Expect(err).NotTo(HaveOccurred(), "failed to wait for endpoints to become ready")
}
