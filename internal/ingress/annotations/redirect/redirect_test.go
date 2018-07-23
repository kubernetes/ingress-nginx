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

package redirect

import (
	"net/http"
	"strconv"
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
	defRedirect = "http://some-site.com"
	defCode     = http.StatusMovedPermanently
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
func TestPermanentRedirectWithDefaultCode(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("permanent-redirect")] = defRedirect
	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if redirect.URL != defRedirect {
		t.Errorf("Expected %v as redirect but returned %s", defRedirect, redirect.URL)
	}
	if redirect.Code != defCode {
		t.Errorf("Expected %v as redirect to have a code %s but had %s", defRedirect, strconv.Itoa(defCode), strconv.Itoa(redirect.Code))
	}
}

func TestPermanentRedirectWithCustomCode(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("permanent-redirect")] = defRedirect
	data[parser.GetAnnotationWithPrefix("permanent-redirect-code")] = "308"
	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a Redirect type")
	}
	if redirect.URL != defRedirect {
		t.Errorf("Expected %v as redirect but returned %s", defRedirect, redirect.URL)
	}
	if redirect.Code != http.StatusPermanentRedirect {
		t.Errorf("Expected %v as redirect to have a code %s but had %s", defRedirect, strconv.Itoa(http.StatusPermanentRedirect), strconv.Itoa(redirect.Code))
	}
}
