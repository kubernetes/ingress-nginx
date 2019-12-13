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

	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EchoService name of the deployment for the echo app
const EchoService = "echo"

// SlowEchoService name of the deployment for the echo app
const SlowEchoService = "slow-echo"

// HTTPBinService name of the deployment for the httpbin app
const HTTPBinService = "httpbin"

// NewEchoDeployment creates a new single replica deployment of the echoserver image in a particular namespace
func (f *Framework) NewEchoDeployment() {
	f.NewEchoDeploymentWithReplicas(1)
}

// NewEchoDeploymentWithReplicas creates a new deployment of the echoserver image in a particular namespace. Number of
// replicas is configurable
func (f *Framework) NewEchoDeploymentWithReplicas(replicas int) {
	f.NewEchoDeploymentWithNameAndReplicas(EchoService, replicas)
}

// NewEchoDeploymentWithNameAndReplicas creates a new deployment of the echoserver image in a particular namespace. Number of
// replicas is configurable and
// name is configurable
func (f *Framework) NewEchoDeploymentWithNameAndReplicas(name string, replicas int) {

	data := map[string]string{}
	data["nginx.conf"] = `#

env HOSTNAME;
env NODE_NAME;
env POD_NAME;
env POD_NAMESPACE;
env POD_IP;

daemon off;

events {
    worker_connections  1024;
}

http {
	default_type 'text/plain';
	client_max_body_size 0;

	init_by_lua_block {
		local template = require "resty.template"

		tmpl = template.compile([[

Hostname: {*os.getenv("HOSTNAME") or "N/A"*}

Pod Information:
{% if os.getenv("POD_NAME") then %}
	node name:	{*os.getenv("NODE_NAME") or "N/A"*}
	pod name:	{*os.getenv("POD_NAME") or "N/A"*}
	pod namespace:	{*os.getenv("POD_NAMESPACE") or "N/A"*}
	pod IP:	{*os.getenv("POD_IP") or "N/A"*}
{% else %}
	-no pod information available-
{% end %}

Server values:
	server_version=nginx: {*ngx.var.nginx_version*} - lua: {*ngx.config.ngx_lua_version*}

Request Information:
	client_address={*ngx.var.remote_addr*}
	method={*ngx.req.get_method()*}
	real path={*ngx.var.request_uri*}
	query={*ngx.var.query_string or ""*}
	request_version={*ngx.req.http_version()*}
	request_scheme={*ngx.var.scheme*}
	request_uri={*ngx.var.scheme.."://"..ngx.var.host..":"..ngx.var.server_port..ngx.var.request_uri*}

Request Headers:
{% for i, key in ipairs(keys) do %}
	{% local val = headers[key] %}
	{% if type(val) == "table" then %}
		{% for i = 1,#val do %}
	{*key*}={*val[i]*}
		{% end %}
	{% else %}
	{*key*}={*val*}
	{% end %}
{% end %}

Request Body:
{*ngx.var.request_body or "	-no body in request-"*}
]])
	}

	server {
		listen 80 default_server reuseport;

		server_name _;

		keepalive_timeout 620s;

		location / {
			lua_need_request_body on;

			header_filter_by_lua_block {
				if ngx.var.arg_hsts == "true" then
					ngx.header["Strict-Transport-Security"] = "max-age=3600; preload"
				end
			}

			content_by_lua_block {
				ngx.header["Server"] = "echoserver"

				local headers = ngx.req.get_headers()
				local keys = {}
				for key, val in pairs(headers) do
					table.insert(keys, key)
				end
				table.sort(keys)

				ngx.say(tmpl({os=os, ngx=ngx, keys=keys, headers=headers}))
			}
		}
	}
}
`

	_, err := f.KubeClientSet.CoreV1().ConfigMaps(f.Namespace).Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.Namespace,
		},
		Data: data,
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create a deployment")

	deployment := newDeployment(name, f.Namespace, "openresty/openresty:1.15.8.2-alpine", 80, int32(replicas),
		[]string{
			"/bin/sh",
			"-c",
			"apk add -U perl curl && opm get bungle/lua-resty-template && openresty",
		},
		[]corev1.VolumeMount{
			{
				Name:      name,
				MountPath: "/usr/local/openresty/nginx/conf/nginx.conf",
				SubPath:   "nginx.conf",
				ReadOnly:  true,
			},
		},
		[]corev1.Volume{
			{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: name,
						},
					},
				},
			},
		},
	)

	d := f.EnsureDeployment(deployment)
	Expect(d).NotTo(BeNil(), "expected a deployment but none returned")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
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

	err = WaitForEndpoints(f.KubeClientSet, DefaultTimeout, name, f.Namespace, replicas)
	Expect(err).NotTo(HaveOccurred(), "failed to wait for endpoints to become ready")
}

// NewSlowEchoDeployment creates a new deployment of the slow echo server image in a particular namespace.
func (f *Framework) NewSlowEchoDeployment() {
	data := map[string]string{}
	data["default.conf"] = `#

server {
	access_log on;
	access_log /dev/stdout;

	listen 80;

	location / {
		echo ok;
	}

	location ~ ^/sleep/(?<sleepTime>[0-9]+)$ {
		echo_sleep $sleepTime;
		echo "ok after $sleepTime seconds";
	}
}

`

	_, err := f.KubeClientSet.CoreV1().ConfigMaps(f.Namespace).Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SlowEchoService,
			Namespace: f.Namespace,
		},
		Data: data,
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create a deployment")

	deployment := newDeployment(SlowEchoService, f.Namespace, "openresty/openresty:1.15.8.2-alpine", 80, 1,
		nil,
		[]corev1.VolumeMount{
			{
				Name:      SlowEchoService,
				MountPath: "/etc/nginx/conf.d",
				ReadOnly:  true,
			},
		},
		[]corev1.Volume{
			{
				Name: SlowEchoService,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: SlowEchoService,
						},
					},
				},
			},
		},
	)

	d := f.EnsureDeployment(deployment)
	Expect(d).NotTo(BeNil(), "expected a deployment but none returned")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SlowEchoService,
			Namespace: f.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": SlowEchoService,
			},
		},
	}

	s := f.EnsureService(service)
	Expect(s).NotTo(BeNil(), "expected a service but none returned")

	err = WaitForEndpoints(f.KubeClientSet, DefaultTimeout, SlowEchoService, f.Namespace, 1)
	Expect(err).NotTo(HaveOccurred(), "failed to wait for endpoints to become ready")
}

func newDeployment(name, namespace, image string, port int32, replicas int32, command []string,
	volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) *appsv1.Deployment {
	probe := &corev1.Probe{
		InitialDelaySeconds: 1,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromString("http"),
				Path: "/",
			},
		},
	}

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
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
							VolumeMounts:   volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if len(command) > 0 {
		d.Spec.Template.Spec.Containers[0].Command = command
	}

	return d
}

// NewHttpbinDeployment creates a new single replica deployment of the httpbin image in a particular namespace.
func (f *Framework) NewHttpbinDeployment() {
	f.NewDeployment(HTTPBinService, "ingress-controller/httpbin:dev", 80, 1)
}

// NewDeployment creates a new deployment in a particular namespace.
func (f *Framework) NewDeployment(name, image string, port int32, replicas int32) {
	deployment := newDeployment(name, f.Namespace, image, port, replicas, nil, nil, nil)

	d := f.EnsureDeployment(deployment)
	Expect(d).NotTo(BeNil(), "expected a deployment but none returned")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.Namespace,
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

	err := WaitForEndpoints(f.KubeClientSet, DefaultTimeout, name, f.Namespace, int(replicas))
	Expect(err).NotTo(HaveOccurred(), "failed to wait for endpoints to become ready")
}

// DeleteDeployment deletes a deployment with a particular name and waits for the pods to be deleted
func (f *Framework) DeleteDeployment(name string) error {
	d, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred(), "failed to get a deployment")
	err = f.KubeClientSet.AppsV1().Deployments(f.Namespace).Delete(name, &metav1.DeleteOptions{})
	Expect(err).NotTo(HaveOccurred(), "failed to delete a deployment")
	return WaitForPodsDeleted(f.KubeClientSet, time.Second*60, f.Namespace, metav1.ListOptions{
		LabelSelector: labelSelectorToString(d.Spec.Selector.MatchLabels),
	})
}

// ScaleDeploymentToZero scales a deployment with a particular name and waits for the pods to be deleted
func (f *Framework) ScaleDeploymentToZero(name string) {
	d, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred(), "failed to get a deployment")
	Expect(d).NotTo(BeNil(), "expected a deployment but none returned")

	d.Spec.Replicas = NewInt32(0)

	d = f.EnsureDeployment(d)
	Expect(d).NotTo(BeNil(), "expected a deployment but none returned")
}
