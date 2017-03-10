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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func TestIsValidClass(t *testing.T) {
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
		{"", "", "nginx", true},
		{"custom", "nginx", "nginx", false},
	}

	ing := &extensions.Ingress{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
	}

	data := map[string]string{}
	ing.SetAnnotations(data)
	for _, test := range tests {
		ing.Annotations[IngressKey] = test.ingress
		b := IsValid(ing, test.controller, test.defClass)
		if b != test.isValid {
			t.Errorf("test %v - expected %v but %v was returned", test, test.isValid, b)
		}
	}
}
