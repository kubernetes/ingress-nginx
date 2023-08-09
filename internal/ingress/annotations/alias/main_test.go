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

package alias

import (
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var annotation = parser.GetAnnotationWithPrefix(serverAliasAnnotation)

func TestParse(t *testing.T) {
	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	testCases := []struct {
		annotations    map[string]string
		expected       []string
		skipValidation bool
		wantErr        bool
	}{
		{map[string]string{annotation: "a.com, b.com, ,    c.com"}, []string{"a.com", "b.com", "c.com"}, false, false},
		{map[string]string{annotation: "www.example.com"}, []string{"www.example.com"}, false, false},
		{map[string]string{annotation: "*.example.com,www.example.*"}, []string{"*.example.com", "www.example.*"}, false, false},
		{map[string]string{annotation: `~^www\d+\.example\.com$`}, []string{`~^www\d+\.example\.com$`}, false, false},
		{map[string]string{annotation: `www.xpto;lala`}, []string{}, false, true},
		{map[string]string{annotation: `www.xpto;lala`}, []string{"www.xpto;lala"}, true, false}, // When we skip validation no error should happen
		{map[string]string{annotation: ""}, []string{}, false, true},
		{map[string]string{}, []string{}, false, true},
		{nil, []string{}, false, true},
	}

	ing := &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{},
	}

	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		if testCase.skipValidation {
			parser.EnableAnnotationValidation = false
		}
		t.Cleanup(func() {
			parser.EnableAnnotationValidation = true
		})
		result, err := ap.Parse(ing)
		if (err != nil) != testCase.wantErr {
			t.Errorf("ParseAliasAnnotation() annotation: %s, error = %v, wantErr %v", testCase.annotations, err, testCase.wantErr)
		}
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("expected %v but returned %v, annotations: %s", testCase.expected, result, testCase.annotations)
		}
	}
}
