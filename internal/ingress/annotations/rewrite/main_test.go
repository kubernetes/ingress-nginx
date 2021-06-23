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
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	defRoute = "/demo"
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
			Name:        "foo",
			Namespace:   api.NamespaceDefault,
			Annotations: map[string]string{},
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

func TestRedirect(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("rewrite-target")] = defRoute
	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if redirect.Target != defRoute {
		t.Errorf("Expected %v as redirect but returned %s", defRoute, redirect.Target)
	}
}

func TestSSLRedirect(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("rewrite-target")] = defRoute
	ing.SetAnnotations(data)

	i, _ := NewParser(mockBackend{redirect: true}).Parse(ing)
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if !redirect.SSLRedirect {
		t.Errorf("Expected true but returned false")
	}

	data[parser.GetAnnotationWithPrefix("ssl-redirect")] = "false"
	ing.SetAnnotations(data)

	i, _ = NewParser(mockBackend{redirect: false}).Parse(ing)
	redirect, ok = i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if redirect.SSLRedirect {
		t.Errorf("Expected false but returned true")
	}
}

func TestForceSSLRedirect(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("rewrite-target")] = defRoute
	ing.SetAnnotations(data)

	i, _ := NewParser(mockBackend{redirect: true}).Parse(ing)
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if redirect.ForceSSLRedirect {
		t.Errorf("Expected false but returned true")
	}

	data[parser.GetAnnotationWithPrefix("force-ssl-redirect")] = "true"
	ing.SetAnnotations(data)

	i, _ = NewParser(mockBackend{redirect: false}).Parse(ing)
	redirect, ok = i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if !redirect.ForceSSLRedirect {
		t.Errorf("Expected true but returned false")
	}
}
func TestAppRoot(t *testing.T) {
	ap := NewParser(mockBackend{redirect: true})

	testCases := []struct {
		title       string
		path        string
		expected    string
		errExpected bool
	}{
		{"Empty path should return an error", "", "", true},
		{"Relative paths are not allowed", "demo", "", true},
		{"Path / should pass", "/", "/", false},
		{"Path /demo should pass", "/demo", "/demo", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			ing := buildIngress()
			ing.Annotations[parser.GetAnnotationWithPrefix("app-root")] = testCase.path
			i, err := ap.Parse(ing)
			if err != nil {
				if testCase.errExpected {
					return
				}

				t.Fatalf("%v: unexpected error obtaining running address/es: %v", testCase.title, err)
			}

			rewrite, ok := i.(*Config)
			if !ok {
				t.Fatalf("expected a rewrite Config")
			}

			if testCase.expected != rewrite.AppRoot {
				t.Fatalf("%v: expected AppRoot with value %v but was returned: %v", testCase.title, testCase.expected, rewrite.AppRoot)
			}
		})
	}
}

func TestUseRegex(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("use-regex")] = "true"
	ing.SetAnnotations(data)

	i, _ := NewParser(mockBackend{redirect: true}).Parse(ing)
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a App Context")
	}
	if !redirect.UseRegex {
		t.Errorf("Unexpected value got in UseRegex")
	}
}
