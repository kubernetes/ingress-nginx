/*
Copyright 2023 The Kubernetes Authors.

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

package customresponseheaders

import (
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
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
		},
	}
}

func TestParseInvalidAnnotations(t *testing.T) {
	ing := buildIngress()

	_, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err == nil {
		t.Errorf("expected error parsing ingress with custom-response-headers")
	}

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("custom-response-headers")] = `
		Content-Type application/json
		Access-Control-Max-Age: 600
		Nameok  	k8s.io/ingress-nginx/internal/ingress/annotations/customresponseheaders	0.003s	coverage: 88.5% of statements

	`
	ing.SetAnnotations(data)
	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err == nil {
		t.Errorf("expected error parsing ingress with custom-response-headers")
	}
	if i != nil {
		t.Errorf("expected %v but got %v", nil, i)
	}
}

func TestParseAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("custom-response-headers")] = `
		Content-Type: application/json
		Access-Control-Max-Age: 600
	`
	ing.SetAnnotations(data)

	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with custom-response-headers")
	}
	val, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a *Config type")
	}

	expected_response_headers := map[string]string{}
	expected_response_headers["Content-Type"] = "application/json"
	expected_response_headers["Access-Control-Max-Age"] = "600"

	c := &Config{expected_response_headers}

	if !reflect.DeepEqual(c, val) {
		t.Errorf("expected %v but got %v", c, val)
	}
}

func TestConfig_Equal(t *testing.T) {
	var nilConfig *Config

	config := &Config{
		ResponseHeaders: map[string]string{
			"nginx.ingress.kubernetes.io/custom-response-headers": "Cache-Control: no-cache",
		},
	}

	config2 := &Config{
		ResponseHeaders: map[string]string{
			"nginx.ingress.kubernetes.io/custom-response-headers": "Cache-Control: cache",
		},
	}

	configCopy := &Config{
		ResponseHeaders: map[string]string{
			"nginx.ingress.kubernetes.io/custom-response-headers": "Cache-Control: no-cache",
		},
	}

	if config.Equal(config2) {
		t.Errorf("config2 should not be equal to config")
	}

	if !config.Equal(configCopy) {
		t.Errorf("config should not be equal to configCopy")
	}

	if config.Equal(nilConfig) {
		t.Errorf("config should not be equal to nilConfig")
	}
}
