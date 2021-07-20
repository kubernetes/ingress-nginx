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

package modsecurity

import (
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func TestParse(t *testing.T) {
	enable := parser.GetAnnotationWithPrefix("enable-modsecurity")
	owasp := parser.GetAnnotationWithPrefix("enable-owasp-core-rules")
	transID := parser.GetAnnotationWithPrefix("modsecurity-transaction-id")
	snippet := parser.GetAnnotationWithPrefix("modsecurity-snippet")

	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	testCases := []struct {
		annotations map[string]string
		expected    Config
	}{
		{map[string]string{enable: "true"}, Config{true, true, false, "", ""}},
		{map[string]string{enable: "false"}, Config{false, true, false, "", ""}},
		{map[string]string{enable: ""}, Config{false, false, false, "", ""}},

		{map[string]string{owasp: "true"}, Config{false, false, true, "", ""}},
		{map[string]string{owasp: "false"}, Config{false, false, false, "", ""}},
		{map[string]string{owasp: ""}, Config{false, false, false, "", ""}},

		{map[string]string{transID: "ok"}, Config{false, false, false, "ok", ""}},
		{map[string]string{transID: ""}, Config{false, false, false, "", ""}},

		{map[string]string{snippet: "ModSecurity Rule"}, Config{false, false, false, "", "ModSecurity Rule"}},
		{map[string]string{snippet: ""}, Config{false, false, false, "", ""}},

		{map[string]string{}, Config{false, false, false, "", ""}},
		{nil, Config{false, false, false, "", ""}},
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
		result, _ := ap.Parse(ing)
		config := result.(*Config)
		if !config.Equal(&testCase.expected) {
			t.Errorf("expected %v but returned %v, annotations: %s", testCase.expected, result, testCase.annotations)
		}
	}
}
