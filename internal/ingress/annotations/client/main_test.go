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

package client

import (
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func TestParse(t *testing.T) {
	annotation := parser.GetAnnotationWithPrefix("client-body-buffer-size")
	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	testCases := []struct {
		annotations map[string]string
		expected    string
	}{
		{map[string]string{annotation: "8k"}, "8k"},
		{map[string]string{annotation: "16k"}, "16k"},
		{map[string]string{annotation: "10000"}, "10000"},
		{map[string]string{annotation: "16R"}, ""},
		{map[string]string{annotation: "16kkk"}, ""},
		{map[string]string{annotation: ""}, ""},
		{map[string]string{}, ""},
		{nil, ""},
	}

	ing := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{},
	}

	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		//nolint:errcheck // Ignore the error since invalid cases will be checked with expected results
		res, _ := ap.Parse(ing)
		c, ok := res.(*Config)
		if !ok {
			t.Fatal("expected a client.Config type")
		}
		if c.BodyBufferSize != testCase.expected {
			t.Errorf("expected %v but returned %v, annotations: %s", testCase.expected, res, testCase.annotations)
		}
	}
}
