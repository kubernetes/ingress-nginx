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

package globalratelimit

import (
	"encoding/json"
	"fmt"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	UID         = "31285d47-b150-4dcf-bd6f-12c46d769f6e"
	expectedUID = "31285d47b1504dcfbd6f12c46d769f6e"
)

func buildIngress() *networking.Ingress {
	defaultBackend := networking.IngressBackend{
		Service: &networking.IngressServiceBackend{
			Name: "default-backend",
			Port: networking.ServiceBackendPort{
				Number: 80,
			},
		},
	}

	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
			UID:       UID,
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-backend",
					Port: networking.ServiceBackendPort{
						Number: 80,
					},
				},
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

func TestGlobalRateLimiting(t *testing.T) {
	ing := buildIngress()

	annRateLimit := parser.GetAnnotationWithPrefix("global-rate-limit")
	annRateLimitWindow := parser.GetAnnotationWithPrefix("global-rate-limit-window")
	annRateLimitKey := parser.GetAnnotationWithPrefix("global-rate-limit-key")
	annRateLimitIgnoredCIDRs := parser.GetAnnotationWithPrefix("global-rate-limit-ignored-cidrs")

	testCases := []struct {
		title          string
		annotations    map[string]string
		expectedConfig *Config
		expectedErr    error
	}{
		{
			"no annotation",
			nil,
			&Config{},
			nil,
		},
		{
			"minimum required annotations",
			map[string]string{
				annRateLimit:       "100",
				annRateLimitWindow: "2m",
			},
			&Config{
				Namespace:    expectedUID,
				Limit:        100,
				WindowSize:   120,
				Key:          "$remote_addr",
				IgnoredCIDRs: make([]string, 0),
			},
			nil,
		},
		{
			"global-rate-limit-key annotation",
			map[string]string{
				annRateLimit:       "100",
				annRateLimitWindow: "2m",
				annRateLimitKey:    "$http_x_api_user",
			},
			&Config{
				Namespace:    expectedUID,
				Limit:        100,
				WindowSize:   120,
				Key:          "$http_x_api_user",
				IgnoredCIDRs: make([]string, 0),
			},
			nil,
		},
		{
			"global-rate-limit-ignored-cidrs annotation",
			map[string]string{
				annRateLimit:             "100",
				annRateLimitWindow:       "2m",
				annRateLimitKey:          "$http_x_api_user",
				annRateLimitIgnoredCIDRs: "127.0.0.1, 200.200.24.0/24",
			},
			&Config{
				Namespace:    expectedUID,
				Limit:        100,
				WindowSize:   120,
				Key:          "$http_x_api_user",
				IgnoredCIDRs: []string{"127.0.0.1", "200.200.24.0/24"},
			},
			nil,
		},
		{
			"global-rate-limit-complex-key",
			map[string]string{
				annRateLimit:       "100",
				annRateLimitWindow: "2m",
				annRateLimitKey:    "${http_x_api_user}${otherinfo}",
			},
			&Config{
				Namespace:    expectedUID,
				Limit:        100,
				WindowSize:   120,
				Key:          "${http_x_api_user}${otherinfo}",
				IgnoredCIDRs: make([]string, 0),
			},
			nil,
		},
		{
			"incorrect duration for window",
			map[string]string{
				annRateLimit:       "100",
				annRateLimitWindow: "2mb",
				annRateLimitKey:    "$http_x_api_user",
			},
			&Config{},
			ing_errors.ValidationError{
				Reason: fmt.Errorf("failed to parse 'global-rate-limit-window' value: annotation nginx.ingress.kubernetes.io/global-rate-limit-window contains invalid value"),
			},
		},
	}

	for _, testCase := range testCases {
		ing.SetAnnotations(testCase.annotations)

		i, actualErr := NewParser(mockBackend{}).Parse(ing)
		if (testCase.expectedErr == nil || actualErr == nil) && testCase.expectedErr != actualErr {
			t.Errorf("%s expected error '%v' but got '%v'", testCase.title, testCase.expectedErr, actualErr)
		} else if testCase.expectedErr != nil && actualErr != nil &&
			testCase.expectedErr.Error() != actualErr.Error() {
			t.Errorf("expected error '%v' but got '%v'", testCase.expectedErr, actualErr)
		}

		actualConfig, ok := i.(*Config)
		if !ok {
			t.Errorf("expected Config type but got %T", i)
		}
		if !testCase.expectedConfig.Equal(actualConfig) {
			expectedJSON, err := json.Marshal(testCase.expectedConfig)
			if err != nil {
				t.Errorf("failed to marshal expected config: %v", err)
			}
			actualJSON, err := json.Marshal(actualConfig)
			if err != nil {
				t.Errorf("failed to marshal actual config: %v", err)
			}
			t.Errorf("%v: expected config '%s' but got '%s'", testCase.title, expectedJSON, actualJSON)
		}
	}
}
