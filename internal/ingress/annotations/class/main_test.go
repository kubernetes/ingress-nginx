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
	"k8s.io/ingress-nginx/internal/k8s"
)

func TestIsValidClass(t *testing.T) {
	dc := DefaultClass
	ic := IngressClass
	k8sic := k8s.IngressClass
	v1Ready := k8s.IsIngressV1Ready
	// restore original values after the tests
	defer func() {
		DefaultClass = dc
		IngressClass = ic
		k8s.IngressClass = k8sic
		k8s.IsIngressV1Ready = v1Ready
	}()

	tests := []struct {
		ingress          string
		controller       string
		defClass         string
		annotation       bool
		ingressClassName bool
		k8sClass         *networking.IngressClass
		v1Ready          bool
		isValid          bool
	}{
		{"", "", "nginx", true, false, nil, false, true},
		{"", "nginx", "nginx", true, false, nil, false, true},
		{"nginx", "nginx", "nginx", true, false, nil, false, true},
		{"custom", "custom", "nginx", true, false, nil, false, true},
		{"", "killer", "nginx", true, false, nil, false, false},
		{"custom", "nginx", "nginx", true, false, nil, false, false},
		{"nginx", "nginx", "nginx", false, true, nil, false, true},
		{"custom", "nginx", "nginx", false, true, nil, true, false},
		{"nginx", "nginx", "nginx", false, true, nil, true, true},
		{"", "custom", "nginx", false, false,
			&networking.IngressClass{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "custom",
				},
			},
			false, false},
		{"", "custom", "nginx", false, false,
			&networking.IngressClass{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "custom",
				},
			},
			true, false},
	}

	for _, test := range tests {
		ing := &networking.Ingress{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "foo",
				Namespace: api.NamespaceDefault,
			},
		}

		data := map[string]string{}
		ing.SetAnnotations(data)
		if test.annotation {
			ing.Annotations[IngressKey] = test.ingress
		}
		if test.ingressClassName {
			ing.Spec.IngressClassName = &[]string{test.ingress}[0]
		}

		IngressClass = test.controller
		DefaultClass = test.defClass
		k8s.IngressClass = test.k8sClass
		k8s.IsIngressV1Ready = test.v1Ready

		b := IsValid(ing)
		if b != test.isValid {
			t.Errorf("test %v - expected %v but %v was returned", test, test.isValid, b)
		}
	}
}
