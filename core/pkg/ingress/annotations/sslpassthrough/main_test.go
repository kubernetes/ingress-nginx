/*
Copyright 2016 The Kubernetes Authors.

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

package sslpassthrough

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"

	"k8s.io/ingress/core/pkg/ingress/defaults"
)

func buildIngress() *extensions.Ingress {
	return &extensions.Ingress{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
		},
	}
}

func TestParseAnnotations(t *testing.T) {
	ing := buildIngress()

	_, err := ParseAnnotations(defaults.Backend{}, ing)
	if err == nil {
		t.Errorf("unexpected error: %v", err)
	}

	data := map[string]string{}
	data[passthrough] = "true"
	ing.SetAnnotations(data)
	// test ingress using the annotation without a TLS section
	val, err := ParseAnnotations(defaults.Backend{}, ing)
	if err == nil {
		t.Errorf("expected error parsing an invalid cidr")
	}

	// test with a valid host
	ing.Spec.TLS = []extensions.IngressTLS{
		{
			Hosts: []string{"foo.bar.com"},
		},
	}
	val, err = ParseAnnotations(defaults.Backend{}, ing)
	if err != nil {
		t.Errorf("expected error parsing an invalid cidr")
	}
	if !val {
		t.Errorf("expected true but false returned")
	}
}
