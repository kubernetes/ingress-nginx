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

package websocket

import (
	"testing"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func TestParse(t *testing.T) {
	annotation := parser.GetAnnotationWithPrefix("enable-websocket")
	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	testCases := []struct {
		annotations map[string]string
		expected    bool
	}{
		{map[string]string{annotation: "true"}, true},
		{map[string]string{annotation: "false"}, false},
		{map[string]string{annotation: ""}, true},
		{map[string]string{}, true},
		{nil, true},
	}

	ing := &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{
			Rules: []extensions.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Path: "/foo",
									Backend: extensions.IngressBackend{
										ServiceName: "test1",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		result, _ := ap.Parse(ing)
		ws, ok := result.(bool)
		if !ok {
			t.Errorf("expected a Config type")
		}

		if !ws == testCase.expected {
			t.Errorf("expected %v but returned %v, annotations: %s", testCase.expected, ws, testCase.annotations)
		}
	}
}
