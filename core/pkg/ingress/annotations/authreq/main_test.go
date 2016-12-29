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

package authreq

import (
	"fmt"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"
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

func TestAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	tests := []struct {
		title    string
		url      string
		method   string
		sendBody bool
		expErr   bool
	}{
		{"empty", "", "", false, true},
		{"no scheme", "bar", "", false, true},
		{"invalid host", "http://", "", false, true},
		{"invalid host (multiple dots)", "http://foo..bar.com", "", false, true},
		{"valid URL", "http://bar.foo.com/external-auth", "", false, false},
		{"valid URL - send body", "http://foo.com/external-auth", "POST", true, false},
		{"valid URL - send body", "http://foo.com/external-auth", "GET", true, false},
	}

	for _, test := range tests {
		data[authURL] = test.url
		data[authBody] = fmt.Sprintf("%v", test.sendBody)
		data[authMethod] = fmt.Sprintf("%v", test.method)

		i, err := NewParser().Parse(ing)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but retuned nil", test.title)
			}
			continue
		}
		u, ok := i.(*External)
		if !ok {
			t.Errorf("%v: expected an External type", test.title)
		}
		if u.URL != test.url {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.url, u.URL)
		}
		if u.Method != test.method {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.method, u.Method)
		}
		if u.SendBody != test.sendBody {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.sendBody, u.SendBody)
		}
	}
}
