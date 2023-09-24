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
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
	defaultBackend := networking.IngressBackend{
		Service: &networking.IngressServiceBackend{
			Name: "default-backend",
			Port: networking.ServiceBackendPort{
				Number: 80,
			},
		},
	}

	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-backend",
					Port: networking.ServiceBackendPort{
						Number: 80,
					},
				},
			},
			Rules: []networking.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
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

func TestIngressCorsConfigValid(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}

	// Valid
	data[parser.GetAnnotationWithPrefix(corsEnableAnnotation)] = "true"
	data[parser.GetAnnotationWithPrefix(corsAllowHeadersAnnotation)] = "DNT,X-CustomHeader, Keep-Alive,User-Agent"
	data[parser.GetAnnotationWithPrefix(corsAllowCredentialsAnnotation)] = "false"
	data[parser.GetAnnotationWithPrefix(corsAllowMethodsAnnotation)] = "GET, PATCH"
	data[parser.GetAnnotationWithPrefix(corsAllowOriginAnnotation)] = "https://origin123.test.com:4443"
	data[parser.GetAnnotationWithPrefix(corsExposeHeadersAnnotation)] = "*, X-CustomResponseHeader"
	data[parser.GetAnnotationWithPrefix(corsMaxAgeAnnotation)] = "600"
	ing.SetAnnotations(data)

	corst, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("error parsing annotations: %v", err)
	}

	nginxCors, ok := corst.(*Config)
	if !ok {
		t.Errorf("expected a Config type but returned %t", corst)
	}

	if !nginxCors.CorsEnabled {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsEnableAnnotation)], nginxCors.CorsEnabled)
	}

	if nginxCors.CorsAllowCredentials {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsAllowCredentialsAnnotation)], nginxCors.CorsAllowCredentials)
	}

	if nginxCors.CorsAllowHeaders != "DNT,X-CustomHeader, Keep-Alive,User-Agent" {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsAllowHeadersAnnotation)], nginxCors.CorsAllowHeaders)
	}

	if nginxCors.CorsAllowMethods != "GET, PATCH" {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsAllowMethodsAnnotation)], nginxCors.CorsAllowMethods)
	}

	if nginxCors.CorsAllowOrigin[0] != "https://origin123.test.com:4443" {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsAllowOriginAnnotation)], nginxCors.CorsAllowOrigin)
	}

	if nginxCors.CorsExposeHeaders != "*, X-CustomResponseHeader" {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsExposeHeadersAnnotation)], nginxCors.CorsExposeHeaders)
	}

	if nginxCors.CorsMaxAge != 600 {
		t.Errorf("expected %v but returned %v", data[parser.GetAnnotationWithPrefix(corsMaxAgeAnnotation)], nginxCors.CorsMaxAge)
	}
}

func TestIngressCorsConfigInvalid(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}

	// Valid
	data[parser.GetAnnotationWithPrefix(corsEnableAnnotation)] = "yes"
	data[parser.GetAnnotationWithPrefix(corsAllowHeadersAnnotation)] = "@alright, #ingress"
	data[parser.GetAnnotationWithPrefix(corsAllowCredentialsAnnotation)] = "no"
	data[parser.GetAnnotationWithPrefix(corsAllowMethodsAnnotation)] = "GET, PATCH, $nginx"
	data[parser.GetAnnotationWithPrefix(corsAllowOriginAnnotation)] = "origin123.test.com:4443"
	data[parser.GetAnnotationWithPrefix(corsExposeHeadersAnnotation)] = "@alright, #ingress"
	data[parser.GetAnnotationWithPrefix(corsMaxAgeAnnotation)] = "abcd"
	ing.SetAnnotations(data)

	corst, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("error parsing annotations: %v", err)
	}

	nginxCors, ok := corst.(*Config)
	if !ok {
		t.Errorf("expected a Config type but returned %t", corst)
	}

	if nginxCors.CorsEnabled {
		t.Errorf("expected %v but returned %v", false, nginxCors.CorsEnabled)
	}

	if !nginxCors.CorsAllowCredentials {
		t.Errorf("expected %v but returned %v", true, nginxCors.CorsAllowCredentials)
	}

	if nginxCors.CorsAllowHeaders != defaultCorsHeaders {
		t.Errorf("expected %v but returned %v", defaultCorsHeaders, nginxCors.CorsAllowHeaders)
	}

	if nginxCors.CorsAllowMethods != defaultCorsMethods {
		t.Errorf("expected %v but returned %v", defaultCorsHeaders, nginxCors.CorsAllowMethods)
	}

	if nginxCors.CorsExposeHeaders != "" {
		t.Errorf("expected %v but returned %v", "", nginxCors.CorsExposeHeaders)
	}

	if nginxCors.CorsMaxAge != defaultCorsMaxAge {
		t.Errorf("expected %v but returned %v", defaultCorsMaxAge, nginxCors.CorsMaxAge)
	}
}
