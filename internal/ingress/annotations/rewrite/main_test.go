/*
Copyright 2015 The Kubernetes Authors.

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

package rewrite

import (
	"testing"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	defRoute = "/demo"
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

type mockBackend struct {
	resolver.Mock
	redirect bool
}

func (m mockBackend) GetDefaultBackend() defaults.Backend {
	return defaults.Backend{SSLRedirect: m.redirect}
}

func TestWithoutAnnotations(t *testing.T) {
	ing := buildIngress()
	_, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error with ingress without annotations: %v", err)
	}
}

func buildRedirect(t *testing.T, data *map[string]string, mock *mockBackend) *Config {
	ing := buildIngress()
	ing.SetAnnotations(*data)

	if mock == nil {
		mock = &mockBackend{}
	}

	i1, err := NewParser(mock).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	redirect, ok := i1.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	return redirect
}

func buildDifferentRedirect(t *testing.T, key string, val string) *Config {
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(key)] = val
	return buildRedirect(t, &data, nil)
}

func TestEqual(t *testing.T) {
	data := map[string]string{}

	redirect := buildRedirect(t, &data, nil)

	if !redirect.Equal(redirect) {
		t.Errorf("Expect the both redirect types to be equal")
	}

	if redirect.Equal(nil) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "location-modifier", "=")) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "rewrite-target", "/")) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "ssl-redirect", "true")) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "force-ssl-redirect", "true")) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "add-base-url", "true")) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "base-url-scheme", "/scheme")) {
		t.Errorf("Expect the both redirect types to be different")
	}

	if redirect.Equal(buildDifferentRedirect(t, "app-root", "/root")) {
		t.Errorf("Expect the both redirect types to be different")
	}
}

func TestRedirect(t *testing.T) {
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("rewrite-target")] = defRoute
	redirect := buildRedirect(t, &data, nil)

	if redirect.Target != defRoute {
		t.Errorf("Expected %v as redirect but returned %s", defRoute, redirect.Target)
	}
}

func TestRegex(t *testing.T) {
	data := map[string]string{}
	modifier := "~"
	data[parser.GetAnnotationWithPrefix("location-modifier")] = modifier

	redirect := buildRedirect(t, &data, nil)
	if redirect.LocationModifier != modifier {
		t.Errorf("Expected %v as location modifier but returned %s", modifier, redirect.LocationModifier)
	}
}

func TestSSLRedirect(t *testing.T) {
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("rewrite-target")] = defRoute

	redirect := buildRedirect(t, &data, &mockBackend{redirect: true})
	if !redirect.SSLRedirect {
		t.Errorf("Expected true but returned false")
	}
	data[parser.GetAnnotationWithPrefix("ssl-redirect")] = "false"

	redirect = buildRedirect(t, &data, &mockBackend{redirect: false})
	if redirect.SSLRedirect {
		t.Errorf("Expected false but returned true")
	}
}

func TestForceSSLRedirect(t *testing.T) {
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("rewrite-target")] = defRoute

	redirect := buildRedirect(t, &data, nil)
	if redirect.ForceSSLRedirect {
		t.Errorf("Expected false but returned true")
	}
	data[parser.GetAnnotationWithPrefix("force-ssl-redirect")] = "true"

	redirect = buildRedirect(t, &data, nil)
	if !redirect.ForceSSLRedirect {
		t.Errorf("Expected true but returned false")
	}
}
func TestAppRoot(t *testing.T) {
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("app-root")] = "/app1"

	redirect := buildRedirect(t, &data, nil)
	if redirect.AppRoot != "/app1" {
		t.Errorf("Unexpected value got in AppRoot")
	}
}
