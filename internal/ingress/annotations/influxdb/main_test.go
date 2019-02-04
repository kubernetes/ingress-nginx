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

package influxdb

import (
	"testing"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
			Rules: []extensions.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Path:    "/foo",
									Backend: defaultBackend,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestIngressInvalidInfluxDB(t *testing.T) {
	ing := buildIngress()

	influx, _ := NewParser(&resolver.Mock{}).Parse(ing)
	nginxInflux, ok := influx.(*Config)
	if !ok {
		t.Errorf("expected a Config type")
	}

	if nginxInflux.InfluxDBEnabled == true {
		t.Errorf("expected influxdb enabled but returned %v", nginxInflux.InfluxDBEnabled)
	}

	if nginxInflux.InfluxDBMeasurement != "default" {
		t.Errorf("expected measurement name not found. Found %v", nginxInflux.InfluxDBMeasurement)
	}

	if nginxInflux.InfluxDBPort != "8089" {
		t.Errorf("expected port not found. Found %v", nginxInflux.InfluxDBPort)
	}

	if nginxInflux.InfluxDBHost != "127.0.0.1" {
		t.Errorf("expected host not found. Found %v", nginxInflux.InfluxDBHost)
	}

	if nginxInflux.InfluxDBServerName != "nginx-ingress" {
		t.Errorf("expected server name not found. Found %v", nginxInflux.InfluxDBServerName)
	}
}

func TestIngressInfluxDB(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("enable-influxdb")] = "true"
	data[parser.GetAnnotationWithPrefix("influxdb-measurement")] = "nginxmeasures"
	data[parser.GetAnnotationWithPrefix("influxdb-port")] = "9091"
	data[parser.GetAnnotationWithPrefix("influxdb-host")] = "10.99.0.13"
	data[parser.GetAnnotationWithPrefix("influxdb-server-name")] = "nginx-test-1"
	ing.SetAnnotations(data)

	influx, _ := NewParser(&resolver.Mock{}).Parse(ing)
	nginxInflux, ok := influx.(*Config)
	if !ok {
		t.Errorf("expected a Config type")
	}

	if !nginxInflux.InfluxDBEnabled {
		t.Errorf("expected influxdb enabled but returned %v", nginxInflux.InfluxDBEnabled)
	}

	if nginxInflux.InfluxDBMeasurement != "nginxmeasures" {
		t.Errorf("expected measurement name not found. Found %v", nginxInflux.InfluxDBMeasurement)
	}

	if nginxInflux.InfluxDBPort != "9091" {
		t.Errorf("expected port not found. Found %v", nginxInflux.InfluxDBPort)
	}

	if nginxInflux.InfluxDBHost != "10.99.0.13" {
		t.Errorf("expected host not found. Found %v", nginxInflux.InfluxDBHost)
	}

	if nginxInflux.InfluxDBServerName != "nginx-test-1" {
		t.Errorf("expected server name not found. Found %v", nginxInflux.InfluxDBServerName)
	}
}
