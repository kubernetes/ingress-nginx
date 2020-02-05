/*
Copyright 2020 The Kubernetes Authors.

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

package httpport

import (
	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"testing"
)

func TestParse(t *testing.T) {

	testCases := map[string]struct {
		Ingress        *networking.Ingress
		ExpectedResult *Config
	}{
		"Neither HTTP not HTTPS port is defined": {
			Ingress: &networking.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			ExpectedResult: &Config{
				HTTPPort:  0,
				HTTPSPort: 0,
			},
		},
		"Only HTTP port is defined": {
			Ingress: &networking.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"nginx.ingress.kubernetes.io/http-port": "8080",
					},
				},
			},
			ExpectedResult: &Config{
				HTTPPort:  8080,
				HTTPSPort: 0,
			},
		},
		"Only HTTPS port is defined": {
			Ingress: &networking.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"nginx.ingress.kubernetes.io/https-port": "9443",
					},
				},
			},
			ExpectedResult: &Config{
				HTTPPort:  0,
				HTTPSPort: 9443,
			},
		},
		"Both HTTP and HTTPS ports are defined": {
			Ingress: &networking.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"nginx.ingress.kubernetes.io/http-port":  "8888",
						"nginx.ingress.kubernetes.io/https-port": "15443",
					},
				},
			},
			ExpectedResult: &Config{
				HTTPPort:  8888,
				HTTPSPort: 15443,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ap := NewParser(&resolver.Mock{})
			if ap == nil {
				t.Fatalf("expected a parser.IngressAnnotation but returned nil")
			}
			cfg, err := ap.Parse(tc.Ingress)
			assert.NoError(t, err)
			assert.Equal(t, tc.ExpectedResult, cfg.(*Config))
		})
	}
}

func TestEqual(t *testing.T) {

	testCases := map[string]struct {
		httpConfig1    *Config
		httpConfig2    *Config
		ExpectedResult bool
	}{
		"Both configs are nil": {
			ExpectedResult: true,
		},
		"Config1 is nil": {
			httpConfig2:    &Config{},
			ExpectedResult: false,
		},
		"Config2 is nil": {
			httpConfig1:    &Config{},
			ExpectedResult: false,
		},
		"HTTP Ports are not equal": {
			httpConfig1: &Config{
				HTTPPort: 80,
			},
			httpConfig2: &Config{
				HTTPPort: 8080,
			},
			ExpectedResult: false,
		},
		"HTTPS Ports are not equal": {
			httpConfig1: &Config{
				HTTPSPort: 443,
			},
			httpConfig2: &Config{
				HTTPSPort: 9443,
			},
			ExpectedResult: false,
		},
		"HTTP Ports are equal": {
			httpConfig1: &Config{
				HTTPPort: 90,
			},
			httpConfig2: &Config{
				HTTPPort: 90,
			},
			ExpectedResult: true,
		},
		"HTTPS Ports are equal": {
			httpConfig1: &Config{
				HTTPSPort: 8443,
			},
			httpConfig2: &Config{
				HTTPSPort: 8443,
			},
			ExpectedResult: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedResult, tc.httpConfig1.Equal(tc.httpConfig2))
		})
	}
}
