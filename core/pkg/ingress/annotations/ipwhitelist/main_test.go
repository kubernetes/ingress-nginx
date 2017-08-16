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
	defaults.Backend
}

func (m mockBackend) GetDefaultBackend() defaults.Backend {
	return m.Backend
}

func TestParseAnnotations(t *testing.T) {
	ing := buildIngress()
	tests := map[string]struct {
		net        string
		expectCidr []string
		expectErr  bool
		errOut     string
	}{
		"test parse a valid net": {
			net:        "10.0.0.0/24",
			expectCidr: []string{"10.0.0.0/24"},
			expectErr:  false,
		},
		"test parse a invalid net": {
			net:       "ww",
			expectErr: true,
			errOut:    "the annotation does not contain a valid IP address or network: invalid CIDR address: ww",
		},
		"test parse a empty net": {
			net:       "",
			expectErr: true,
			errOut:    "the annotation does not contain a valid IP address or network: invalid CIDR address: ",
		},
		"test parse multiple valid cidr": {
			net:        "2.2.2.2/32,1.1.1.1/32,3.3.3.0/24",
			expectCidr: []string{"1.1.1.1/32", "2.2.2.2/32", "3.3.3.0/24"},
			expectErr:  false,
		},
	}

	for testName, test := range tests {
		data := map[string]string{}
		data[whitelist] = test.net
		ing.SetAnnotations(data)
		p := NewParser(mockBackend{})
		i, err := p.Parse(ing)
		if err != nil && !test.expectErr {
			t.Errorf("%v:unexpected error: %v", testName, err)
		}
		if test.expectErr {
			if err.Error() != test.errOut {
				t.Errorf("%v:expected error: %v but %v return", testName, test.errOut, err.Error())
			}
		}
		if !test.expectErr {
			sr, ok := i.(*SourceRange)
			if !ok {
				t.Errorf("%v:expected a SourceRange type", testName)
			}
			if !strsEquals(sr.CIDR, test.expectCidr) {
				t.Errorf("%v:expected %v CIDR but %v returned", testName, test.expectCidr, sr.CIDR)
			}
		}
	}
}

// Test that when we have a whitelist set on the Backend that is used when we
// don't have the annotation
func TestParseAnnotationsWithDefaultConfig(t *testing.T) {
	ing := buildIngress()
	mockBackend := mockBackend{}
	mockBackend.Backend.WhitelistSourceRange = []string{"4.4.4.0/24", "1.2.3.4/32"}
	tests := map[string]struct {
		net        string
		expectCidr []string
		expectErr  bool
		errOut     string
	}{
		"test parse a valid net": {
			net:        "10.0.0.0/24",
			expectCidr: []string{"10.0.0.0/24"},
			expectErr:  false,
		},
		"test parse a invalid net": {
			net:       "ww",
			expectErr: true,
			errOut:    "the annotation does not contain a valid IP address or network: invalid CIDR address: ww",
		},
		"test parse a empty net": {
			net:       "",
			expectErr: true,
			errOut:    "the annotation does not contain a valid IP address or network: invalid CIDR address: ",
		},
		"test parse multiple valid cidr": {
			net:        "2.2.2.2/32,1.1.1.1/32,3.3.3.0/24",
			expectCidr: []string{"1.1.1.1/32", "2.2.2.2/32", "3.3.3.0/24"},
			expectErr:  false,
		},
	}

	for testName, test := range tests {
		data := map[string]string{}
		data[whitelist] = test.net
		ing.SetAnnotations(data)
		p := NewParser(mockBackend)
		i, err := p.Parse(ing)
		if err != nil && !test.expectErr {
			t.Errorf("%v:unexpected error: %v", testName, err)
		}
		if test.expectErr {
			if err.Error() != test.errOut {
				t.Errorf("%v:expected error: %v but %v return", testName, test.errOut, err.Error())
			}
		}
		if !test.expectErr {
			sr, ok := i.(*SourceRange)
			if !ok {
				t.Errorf("%v:expected a SourceRange type", testName)
			}
			if !strsEquals(sr.CIDR, test.expectCidr) {
				t.Errorf("%v:expected %v CIDR but %v returned", testName, test.expectCidr, sr.CIDR)
			}
		}
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
