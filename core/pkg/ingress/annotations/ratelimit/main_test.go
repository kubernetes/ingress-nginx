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

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
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

type mockBackend struct {
}

func (m mockBackend) GetDefaultBackend() defaults.Backend {
	return defaults.Backend{
		LimitRateAfter: 0,
		LimitRate:      0,
	}
}

func TestWithoutAnnotations(t *testing.T) {
	ing := buildIngress()
	_, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Error("unexpected error with ingress without annotations")
	}
}

func TestBadRateLimiting(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[limitIP] = "0"
	data[limitRPS] = "0"
	data[limitRPM] = "0"
	ing.SetAnnotations(data)

	_, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error with invalid limits (0)")
	}

	data = map[string]string{}
	data[limitIP] = "5"
	data[limitRPS] = "100"
	data[limitRPM] = "10"
	data[limitRATEAFTER] = "100"
	data[limitRATE] = "10"

	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
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
	if rateLimit.RPM.Limit != 10 {
		t.Errorf("expected 10 in limit by rpm but %v was returend", rateLimit.RPM)
	}
	if rateLimit.LimitRateAfter != 100 {
		t.Errorf("expected 100 in limit by limitrateafter but %v was returend", rateLimit.LimitRateAfter)
	}
	if rateLimit.LimitRate != 10 {
		t.Errorf("expected 10 in limit by limitrate but %v was returend", rateLimit.LimitRate)
	}
}
