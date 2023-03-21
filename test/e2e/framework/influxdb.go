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

package framework

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const influxConfig = `
reporting-disabled = true
bind-address = "0.0.0.0:8088"

[meta]
  dir = "/var/lib/influxdb/meta"
  retention-autocreate = true
  logging-enabled = true

[data]
  dir = "/var/lib/influxdb/data"
  index-version = "inmem"
  wal-dir = "/var/lib/influxdb/wal"
  wal-fsync-delay = "0s"
  query-log-enabled = true
  cache-max-memory-size = 1073741824
  cache-snapshot-memory-size = 26214400
  cache-snapshot-write-cold-duration = "10m0s"
  compact-full-write-cold-duration = "4h0m0s"
  max-series-per-database = 1000000
  max-values-per-tag = 100000
  max-concurrent-compactions = 0
  trace-logging-enabled = false

[[udp]]
  enabled = true
  bind-address = ":8089"
  database = "nginx"
`

// NewInfluxDBDeployment creates an InfluxDB server configured to reply
// on 8086/tcp and 8089/udp
func (f *Framework) NewInfluxDBDeployment() {
	configuration := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "influxdb-config",
			Namespace: f.Namespace,
		},
		Data: map[string]string{
			"influxd.conf": influxConfig,
		},
	}

	f.EnsureConfigMap(configuration)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "influxdb",
			Namespace: f.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: NewInt32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "influxdb",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "influxdb",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(0),
					Volumes: []corev1.Volume{
						{
							Name: "influxdb-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "influxdb-config",
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "influxdb",
							Image:   "docker.io/influxdb:1.5",
							Env:     []corev1.EnvVar{},
							Command: []string{"influxd", "-config", "/influxdb-config/influxd.conf"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "influxdb-config",
									ReadOnly:  true,
									MountPath: "/influxdb-config",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8086,
								},
								{
									Name:          "udp",
									ContainerPort: 8089,
								},
							},
						},
					},
				},
			},
		},
	}

	d := f.EnsureDeployment(deployment)

	err := waitForPodsReady(f.KubeClientSet, DefaultTimeout, 1, f.Namespace, metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(d.Spec.Template.ObjectMeta.Labels)).String(),
	})
	assert.Nil(ginkgo.GinkgoT(), err, "waiting for influxdb pod to become ready")
}
