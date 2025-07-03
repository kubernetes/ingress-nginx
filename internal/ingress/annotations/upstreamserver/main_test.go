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

package upstreamserver

import (
	"fmt"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
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

func TestIngressAnnotationUpstreamServerMaxConnsSetPositive(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(upstreamServerMaxConnsAnnotation)] = "100"
	ing.SetAnnotations(data)

	val, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	config, ok := val.(*Config)
	if !ok {
		t.Errorf("expected a *Config type")
	}

	if config.MaxConns != 100 {
		t.Errorf("expected annotation value to be 100, got %d", config.MaxConns)
	}
}

func TestIngressAnnotationUpstreamServerMaxConnsSetNegative(t *testing.T) {
	ing := buildIngress()

	// Test with explicitly set to negative value
	expectedErr := fmt.Errorf("annotation nginx.ingress.kubernetes.io/%s contains invalid value", upstreamServerMaxConnsAnnotation)
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(upstreamServerMaxConnsAnnotation)] = "-1"
	ing.SetAnnotations(data)

	_, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected error: %v, got %v", expectedErr, err)
	}
}

func TestIngressAnnotationUpstreamServerMaxConnsSetString(t *testing.T) {
	ing := buildIngress()

	// Test with explicitly set to negative value
	expectedErr := fmt.Errorf("annotation nginx.ingress.kubernetes.io/%s contains invalid value", upstreamServerMaxConnsAnnotation)
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(upstreamServerMaxConnsAnnotation)] = "uhi"
	ing.SetAnnotations(data)

	_, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected error: %v, got %v", expectedErr, err)
	}
}

type mockBackend struct {
	resolver.Mock
}

// GetDefaultBackend returns the backend that must be used as default
func (m mockBackend) GetDefaultBackend() defaults.Backend {
	return defaults.Backend{
		UpstreamServerMaxConns: 0,
	}
}

// Test that when we have a default configuration set on the Backend that is used
// when we don't have the annotation
func TestParseAnnotationsWithDefaultConfig(t *testing.T) {
	ing := buildIngress()

	val, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	config, ok := val.(*Config)

	if !ok {
		t.Errorf("expected a *Config type")
	}

	if config.MaxConns != 0 {
		t.Errorf("expected annotation value to be 0, got %d", config.MaxConns)
	}
}

// Test that the annotation will disable the service upstream when enabled
// in the default configuration
func TestParseAnnotationsOverridesDefaultConfig(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(upstreamServerMaxConnsAnnotation)] = "0"
	ing.SetAnnotations(data)

	val, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	config, ok := val.(*Config)

	if !ok {
		t.Errorf("expected a *Config type")
	}

	if config.MaxConns != 0 {
		t.Errorf("expected annotation value to be 0, got %d", config.MaxConns)
	}
}
