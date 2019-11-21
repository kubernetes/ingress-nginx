/*
Copyright 2019 The Kubernetes Authors.

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

package certificatewithouthostnameverification

import (
	"fmt"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
	defaultBackend := networking.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{
			Backend: &networking.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
			Rules: []networking.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
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

func TestCertificateWithoutHostnameVerification(t *testing.T) {
	tests := []struct {
		title                                  string
		certificateWithoutHostnameVerification *bool
		exp                                    bool
	}{
		{"false", newFalse(), false},
		{"true", newTrue(), true},
		{"no annotation (expect default to false because previous behavior was to verify hostname)", nil, false},
	}

	for _, test := range tests {
		ing := buildIngress()

		data := map[string]string{}
		if test.certificateWithoutHostnameVerification != nil {
			data[parser.GetAnnotationWithPrefix("use-certificate-without-hostname-verification")] = fmt.Sprintf("%v", *test.certificateWithoutHostnameVerification)
		}
		ing.SetAnnotations(data)

		i, err := NewParser(resolver.Mock{}).Parse(ing)
		if err != nil {
			t.Errorf("unexpected error parsing a valid annotation: %v", err)
		}
		p, ok := i.(bool)
		if !ok {
			t.Errorf("expected a bool type")
		}

		if p != test.exp {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.exp, p)
		}
	}
}

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}
