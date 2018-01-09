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

package cors

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

func TestIngressCorsConfig(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("enable-cors")] = "true"
	data[parser.GetAnnotationWithPrefix("cors-allow-headers")] = "DNT,X-CustomHeader, Keep-Alive,User-Agent"
	data[parser.GetAnnotationWithPrefix("cors-allow-credentials")] = "false"
	data[parser.GetAnnotationWithPrefix("cors-allow-methods")] = "PUT, GET,OPTIONS, PATCH, $nginx_version"
	data[parser.GetAnnotationWithPrefix("cors-allow-origin")] = "https://origin123.test.com:4443"
	data[parser.GetAnnotationWithPrefix("cors-max-age")] = "600"
	ing.SetAnnotations(data)

	corst, _ := NewParser(&resolver.Mock{}).Parse(ing)
	nginxCors, ok := corst.(*Config)
	if !ok {
		t.Errorf("expected a Config type")
	}

	if !nginxCors.CorsEnabled {
		t.Errorf("expected cors enabled but returned %v", nginxCors.CorsEnabled)
	}

	if nginxCors.CorsAllowHeaders != "DNT,X-CustomHeader, Keep-Alive,User-Agent" {
		t.Errorf("expected headers not found. Found %v", nginxCors.CorsAllowHeaders)
	}

	if nginxCors.CorsAllowMethods != "GET, PUT, POST, DELETE, PATCH, OPTIONS" {
		t.Errorf("expected default methods, but got  %v", nginxCors.CorsAllowMethods)
	}

	if nginxCors.CorsAllowOrigin != "https://origin123.test.com:4443" {
		t.Errorf("expected origin https://origin123.test.com:4443, but got  %v", nginxCors.CorsAllowOrigin)
	}

	if nginxCors.CorsMaxAge != 600 {
		t.Errorf("expected max age 600, but got  %v", nginxCors.CorsMaxAge)
	}
}
