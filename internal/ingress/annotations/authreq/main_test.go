/*
Copyright 2015 The Kubernetes Authors.

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

package authreq

import (
	"fmt"
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
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

func boolToString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func TestAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	tests := []struct {
		title                  string
		url                    string
		signinURL              string
		signinURLRedirectParam string
		method                 string
		requestRedirect        string
		authSnippet            string
		authCacheKey           string
		authAlwaysSetCookie    bool
		expErr                 bool
	}{
		{"empty", "", "", "", "", "", "", "", false, true},
		{"no scheme", "bar", "bar", "", "", "", "", "", false, true},
		{"invalid host", "http://", "http://", "", "", "", "", "", false, true},
		{"invalid host (multiple dots)", "http://foo..bar.com", "http://foo..bar.com", "", "", "", "", "", false, true},
		{"valid URL", "http://bar.foo.com/external-auth", "http://bar.foo.com/external-auth", "", "", "", "", "", false, false},
		{"valid URL - send body", "http://foo.com/external-auth", "http://foo.com/external-auth", "", "POST", "", "", "", false, false},
		{"valid URL - send body", "http://foo.com/external-auth", "http://foo.com/external-auth", "", "GET", "", "", "", false, false},
		{"valid URL - request redirect", "http://foo.com/external-auth", "http://foo.com/external-auth", "", "GET", "http://foo.com/redirect-me", "", "", false, false},
		{"auth snippet", "http://foo.com/external-auth", "http://foo.com/external-auth", "", "", "", "proxy_set_header My-Custom-Header 42;", "", false, false},
		{"auth cache ", "http://foo.com/external-auth", "http://foo.com/external-auth", "", "", "", "", "$foo$bar", false, false},
		{"redirect param", "http://bar.foo.com/external-auth", "http://bar.foo.com/external-auth", "origUrl", "", "", "", "", true, false},
	}

	for _, test := range tests {
		data[parser.GetAnnotationWithPrefix("auth-url")] = test.url
		data[parser.GetAnnotationWithPrefix("auth-signin")] = test.signinURL
		data[parser.GetAnnotationWithPrefix("auth-signin-redirect-param")] = test.signinURLRedirectParam
		data[parser.GetAnnotationWithPrefix("auth-method")] = fmt.Sprintf("%v", test.method)
		data[parser.GetAnnotationWithPrefix("auth-request-redirect")] = test.requestRedirect
		data[parser.GetAnnotationWithPrefix("auth-snippet")] = test.authSnippet
		data[parser.GetAnnotationWithPrefix("auth-cache-key")] = test.authCacheKey
		data[parser.GetAnnotationWithPrefix("auth-always-set-cookie")] = boolToString(test.authAlwaysSetCookie)

		i, err := NewParser(&resolver.Mock{}).Parse(ing)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but returned nil", test.title)
			}
			continue
		}
		if err != nil {
			t.Errorf("%v: unexpected error: %v", test.title, err)
		}

		u, ok := i.(*Config)
		if !ok {
			t.Errorf("%v: expected an External type", test.title)
		}
		if u.URL != test.url {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.url, u.URL)
		}
		if u.SigninURL != test.signinURL {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.signinURL, u.SigninURL)
		}
		if u.SigninURLRedirectParam != test.signinURLRedirectParam {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.signinURLRedirectParam, u.SigninURLRedirectParam)
		}
		if u.Method != test.method {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.method, u.Method)
		}
		if u.RequestRedirect != test.requestRedirect {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.requestRedirect, u.RequestRedirect)
		}
		if u.AuthSnippet != test.authSnippet {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.authSnippet, u.AuthSnippet)
		}
		if u.AuthCacheKey != test.authCacheKey {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.authCacheKey, u.AuthCacheKey)
		}

		if u.AlwaysSetCookie != test.authAlwaysSetCookie {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.authAlwaysSetCookie, u.AlwaysSetCookie)
		}
	}
}

func TestHeaderAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	tests := []struct {
		title         string
		url           string
		headers       string
		parsedHeaders []string
		expErr        bool
	}{
		{"single header", "http://goog.url", "h1", []string{"h1"}, false},
		{"nothing", "http://goog.url", "", []string{}, false},
		{"spaces", "http://goog.url", "  ", []string{}, false},
		{"two headers", "http://goog.url", "1,2", []string{"1", "2"}, false},
		{"two headers and empty entries", "http://goog.url", ",1,,2,", []string{"1", "2"}, false},
		{"header with spaces", "http://goog.url", "1 2", []string{}, true},
		{"header with other bad symbols", "http://goog.url", "1+2", []string{}, true},
	}

	for _, test := range tests {
		data[parser.GetAnnotationWithPrefix("auth-url")] = test.url
		data[parser.GetAnnotationWithPrefix("auth-response-headers")] = test.headers
		data[parser.GetAnnotationWithPrefix("auth-method")] = "GET"

		i, err := NewParser(&resolver.Mock{}).Parse(ing)
		if test.expErr {
			if err == nil {
				t.Error("expected error but retuned nil")
			}
			continue
		}

		u, ok := i.(*Config)
		if !ok {
			t.Errorf("%v: expected an External type", test.title)
			continue
		}

		if !reflect.DeepEqual(u.ResponseHeaders, test.parsedHeaders) {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.headers, u.ResponseHeaders)
		}
	}
}

func TestCacheDurationAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	tests := []struct {
		title          string
		url            string
		duration       string
		parsedDuration []string
		expErr         bool
	}{
		{"nothing", "http://goog.url", "", []string{DefaultCacheDuration}, false},
		{"spaces", "http://goog.url", "  ", []string{DefaultCacheDuration}, false},
		{"one duration", "http://goog.url", "5m", []string{"5m"}, false},
		{"two durations", "http://goog.url", "200 202 10m, 401 5m", []string{"200 202 10m", "401 5m"}, false},
		{"two durations and empty entries", "http://goog.url", ",5m,,401 10m,", []string{"5m", "401 10m"}, false},
		{"only status code provided", "http://goog.url", "200", []string{DefaultCacheDuration}, true},
		{"mixed valid/invalid", "http://goog.url", "5m, xaxax", []string{DefaultCacheDuration}, true},
		{"code after duration", "http://goog.url", "5m 200", []string{DefaultCacheDuration}, true},
	}

	for _, test := range tests {
		data[parser.GetAnnotationWithPrefix("auth-url")] = test.url
		data[parser.GetAnnotationWithPrefix("auth-cache-duration")] = test.duration

		i, err := NewParser(&resolver.Mock{}).Parse(ing)
		if test.expErr {
			if err == nil {
				t.Errorf("expected error but retuned nil")
			}
			continue
		}

		u, ok := i.(*Config)
		if !ok {
			t.Errorf("%v: expected an External type", test.title)
			continue
		}

		if !reflect.DeepEqual(u.AuthCacheDuration, test.parsedDuration) {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.duration, u.AuthCacheDuration)
		}
	}
}

func TestKeepaliveAnnotations(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	tests := []struct {
		title                string
		url                  string
		keepaliveConnections string
		keepaliveRequests    string
		keepaliveTimeout     string
		expectedConnections  int
		expectedRequests     int
		expectedTimeout      int
	}{
		{"all set", "http://goog.url", "5", "500", "50", 5, 500, 50},
		{"no annotation", "http://goog.url", "", "", "", defaultKeepaliveConnections, defaultKeepaliveRequests, defaultKeepaliveTimeout},
		{"default for connections", "http://goog.url", "x", "500", "50", defaultKeepaliveConnections, 500, 50},
		{"default for requests", "http://goog.url", "5", "x", "50", 5, defaultKeepaliveRequests, 50},
		{"default for invalid timeout", "http://goog.url", "5", "500", "x", 5, 500, defaultKeepaliveTimeout},
		{"variable in host", "http://$host:5000/a/b", "5", "", "", 0, defaultKeepaliveRequests, defaultKeepaliveTimeout},
		{"variable in path", "http://goog.url:5000/$path", "5", "", "", 5, defaultKeepaliveRequests, defaultKeepaliveTimeout},
		{"negative connections", "http://goog.url", "-2", "", "", 0, defaultKeepaliveRequests, defaultKeepaliveTimeout},
		{"negative requests", "http://goog.url", "5", "-1", "", 0, -1, defaultKeepaliveTimeout},
		{"negative timeout", "http://goog.url", "5", "", "-1", 0, defaultKeepaliveRequests, -1},
		{"negative request and timeout", "http://goog.url", "5", "-2", "-3", 0, -2, -3},
	}

	for _, test := range tests {
		data[parser.GetAnnotationWithPrefix("auth-url")] = test.url
		data[parser.GetAnnotationWithPrefix("auth-keepalive")] = test.keepaliveConnections
		data[parser.GetAnnotationWithPrefix("auth-keepalive-timeout")] = test.keepaliveTimeout
		data[parser.GetAnnotationWithPrefix("auth-keepalive-requests")] = test.keepaliveRequests

		i, err := NewParser(&resolver.Mock{}).Parse(ing)
		if err != nil {
			t.Errorf("%v: unexpected error: %v", test.title, err)
			continue
		}

		u, ok := i.(*Config)
		if !ok {
			t.Errorf("%v: expected an External type", test.title)
			continue
		}

		if u.URL != test.url {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.url, u.URL)
		}

		if u.KeepaliveConnections != test.expectedConnections {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.expectedConnections, u.KeepaliveConnections)
		}

		if u.KeepaliveRequests != test.expectedRequests {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.expectedRequests, u.KeepaliveRequests)
		}

		if u.KeepaliveTimeout != test.expectedTimeout {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.expectedTimeout, u.KeepaliveTimeout)
		}
	}
}

func TestParseStringToCacheDurations(t *testing.T) {

	tests := []struct {
		title             string
		duration          string
		expectedDurations []string
		expErr            bool
	}{
		{"empty", "", []string{DefaultCacheDuration}, false},
		{"invalid", ",200,", []string{DefaultCacheDuration}, true},
		{"single", ",200 5m,", []string{"200 5m"}, false},
		{"multiple with duration", ",5m,,401 10m,", []string{"5m", "401 10m"}, false},
		{"multiple durations", "200 202 401 5m, 418 30m", []string{"200 202 401 5m", "418 30m"}, false},
	}

	for _, test := range tests {

		dur, err := ParseStringToCacheDurations(test.duration)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but nil was returned", test.title)
			}
			continue
		}

		if !reflect.DeepEqual(dur, test.expectedDurations) {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.expectedDurations, dur)
		}
	}
}

func TestProxySetHeaders(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	ing.SetAnnotations(data)

	tests := []struct {
		title   string
		url     string
		headers map[string]string
		expErr  bool
	}{
		{"single header", "http://goog.url", map[string]string{"header": "h1"}, false},
		{"no header map", "http://goog.url", nil, true},
		{"header with spaces", "http://goog.url", map[string]string{"header": "bad value"}, false},
		{"header with other bad symbols", "http://goog.url", map[string]string{"header": "bad+value"}, false},
	}

	for _, test := range tests {
		data[parser.GetAnnotationWithPrefix("auth-url")] = test.url
		data[parser.GetAnnotationWithPrefix("auth-proxy-set-headers")] = "proxy-headers-map"
		data[parser.GetAnnotationWithPrefix("auth-method")] = "GET"

		configMapResolver := &resolver.Mock{
			ConfigMaps: map[string]*api.ConfigMap{},
		}

		if test.headers != nil {
			configMapResolver.ConfigMaps["proxy-headers-map"] = &api.ConfigMap{Data: test.headers}
		}

		i, err := NewParser(configMapResolver).Parse(ing)
		if test.expErr {
			if err == nil {
				t.Errorf("expected error but retuned nil")
			}
			continue
		}

		u, ok := i.(*Config)
		if !ok {
			t.Errorf("%v: expected an External type", test.title)
			continue
		}

		if !reflect.DeepEqual(u.ProxySetHeaders, test.headers) {
			t.Errorf("%v: expected \"%v\" but \"%v\" was returned", test.title, test.headers, u.ProxySetHeaders)
		}
	}
}
