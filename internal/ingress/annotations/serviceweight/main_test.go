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

package serviceweight

import (
	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"

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

func TestParseAnnotations(t *testing.T) {
	ing := buildIngress()

	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	annotation := parser.GetAnnotationWithPrefix("service-weight")
	data := map[string]string{annotation: `new-svc: 20, old-svc: 60`}
	ing.SetAnnotations(data)

	s, err := ap.Parse(ing)
	if err != nil {
		t.Fatalf("parse service-weight annotation error")
	} else if config, ok := s.(*Config); ok {
		if w, ok := config.ServiceWeight["new-svc"]; !ok || w != "20" {
			t.Fatalf("parse service-weight annotation error")
		}
		if w, ok := config.ServiceWeight["old-svc"]; !ok || w != "60" {
			t.Fatalf("parse service-weight annotation error")
		}
	} else {
		t.Fatalf("parse service-weight annotation error")
	}
}
