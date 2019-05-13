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
	"reflect"
	"sort"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
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

type mockBackend struct {
	resolver.Mock
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

func TestParseCIDRs(t *testing.T) {
	cidr, _ := parseCIDRs("invalid.com")
	if cidr != nil {
		t.Errorf("expected %v but got %v", nil, cidr)
	}

	expected := []string{"192.0.0.1", "192.0.1.0/24"}
	cidr, err := parseCIDRs("192.0.0.1, 192.0.1.0/24")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	sort.Strings(cidr)
	if !reflect.DeepEqual(expected, cidr) {
		t.Errorf("expected %v but got %v", expected, cidr)
	}
}

func TestRateLimiting(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("limit-connections")] = "0"
	data[parser.GetAnnotationWithPrefix("limit-rps")] = "0"
	data[parser.GetAnnotationWithPrefix("limit-rpm")] = "0"
	ing.SetAnnotations(data)

	_, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error with invalid limits (0)")
	}

	data = map[string]string{}
	data[parser.GetAnnotationWithPrefix("limit-connections")] = "5"
	data[parser.GetAnnotationWithPrefix("limit-rps")] = "100"
	data[parser.GetAnnotationWithPrefix("limit-rpm")] = "10"
	data[parser.GetAnnotationWithPrefix("limit-rate-after")] = "100"
	data[parser.GetAnnotationWithPrefix("limit-rate")] = "10"

	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	rateLimit, ok := i.(*Config)
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
