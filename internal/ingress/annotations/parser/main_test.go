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

package parser

import (
	"net/url"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildIngress() *networking.Ingress {
	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{},
	}
}

func TestGetBoolAnnotation(t *testing.T) {
	ing := buildIngress()

	_, err := GetBoolAnnotation("", nil)
	if err == nil {
		t.Errorf("expected error but retuned nil")
	}

	tests := []struct {
		name   string
		field  string
		value  string
		exp    bool
		expErr bool
	}{
		{"valid - false", "bool", "false", false, false},
		{"valid - true", "bool", "true", true, false},
	}

	data := map[string]string{}
	ing.SetAnnotations(data)

	for _, test := range tests {
		data[GetAnnotationWithPrefix(test.field)] = test.value

		u, err := GetBoolAnnotation(test.field, ing)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but retuned nil", test.name)
			}
			continue
		}
		if u != test.exp {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.name, test.exp, u)
		}

		delete(data, test.field)
	}
}

func TestGetStringAnnotation(t *testing.T) {
	ing := buildIngress()

	_, err := GetStringAnnotation("", nil)
	if err == nil {
		t.Errorf("expected error but none returned")
	}

	tests := []struct {
		name   string
		field  string
		value  string
		exp    string
		expErr bool
	}{
		{"valid - A", "string", "A ", "A", false},
		{"valid - B", "string", "	B", "B", false},
		{"empty", "string", " ", "", true},
		{"valid multiline", "string", `
		rewrite (?i)/arcgis/rest/services/Utilities/Geometry/GeometryServer(.*)$ /arcgis/rest/services/Utilities/Geometry/GeometryServer$1 break;
		rewrite (?i)/arcgis/services/Utilities/Geometry/GeometryServer(.*)$ /arcgis/services/Utilities/Geometry/GeometryServer$1 break;
		`, `
rewrite (?i)/arcgis/rest/services/Utilities/Geometry/GeometryServer(.*)$ /arcgis/rest/services/Utilities/Geometry/GeometryServer$1 break;
rewrite (?i)/arcgis/services/Utilities/Geometry/GeometryServer(.*)$ /arcgis/services/Utilities/Geometry/GeometryServer$1 break;
`,
			false},
	}

	data := map[string]string{}
	ing.SetAnnotations(data)

	for _, test := range tests {
		data[GetAnnotationWithPrefix(test.field)] = test.value

		s, err := GetStringAnnotation(test.field, ing)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but none returned", test.name)
			}
			continue
		}
		if !test.expErr {
			if err != nil {
				t.Errorf("%v: didn't expected error but error was returned: %v", test.name, err)
			}
			continue
		}
		if s != test.exp {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.name, test.exp, s)
		}

		delete(data, test.field)
	}
}

func TestGetFloatAnnotation(t *testing.T) {
	ing := buildIngress()

	_, err := GetFloatAnnotation("", nil)
	if err == nil {
		t.Errorf("expected error but retuned nil")
	}

	tests := []struct {
		name   string
		field  string
		value  string
		exp    float32
		expErr bool
	}{
		{"valid - A", "string", "1.5", 1.5, false},
		{"valid - B", "string", "2", 2, false},
		{"valid - C", "string", "100.0", 100, false},
	}

	data := map[string]string{}
	ing.SetAnnotations(data)

	for _, test := range tests {
		data[GetAnnotationWithPrefix(test.field)] = test.value

		s, err := GetFloatAnnotation(test.field, ing)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but retuned nil", test.name)
			}
			continue
		}
		if s != test.exp {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.name, test.exp, s)
		}

		delete(data, test.field)
	}
}

func TestGetIntAnnotation(t *testing.T) {
	ing := buildIngress()

	_, err := GetIntAnnotation("", nil)
	if err == nil {
		t.Errorf("expected error but retuned nil")
	}

	tests := []struct {
		name   string
		field  string
		value  string
		exp    int
		expErr bool
	}{
		{"valid - A", "string", "1", 1, false},
		{"valid - B", "string", "2", 2, false},
	}

	data := map[string]string{}
	ing.SetAnnotations(data)

	for _, test := range tests {
		data[GetAnnotationWithPrefix(test.field)] = test.value

		s, err := GetIntAnnotation(test.field, ing)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but retuned nil", test.name)
			}
			continue
		}
		if s != test.exp {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.name, test.exp, s)
		}

		delete(data, test.field)
	}
}

func TestStringToURL(t *testing.T) {
	validURL := "http://bar.foo.com/external-auth"
	validParsedURL, _ := url.Parse(validURL)

	tests := []struct {
		title   string
		url     string
		message string
		parsed  *url.URL
		expErr  bool
	}{
		{"empty", "", "url scheme is empty", nil, true},
		{"no scheme", "bar", "url scheme is empty", nil, true},
		{"invalid host", "http://", "url host is empty", nil, true},
		{"invalid host (multiple dots)", "http://foo..bar.com", "invalid url host", nil, true},
		{"valid URL", validURL, "", validParsedURL, false},
	}

	for _, test := range tests {
		i, err := StringToURL(test.url)
		if test.expErr {
			if err == nil {
				t.Fatalf("%v: expected error but none returned", test.title)
			}

			if err.Error() != test.message {
				t.Errorf("%v: expected error \"%v\" but \"%v\" was returned", test.title, test.message, err)
			}
			continue
		}

		if i.String() != test.parsed.String() {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.parsed, i)
		}
	}
}
