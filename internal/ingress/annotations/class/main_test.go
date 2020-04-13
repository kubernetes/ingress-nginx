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
	// restore original value after the tests
	ic := IngressClass
	defer func() {
		IngressClass = ic
	}()

	tests := []struct {
		className    string
		ingressClass string
		isValid      bool
	}{
		{"", "", true},
		{"", "nginx", true},
		{"nginx", "nginx", true},
		{"custom", "custom", true},
		{"", "killer", false},
		{"custom", "nginx", false},
	}

	ing := &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
	}

	for _, test := range tests {
		if test.className != "" {
			ing.Spec.IngressClassName = &test.className
		}

		IngressClass = test.ingressClass

		b := IsValid(ing)
		if b != test.isValid {
			t.Errorf("test %v - expected %v but %v was returned", test, test.isValid, b)
		}
	}
}
