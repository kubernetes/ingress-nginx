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

package controller

import (
	"testing"

	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress"
)

func TestExtractTLSSecretName(t *testing.T) {
	tests := []struct {
		host    string
		ingress *extensions.Ingress
		fn      func(string) (*ingress.SSLCert, error)
		expName string
	}{
		{
			"foo.bar",
			nil,
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"",
		},
		{
			"foo.bar",
			&extensions.Ingress{},
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"",
		},
		{
			"foo.bar",
			&extensions.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: extensions.IngressSpec{
					TLS: []extensions.IngressTLS{
						{SecretName: "demo"},
					},
					Rules: []extensions.IngressRule{
						{
							Host: "foo.bar",
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"",
		},
		{
			"foo.bar",
			&extensions.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: extensions.IngressSpec{
					TLS: []extensions.IngressTLS{
						{SecretName: "demo"},
					},
					Rules: []extensions.IngressRule{
						{
							Host: "foo.bar",
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return &ingress.SSLCert{
					CN: []string{"foo.bar", "example.com"},
				}, nil
			},
			"demo",
		},
		{
			"foo.bar",
			&extensions.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: extensions.IngressSpec{
					TLS: []extensions.IngressTLS{
						{
							Hosts:      []string{"foo.bar", "example.com"},
							SecretName: "demo",
						},
					},
					Rules: []extensions.IngressRule{
						{
							Host: "foo.bar",
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return &ingress.SSLCert{
					CN: []string{"foo.bar", "example.com"},
				}, nil
			},
			"demo",
		},
	}

	for _, testCase := range tests {
		name := extractTLSSecretName(testCase.host, testCase.ingress, testCase.fn)
		if name != testCase.expName {
			t.Errorf("expected %v as the name of the secret but got %v", testCase.expName, name)
		}
	}
}
