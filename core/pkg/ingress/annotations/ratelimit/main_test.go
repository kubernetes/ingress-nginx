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

package ratelimit

import (
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

func TestWithoutAnnotations(t *testing.T) {
	ing := buildIngress()
	_, err := NewParser().Parse(ing)
	if err != nil {
		t.Error("unexpected error with ingress without annotations")
	}
}

func TestBadRateLimiting(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[limitIP] = "0"
	data[limitRPS] = "0"
	ing.SetAnnotations(data)

	_, err := NewParser().Parse(ing)
	if err != nil {
		t.Errorf("unexpected error with invalid limits (0)")
	}

	data = map[string]string{}
	data[limitIP] = "5"
	data[limitRPS] = "100"
	ing.SetAnnotations(data)

	i, err := NewParser().Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	rateLimit, ok := i.(*RateLimit)
	if !ok {
		t.Errorf("expected a RateLimit type")
	}
	if rateLimit.Connections.Limit != 5 {
		t.Errorf("expected 5 in limit by ip but %v was returend", rateLimit.Connections)
	}
	if rateLimit.RPS.Limit != 100 {
		t.Errorf("expected 100 in limit by rps but %v was returend", rateLimit.RPS)
	}
}
