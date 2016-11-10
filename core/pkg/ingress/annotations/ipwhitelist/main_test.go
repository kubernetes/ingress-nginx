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

package ipwhitelist

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"

	"k8s.io/ingress/core/pkg/ingress/defaults"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

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
			Rules: []extensions.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Path:    "/foo",
									Backend: defaultBackend,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestParseAnnotations(t *testing.T) {
	ing := buildIngress()

	testNet := "10.0.0.0/24"
	enet := []string{testNet}

	data := map[string]string{}
	data[whitelist] = testNet
	ing.SetAnnotations(data)

	expected := &SourceRange{
		CIDR: enet,
	}

	sr, err := ParseAnnotations(defaults.Backend{}, ing)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(sr, expected) {
		t.Errorf("Expected %v but returned %s", sr, expected)
	}

	data[whitelist] = "www"
	ing.SetAnnotations(data)
	_, err = ParseAnnotations(defaults.Backend{}, ing)
	if err == nil {
		t.Errorf("Expected error parsing an invalid cidr")
	}

	delete(data, "whitelist")
	ing.SetAnnotations(data)
	//sr, _ = ParseAnnotations(defaults.Backend{}, ing)
	// TODO: fix test
	/*
		if !reflect.DeepEqual(sr.CIDR, []string{}) {
			t.Errorf("Expected empty CIDR but %v returned", sr.CIDR)
		}

		sr, _ = ParseAnnotations(defaults.Upstream{}, &extensions.Ingress{})
		if !reflect.DeepEqual(sr.CIDR, []string{}) {
			t.Errorf("Expected empty CIDR but %v returned", sr.CIDR)
		}
	*/
}
