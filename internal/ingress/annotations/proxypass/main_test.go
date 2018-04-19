/*
Copyright 2018 The Kubernetes Authors.

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

package proxypass

import (
	"testing"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func TestParse(t *testing.T) {
	proxyPassAddressAnnotation := parser.GetAnnotationWithPrefix("proxy-pass-address")
	proxyPassPortAnnotation := parser.GetAnnotationWithPrefix("proxy-pass-port")
	proxyToLocalNodeAnnotation := parser.GetAnnotationWithPrefix("proxy-to-local-node")

	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	testCases := []struct {
		annotations map[string]string
		expected    *Config
	}{
		{nil, &Config{}},
		{map[string]string{}, &Config{}},

		// Everything passed in
		{map[string]string{proxyPassAddressAnnotation: "localhost", proxyPassPortAnnotation: "4140", proxyToLocalNodeAnnotation: "true"}, &Config{Address: "localhost", Port: "4140", ProxyToLocalNode: true}},
		// Only address omitted
		{map[string]string{proxyPassPortAnnotation: "4140", proxyToLocalNodeAnnotation: "true"}, &Config{Address: "", Port: "4140", ProxyToLocalNode: true}},
		// Address and port omitted
		{map[string]string{proxyToLocalNodeAnnotation: "true"}, &Config{Address: "", Port: "", ProxyToLocalNode: true}},

		// Everything passed in, proxyToLocalNode explicitly false
		{map[string]string{proxyPassAddressAnnotation: "localhost", proxyPassPortAnnotation: "4140", proxyToLocalNodeAnnotation: "false"}, &Config{Address: "localhost", Port: "4140", ProxyToLocalNode: false}},
		// Address omitted, proxyToLocalNode explicitly false
		{map[string]string{proxyPassPortAnnotation: "4140", proxyToLocalNodeAnnotation: "false"}, &Config{Address: "", Port: "4140", ProxyToLocalNode: false}},
		// Address and port omitted, proxyToLocalNode explicitly false
		{map[string]string{proxyToLocalNodeAnnotation: "false"}, &Config{Address: "", Port: "", ProxyToLocalNode: false}},

		// Only proxyToLocalNode omitted
		{map[string]string{proxyPassAddressAnnotation: "localhost", proxyPassPortAnnotation: "4140"}, &Config{Address: "localhost", Port: "4140", ProxyToLocalNode: false}},

		// Only address passed in
		{map[string]string{proxyPassAddressAnnotation: "localhost"}, &Config{Address: "localhost", Port: "", ProxyToLocalNode: false}},
		// Only port passed in
		{map[string]string{proxyPassPortAnnotation: "4140"}, &Config{Address: "", Port: "4140", ProxyToLocalNode: false}},
	}

	ing := &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{},
	}

	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		result, _ := ap.Parse(ing)
		config := result.(*Config)
		if !config.Equal(testCase.expected) {
			t.Errorf("expected %v but returned %v, annotations: %s", testCase.expected, result, testCase.annotations)
		}
	}
}
