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

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/ingress/core/pkg/ingress/errors"
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
	defaults.Backend
}

func (m mockBackend) GetDefaultBackend() defaults.Backend {
	return m.Backend
}

func TestParseAnnotations(t *testing.T) {
	// TODO: convert test cases to tables
	ing := buildIngress()

	testNet := "10.0.0.0/24"
	enet := []string{testNet}

	data := map[string]string{}
	data[whitelist] = testNet
	ing.SetAnnotations(data)

	expected := &SourceRange{
		CIDR: enet,
	}

	p := NewParser(mockBackend{})

	i, err := p.Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sr, ok := i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}

	if !reflect.DeepEqual(sr, expected) {
		t.Errorf("expected %v but returned %s", sr, expected)
	}

	data[whitelist] = "www"
	_, err = p.Parse(ing)
	if err == nil {
		t.Errorf("expected error parsing an invalid cidr")
	}

	if !errors.IsLocationDenied(err) {
		t.Errorf("expected LocationDenied error: %+v", err)
	}

	delete(data, whitelist)
	i, err = p.Parse(ing)

	if err != nil {
		t.Errorf("unexpected error when no annotation present: %v", err)
	}

	sr, ok = i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}
	if !strsEquals(sr.CIDR, []string{}) {
		t.Errorf("expected empty CIDR but %v returned", sr.CIDR)
	}

	i, _ = p.Parse(&extensions.Ingress{})
	sr, ok = i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}
	if !strsEquals(sr.CIDR, []string{}) {
		t.Errorf("expected empty CIDR but %v returned", sr.CIDR)
	}

	data[whitelist] = "2.2.2.2/32,1.1.1.1/32,3.3.3.0/24"
	i, _ = p.Parse(ing)
	sr, ok = i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}
	ecidr := []string{"1.1.1.1/32", "2.2.2.2/32", "3.3.3.0/24"}
	if !strsEquals(sr.CIDR, ecidr) {
		t.Errorf("Expected %v CIDR but %v returned", ecidr, sr.CIDR)
	}
}

// Test that when we have a whitelist set on the Backend that is used when we
// don't have the annotation
func TestParseAnnotationsWithDefaultConfig(t *testing.T) {
	// TODO: convert test cases to tables
	ing := buildIngress()

	mockBackend := mockBackend{}
	mockBackend.Backend.WhitelistSourceRange = []string{"4.4.4.0/24", "1.2.3.4/32"}
	testNet := "10.0.0.0/24"
	enet := []string{testNet}

	data := map[string]string{}
	data[whitelist] = testNet
	ing.SetAnnotations(data)

	expected := &SourceRange{
		CIDR: enet,
	}

	p := NewParser(mockBackend)

	i, err := p.Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sr, ok := i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}

	if !reflect.DeepEqual(sr, expected) {
		t.Errorf("expected %v but returned %s", sr, expected)
	}

	data[whitelist] = "www"
	_, err = p.Parse(ing)
	if err == nil {
		t.Errorf("expected error parsing an invalid cidr")
	}
	if !errors.IsLocationDenied(err) {
		t.Errorf("expected LocationDenied error: %+v", err)
	}

	delete(data, whitelist)
	i, err = p.Parse(ing)

	if err != nil {
		t.Errorf("unexpected error when no annotation present: %v", err)
	}

	sr, ok = i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}
	if !strsEquals(sr.CIDR, mockBackend.WhitelistSourceRange) {
		t.Errorf("expected fallback CIDR but %v returned", sr.CIDR)
	}

	i, _ = p.Parse(&extensions.Ingress{})
	sr, ok = i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}
	if !strsEquals(sr.CIDR, mockBackend.WhitelistSourceRange) {
		t.Errorf("expected fallback CIDR but %v returned", sr.CIDR)
	}

	data[whitelist] = "2.2.2.2/32,1.1.1.1/32,3.3.3.0/24"
	i, _ = p.Parse(ing)
	sr, ok = i.(*SourceRange)
	if !ok {
		t.Errorf("expected a SourceRange type")
	}
	ecidr := []string{"1.1.1.1/32", "2.2.2.2/32", "3.3.3.0/24"}
	if !strsEquals(sr.CIDR, ecidr) {
		t.Errorf("Expected %v CIDR but %v returned", ecidr, sr.CIDR)
	}
}

func strsEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
