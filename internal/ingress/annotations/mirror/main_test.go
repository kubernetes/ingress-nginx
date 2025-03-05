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

package mirror

import (
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func TestParse(t *testing.T) {
	requestBody := parser.GetAnnotationWithPrefix("mirror-request-body")
	backendURL := parser.GetAnnotationWithPrefix("mirror-target")
	host := parser.GetAnnotationWithPrefix("mirror-host")

	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	ngxURI := "/_mirror-c89a5111-b2e9-4af8-be19-c2a4a924c256"
	testCases := []struct {
		annotations map[string]string
		expected    *Config
	}{
		{map[string]string{backendURL: "https://test.env.com/$request_uri"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "https://test.env.com/$request_uri",
			Host:        "test.env.com",
		}},
		{map[string]string{backendURL: "https://test.env.com$request_uri"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "https://test.env.com$request_uri",
			Host:        "test.env.com",
		}},
		{map[string]string{backendURL: "https://test.env.com:8080$request_uri"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "https://test.env.com:8080$request_uri",
			Host:        "test.env.com",
		}},
		{map[string]string{backendURL: "https://test.env.com:8080/$request_uri"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "https://test.env.com:8080/$request_uri",
			Host:        "test.env.com",
		}},
		{map[string]string{requestBody: "off"}, &Config{
			Source:      "",
			RequestBody: "off",
			Target:      "",
			Host:        "",
		}},
		{map[string]string{host: "test.env.com", backendURL: "http://some.test.env.com/$someparam"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "http://some.test.env.com/$someparam",
			Host:        "test.env.com",
		}},
		{map[string]string{backendURL: "IamNotAURL"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "IamNotAURL",
			Host:        "",
		}},
		{map[string]string{backendURL: "http://some.test.env.com:2121/$someparam=1&$someotherparam=2"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "http://some.test.env.com:2121/$someparam=1&$someotherparam=2",
			Host:        "some.test.env.com",
		}},
		{map[string]string{backendURL: "http://some.test.env.com", host: "someInvalidParam.%^&*()_=!@#'\""}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "http://some.test.env.com",
			Host:        "some.test.env.com",
		}},
		{map[string]string{backendURL: "http://some.test.env.com", host: "_sbrubles-i\"@xpto:12345"}, &Config{
			Source:      ngxURI,
			RequestBody: "on",
			Target:      "http://some.test.env.com",
			Host:        "some.test.env.com",
		}},
	}

	ing := &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
			UID:       "c89a5111-b2e9-4af8-be19-c2a4a924c256",
		},
		Spec: networking.IngressSpec{},
	}

	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		result, err := ap.Parse(ing)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("expected %+v but returned %+v, annotations: %s", testCase.expected, result, testCase.annotations)
		}
	}
}
