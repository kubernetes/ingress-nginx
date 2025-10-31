/*
Copyright 2019 The Kubernetes Authors.

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

package customhttperrors

import (
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const DefaultHTTPErrorsString = "400,404,500,502"

var DefaultHTTPErrorsSlice = []int{400, 404, 500, 502}

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

func TestParseCodes(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("custom-http-errors")] = DefaultHTTPErrorsString
	ing.SetAnnotations(data)

	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with custom-http-errors")
	}
	val, ok := i.([]int)
	if !ok {
		t.Errorf("expected a []int type")
	}

	expected := DefaultHTTPErrorsSlice
	if !reflect.DeepEqual(expected, val) {
		t.Errorf("expected %v but got %v", expected, val)
	}
}

func TestEnabledSwitch(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("enable-custom-http-errors")] = "true"
	data[parser.GetAnnotationWithPrefix("custom-http-errors")] = DefaultHTTPErrorsString
	ing.SetAnnotations(data)

	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with custom-http-errors")
	}
	val, ok := i.([]int)
	if !ok {
		t.Errorf("expected a []int type")
	}

	expected := DefaultHTTPErrorsSlice
	if !reflect.DeepEqual(expected, val) {
		t.Errorf("expected %v but got %v", expected, val)
	}
}

func TestDisabledSwitch(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("enable-custom-http-errors")] = "false"
	data[parser.GetAnnotationWithPrefix("custom-http-errors")] = DefaultHTTPErrorsString
	ing.SetAnnotations(data)

	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with custom-http-errors")
	}
	val, ok := i.([]int)
	if !ok {
		t.Errorf("expected a []int type")
	}

	expected := []int{}
	if !reflect.DeepEqual(expected, val) {
		t.Errorf("expected %v but got %v", expected, val)
	}
}

func TestEnabledByDefault(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("enable-custom-http-errors")] = "fakebool"
	data[parser.GetAnnotationWithPrefix("custom-http-errors")] = DefaultHTTPErrorsString
	ing.SetAnnotations(data)

	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with custom-http-errors")
	}
	val, ok := i.([]int)
	if !ok {
		t.Errorf("expected a []int type")
	}

	expected := DefaultHTTPErrorsSlice
	if !reflect.DeepEqual(expected, val) {
		t.Errorf("expected %v but got %v", expected, val)
	}
}

func TestErrorOnInvalidCode(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("enable-custom-http-errors")] = "true"
	data[parser.GetAnnotationWithPrefix("custom-http-errors")] = "400,404,fakeint,502"
	ing.SetAnnotations(data)

	i, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err == nil {
		t.Errorf("expected error parsing ingress with custom-http-errors")
	}
	if i != nil {
		t.Errorf("expected %v but got %v", nil, i)
	}
}
