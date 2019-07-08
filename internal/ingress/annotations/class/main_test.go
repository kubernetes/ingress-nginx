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

package class

import (
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsValidClass(t *testing.T) {
	dc := DefaultClass
	ic := IngressClass
	// restore original values after the tests
	defer func() {
		DefaultClass = dc
		IngressClass = ic
	}()

	tests := []struct {
		ingress    string
		controller string
		defClass   string
		isValid    bool
	}{
		{"", "", "nginx", true},
		{"", "nginx", "nginx", true},
		{"nginx", "nginx", "nginx", true},
		{"custom", "custom", "nginx", true},
		{"", "killer", "nginx", false},
		{"custom", "nginx", "nginx", false},
	}

	ing := &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
	}

	data := map[string]string{}
	ing.SetAnnotations(data)
	for _, test := range tests {
		ing.Annotations[IngressKey] = test.ingress

		IngressClass = test.controller
		DefaultClass = test.defClass

		b := IsValid(ing)
		if b != test.isValid {
			t.Errorf("test %v - expected %v but %v was returned", test, test.isValid, b)
		}
	}
}
