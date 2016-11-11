/*
Copyright 2015 The Kubernetes Authors.

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

package controller

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func TestIsValidClass(t *testing.T) {
	ing := &extensions.Ingress{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
	}

	b := IsValidClass(ing, "")
	if !b {
		t.Error("Expected a valid class (missing annotation)")
	}

	data := map[string]string{}
	data[ingressClassKey] = "custom"
	ing.SetAnnotations(data)
	b = IsValidClass(ing, "custom")
	if !b {
		t.Errorf("Expected valid class but %v returned", b)
	}
	b = IsValidClass(ing, "nginx")
	if b {
		t.Errorf("Expected invalid class but %v returned", b)
	}
}
