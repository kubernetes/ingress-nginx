/*
Copyright 2017 The Kubernetes Authors.

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

package sslcipher

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
	ap := NewParser(&resolver.Mock{})
	if ap == nil {
		t.Fatalf("expected a parser.IngressAnnotation but returned nil")
	}

	annotationSSLCiphers := parser.GetAnnotationWithPrefix(sslCipherAnnotation)
	annotationSSLPreferServerCiphers := parser.GetAnnotationWithPrefix(sslPreferServerCipherAnnotation)

	testCases := []struct {
		annotations map[string]string
		expected    Config
		expectErr   bool
	}{
		{map[string]string{annotationSSLCiphers: "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP"}, Config{"ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP", ""}, false},
		{map[string]string{annotationSSLCiphers: "ALL:!aNULL:!EXPORT56@SECLEVEL=2:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP"}, Config{"ALL:!aNULL:!EXPORT56@SECLEVEL=2:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP", ""}, false},
		{map[string]string{annotationSSLCiphers: "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP@STRENGTH"}, Config{"ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP@STRENGTH", ""}, false},
		{map[string]string{annotationSSLCiphers: "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP@STRENGTH@SECLEVEL=3"}, Config{"ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP@STRENGTH@SECLEVEL=3", ""}, false},
		{map[string]string{annotationSSLCiphers: "ALL:!aNULL:!EXPORT56:RC4+RSA@STRENGTH:+HIGH@SECLEVEL=5:+MEDIUM:+LOW:+SSLv2:+EXP"}, Config{"ALL:!aNULL:!EXPORT56:RC4+RSA@STRENGTH:+HIGH@SECLEVEL=5:+MEDIUM:+LOW:+SSLv2:+EXP", ""}, false},
		{
			map[string]string{annotationSSLCiphers: "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256"},
			Config{"ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256", ""},
			false,
		},
		{map[string]string{annotationSSLCiphers: ""}, Config{"", ""}, false},
		{map[string]string{annotationSSLPreferServerCiphers: "true"}, Config{"", "on"}, false},
		{map[string]string{annotationSSLPreferServerCiphers: "false"}, Config{"", "off"}, false},
		{map[string]string{annotationSSLCiphers: "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP", annotationSSLPreferServerCiphers: "true"}, Config{"ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP", "on"}, false},
		{map[string]string{annotationSSLCiphers: "ALL:SOMETHING:;locationXPTO"}, Config{"", ""}, true},
		{map[string]string{}, Config{"", ""}, false},
		{nil, Config{"", ""}, false},
	}

	ing := &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{},
	}

	for i, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)
		result, err := ap.Parse(ing)
		if (err != nil) != testCase.expectErr {
			t.Fatalf("expected error: %t got error: %t err value: %s. %+v", testCase.expectErr, err != nil, err, testCase.annotations)
		}
		if !reflect.DeepEqual(result, &testCases[i].expected) {
			t.Errorf("expected %v but returned %v, annotations: %s", testCase.expected, result, testCase.annotations)
		}
	}
}
