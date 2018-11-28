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

package proxy

import (
	"testing"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
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
	resolver.Mock
}

func TestProxy(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("proxy-connect-timeout")] = "1"
	data[parser.GetAnnotationWithPrefix("proxy-send-timeout")] = "2"
	data[parser.GetAnnotationWithPrefix("proxy-read-timeout")] = "3"
	data[parser.GetAnnotationWithPrefix("proxy-buffer-size")] = "1k"
	data[parser.GetAnnotationWithPrefix("proxy-body-size")] = "2k"
	data[parser.GetAnnotationWithPrefix("proxy-next-upstream")] = "off"
	data[parser.GetAnnotationWithPrefix("proxy-next-upstream-tries")] = "3"
	data[parser.GetAnnotationWithPrefix("proxy-request-buffering")] = "off"
	data[parser.GetAnnotationWithPrefix("proxy-buffering")] = "on"
	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Fatalf("unexpected error parsing a valid")
	}
	p, ok := i.(*Config)
	if !ok {
		t.Fatalf("expected a Config type")
	}
	if p.ConnectTimeout != 1 {
		t.Errorf("expected 1 as connect-timeout but returned %v", p.ConnectTimeout)
	}
	if p.SendTimeout != 2 {
		t.Errorf("expected 2 as send-timeout but returned %v", p.SendTimeout)
	}
	if p.ReadTimeout != 3 {
		t.Errorf("expected 3 as read-timeout but returned %v", p.ReadTimeout)
	}
	if p.BufferSize != "1k" {
		t.Errorf("expected 1k as buffer-size but returned %v", p.BufferSize)
	}
	if p.BodySize != "2k" {
		t.Errorf("expected 2k as body-size but returned %v", p.BodySize)
	}
	if p.NextUpstream != "off" {
		t.Errorf("expected off as next-upstream but returned %v", p.NextUpstream)
	}
	if p.NextUpstreamTries != 3 {
		t.Errorf("expected 3 as next-upstream-tries but returned %v", p.NextUpstreamTries)
	}
	if p.RequestBuffering != "off" {
		t.Errorf("expected off as request-buffering but returned %v", p.RequestBuffering)
	}
	if p.ProxyBuffering != "on" {
		t.Errorf("expected on as proxy-buffering but returned %v", p.ProxyBuffering)
	}
}

func TestProxyWithNoAnnotation(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	i, err := NewParser(mockBackend{}).Parse(ing)
	if err != nil {
		t.Fatalf("unexpected error parsing a valid")
	}
	p, ok := i.(*Config)
	if !ok {
		t.Fatalf("expected a Config type")
	}
	if p.ConnectTimeout != 5 {
		t.Errorf("expected 5 as connect-timeout but returned %v", p.ConnectTimeout)
	}
	if p.SendTimeout != 60 {
		t.Errorf("expected 60 as send-timeout but returned %v", p.SendTimeout)
	}
	if p.ReadTimeout != 60 {
		t.Errorf("expected 60 as read-timeout but returned %v", p.ReadTimeout)
	}
	if p.BufferSize != "4k" {
		t.Errorf("expected 4k as buffer-size but returned %v", p.BufferSize)
	}
	if p.BodySize != "1m" {
		t.Errorf("expected 1m as body-size but returned %v", p.BodySize)
	}
	if p.NextUpstream != "error timeout" {
		t.Errorf("expected error timeout as next-upstream but returned %v", p.NextUpstream)
	}
	if p.NextUpstreamTries != 3 {
		t.Errorf("expected 3 as next-upstream-tries but returned %v", p.NextUpstreamTries)
	}
	if p.RequestBuffering != "on" {
		t.Errorf("expected on as request-buffering but returned %v", p.RequestBuffering)
	}
}
