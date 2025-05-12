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

package template

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/pmezard/go-difflib/difflib"
	apiv1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/authreq"
	"k8s.io/ingress-nginx/internal/ingress/annotations/modsecurity"
	"k8s.io/ingress-nginx/internal/ingress/annotations/opentelemetry"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ratelimit"
	"k8s.io/ingress-nginx/internal/ingress/annotations/rewrite"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

func init() {
	// the default value of nginx.TemplatePath assumes the template exists in
	// the root filesystem and not in the rootfs directory
	absPath, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "rootfs", nginx.TemplatePath))
	if err == nil {
		nginx.TemplatePath = absPath
	}
}

var (
	pathPrefix networking.PathType = networking.PathTypePrefix

	// TODO: add tests for SSLPassthrough
	tmplFuncTestcases = map[string]struct {
		Path              string
		Target            string
		Location          string
		ProxyPass         string
		AutoHTTPProxyPass string
		Sticky            bool
		XForwardedPrefix  string
		SecureBackend     bool
		enforceRegex      bool
	}{
		"when secure backend enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://upstream_balancer;",
			"proxy_pass https://upstream_balancer;",
			false,
			"",
			true,
			false,
		},
		"when secure backend and dynamic config enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://upstream_balancer;",
			"proxy_pass https://upstream_balancer;",
			false,
			"",
			true,
			false,
		},
		"when secure backend, stickiness and dynamic config enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://upstream_balancer;",
			"proxy_pass https://upstream_balancer;",
			true,
			"",
			true,
			false,
		},
		"invalid redirect / to / with dynamic config enabled": {
			"/",
			"/",
			"/",
			"proxy_pass http://upstream_balancer;",
			"proxy_pass $scheme://upstream_balancer;",
			false,
			"",
			false,
			false,
		},
		"invalid redirect / to /": {
			"/",
			"/",
			"/",
			"proxy_pass http://upstream_balancer;",
			"proxy_pass $scheme://upstream_balancer;",
			false,
			"",
			false,
			false,
		},
		"redirect / to /jenkins": {
			"/",
			"/jenkins",
			`~* "^/"`,
			`
rewrite "(?i)/" /jenkins break;
proxy_pass http://upstream_balancer;`,
			`
rewrite "(?i)/" /jenkins break;
proxy_pass $scheme://upstream_balancer;`,
			false,
			"",
			false,
			true,
		},
		"redirect / to /something with sticky enabled": {
			"/",
			"/something",
			`~* "^/"`,
			`
rewrite "(?i)/" /something break;
proxy_pass http://upstream_balancer;`,
			`
rewrite "(?i)/" /something break;
proxy_pass $scheme://upstream_balancer;`,
			true,
			"",
			false,
			true,
		},
		"redirect / to /something with sticky and dynamic config enabled": {
			"/",
			"/something",
			`~* "^/"`,
			`
rewrite "(?i)/" /something break;
proxy_pass http://upstream_balancer;`,
			`
rewrite "(?i)/" /something break;
proxy_pass $scheme://upstream_balancer;`,
			true,
			"",
			false,
			true,
		},
		"add the X-Forwarded-Prefix header": {
			"/there",
			"/something",
			`~* "^/there"`,
			`
rewrite "(?i)/there" /something break;
proxy_set_header X-Forwarded-Prefix "/there";
proxy_pass http://upstream_balancer;`,
			`
rewrite "(?i)/there" /something break;
proxy_set_header X-Forwarded-Prefix "/there";
proxy_pass $scheme://upstream_balancer;`,
			true,
			"/there",
			false,
			true,
		},
		"use ~* location modifier when ingress does not use rewrite/regex target but at least one other ingress does": {
			"/something",
			"/something",
			`~* "^/something"`,
			"proxy_pass http://upstream_balancer;",
			"proxy_pass $scheme://upstream_balancer;",
			false,
			"",
			false,
			true,
		},
	}
)

const (
	defaultBackend = "upstream-name"
	defaultHost    = "example.com"
	fooAuthHost    = "foo.com/auth"
)

func getTestDataDir() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(pwd, "../../../../test/data"), nil
}

func TestBuildLuaSharedDictionaries(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""

	// config lua dict
	cfg := config.Configuration{
		LuaSharedDicts: map[string]int{
			"configuration_data": 10240, "certificate_data": 20480,
		},
	}
	actual := buildLuaSharedDictionaries(cfg, invalidType)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	servers := []*ingress.Server{
		{
			Hostname:  "foo.bar",
			Locations: []*ingress.Location{{Path: "/", PathType: &pathPrefix}},
		},
		{
			Hostname:  "another.host",
			Locations: []*ingress.Location{{Path: "/", PathType: &pathPrefix}},
		},
	}
	// returns value from config
	configuration := buildLuaSharedDictionaries(cfg, servers)
	if !strings.Contains(configuration, "lua_shared_dict configuration_data 10M;\n") {
		t.Errorf("expected to include 'configuration_data' but got %s", configuration)
	}
	if !strings.Contains(configuration, "lua_shared_dict certificate_data 20M;\n") {
		t.Errorf("expected to include 'certificate_data' but got %s", configuration)
	}
	// test invalid config
	configuration = buildLuaSharedDictionaries(invalidType, servers)
	if configuration != "" {
		t.Errorf("expected an empty string, but got %s", configuration)
	}

	if actual != expected {
		t.Errorf("Expected '%v' but returned '%v' ", expected, actual)
	}
}

func TestLuaConfigurationRequestBodySize(t *testing.T) {
	cfg := config.Configuration{
		LuaSharedDicts: map[string]int{
			"configuration_data": 10240, "certificate_data": 20480,
		},
	}

	size := luaConfigurationRequestBodySize(cfg)
	if size != "21M" {
		t.Errorf("expected the size to be 21M but got: %v", size)
	}
}

func TestFormatIP(t *testing.T) {
	cases := map[string]struct {
		Input, Output string
	}{
		"ipv4-localhost": {"127.0.0.1", "127.0.0.1"},
		"ipv4-internet":  {"8.8.8.8", "8.8.8.8"},
		"ipv6-localhost": {"::1", "[::1]"},
		"ipv6-internet":  {"2001:4860:4860::8888", "[2001:4860:4860::8888]"},
		"invalid-ip":     {"nonsense", "nonsense"},
		"empty-ip":       {"", ""},
	}
	for k, tc := range cases {
		res := formatIP(tc.Input)
		if res != tc.Output {
			t.Errorf("%s: called formatIp('%s'); expected '%v' but returned '%v'", k, tc.Input, tc.Output, res)
		}
	}
}

func TestQuote(t *testing.T) {
	foo := "foo"
	cases := map[interface{}]string{
		"foo":      `"foo"`,
		"\"foo\"":  `"\"foo\""`,
		"foo\nbar": `"foo\nbar"`,
		&foo:       `"foo"`,
		10:         `"10"`,
	}
	for input, output := range cases {
		actual := quote(input)
		if actual != output {
			t.Errorf("quote('%s'): expected '%v' but returned '%v'", input, output, actual)
		}
	}
}

func TestBuildLocation(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := "/"
	actual := buildLocation(invalidType, true)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:     tc.Path,
			PathType: &pathPrefix,
			Rewrite:  rewrite.Config{Target: tc.Target},
		}

		newLoc := buildLocation(loc, tc.enforceRegex)
		if tc.Location != newLoc {
			t.Errorf("%s: expected '%v' but returned %v", k, tc.Location, newLoc)
		}
	}
}

func TestBuildProxyPass(t *testing.T) {
	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:             tc.Path,
			PathType:         &pathPrefix,
			Rewrite:          rewrite.Config{Target: tc.Target},
			Backend:          defaultBackend,
			XForwardedPrefix: tc.XForwardedPrefix,
		}

		if tc.SecureBackend {
			loc.BackendProtocol = httpsProtocol
		}

		backend := &ingress.Backend{
			Name: defaultBackend,
		}

		if tc.Sticky {
			backend.SessionAffinity = ingress.SessionAffinityConfig{
				AffinityType: "cookie",
				CookieSessionAffinity: ingress.CookieSessionAffinity{
					Locations: map[string][]string{
						defaultHost: {tc.Path},
					},
				},
			}
		}

		backends := []*ingress.Backend{backend}

		pp := buildProxyPass(defaultHost, backends, loc)
		if !strings.EqualFold(tc.ProxyPass, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, tc.ProxyPass, pp)
		}
	}
}

func TestBuildProxyPassAutoHttp(t *testing.T) {
	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:             tc.Path,
			Rewrite:          rewrite.Config{Target: tc.Target},
			Backend:          defaultBackend,
			XForwardedPrefix: tc.XForwardedPrefix,
		}

		if tc.SecureBackend {
			loc.BackendProtocol = httpsProtocol
		} else {
			loc.BackendProtocol = autoHTTPProtocol
		}

		backend := &ingress.Backend{
			Name: defaultBackend,
		}

		if tc.Sticky {
			backend.SessionAffinity = ingress.SessionAffinityConfig{
				AffinityType: "cookie",
				CookieSessionAffinity: ingress.CookieSessionAffinity{
					Locations: map[string][]string{
						defaultHost: {tc.Path},
					},
				},
			}
		}

		backends := []*ingress.Backend{backend}

		pp := buildProxyPass(defaultHost, backends, loc)
		if !strings.EqualFold(tc.AutoHTTPProxyPass, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, tc.ProxyPass, pp)
		}
	}
}

func TestBuildAuthLocation(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildAuthLocation(invalidType, "")

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	authURL := fooAuthHost
	globalAuthURL := "foo.com/global-auth"

	loc := &ingress.Location{
		ExternalAuth: authreq.Config{
			URL: authURL,
		},
		Path:             "/cat",
		EnableGlobalAuth: true,
	}

	encodedAuthURL := strings.ReplaceAll(base64.URLEncoding.EncodeToString([]byte(loc.Path)), "=", "")
	externalAuthPath := fmt.Sprintf("/_external-auth-%v-default", encodedAuthURL)

	testCases := []struct {
		title                    string
		authURL                  string
		globalAuthURL            string
		enableglobalExternalAuth bool
		expected                 string
	}{
		{"authURL, globalAuthURL and enabled", authURL, globalAuthURL, true, externalAuthPath},
		{"authURL, globalAuthURL and disabled", authURL, globalAuthURL, false, externalAuthPath},
		{"authURL, empty globalAuthURL and enabled", authURL, "", true, externalAuthPath},
		{"authURL, empty globalAuthURL and disabled", authURL, "", false, externalAuthPath},
		{"globalAuthURL and enabled", "", globalAuthURL, true, externalAuthPath},
		{"globalAuthURL and disabled", "", globalAuthURL, false, ""},
		{"all empty and enabled", "", "", true, ""},
		{"all empty and disabled", "", "", false, ""},
	}

	for _, testCase := range testCases {
		loc.ExternalAuth.URL = testCase.authURL
		loc.EnableGlobalAuth = testCase.enableglobalExternalAuth

		str := buildAuthLocation(loc, testCase.globalAuthURL)
		if str != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, str)
		}
	}
}

func TestShouldApplyGlobalAuth(t *testing.T) {
	authURL := fooAuthHost
	globalAuthURL := "foo.com/global-auth"

	loc := &ingress.Location{
		ExternalAuth: authreq.Config{
			URL: authURL,
		},
		Path:             "/cat",
		EnableGlobalAuth: true,
	}

	testCases := []struct {
		title                    string
		authURL                  string
		globalAuthURL            string
		enableglobalExternalAuth bool
		expected                 bool
	}{
		{"authURL, globalAuthURL and enabled", authURL, globalAuthURL, true, false},
		{"authURL, globalAuthURL and disabled", authURL, globalAuthURL, false, false},
		{"authURL, empty globalAuthURL and enabled", authURL, "", true, false},
		{"authURL, empty globalAuthURL and disabled", authURL, "", false, false},
		{"globalAuthURL and enabled", "", globalAuthURL, true, true},
		{"globalAuthURL and disabled", "", globalAuthURL, false, false},
		{"all empty and enabled", "", "", true, false},
		{"all empty and disabled", "", "", false, false},
	}

	for _, testCase := range testCases {
		loc.ExternalAuth.URL = testCase.authURL
		loc.EnableGlobalAuth = testCase.enableglobalExternalAuth

		result := shouldApplyGlobalAuth(loc, testCase.globalAuthURL)
		if result != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, result)
		}
	}
}

func TestBuildAuthResponseHeaders(t *testing.T) {
	externalAuthResponseHeaders := []string{"h1", "H-With-Caps-And-Dashes"}
	tests := []struct {
		headers  []string
		expected []string
		lua      bool
	}{
		{
			headers: externalAuthResponseHeaders,
			lua:     false,
			expected: []string{
				"auth_request_set $authHeader0 $upstream_http_h1;",
				"proxy_set_header 'h1' $authHeader0;",
				"auth_request_set $authHeader1 $upstream_http_h_with_caps_and_dashes;",
				"proxy_set_header 'H-With-Caps-And-Dashes' $authHeader1;",
			},
		},
		{
			headers: externalAuthResponseHeaders,
			lua:     true,
			expected: []string{
				"set $authHeader0 '';",
				"proxy_set_header 'h1' $authHeader0;",
				"set $authHeader1 '';",
				"proxy_set_header 'H-With-Caps-And-Dashes' $authHeader1;",
			},
		},
	}

	for _, test := range tests {
		got := buildAuthResponseHeaders(proxySetHeader(nil), test.headers, test.lua)
		if !reflect.DeepEqual(test.expected, got) {
			t.Errorf("Expected \n'%v'\nbut returned \n'%v'", test.expected, got)
		}
	}
}

func TestBuildAuthResponseLua(t *testing.T) {
	externalAuthResponseHeaders := []string{"h1", "H-With-Caps-And-Dashes"}
	expected := "h1,H-With-Caps-And-Dashes"

	headers := buildAuthUpstreamLuaHeaders(externalAuthResponseHeaders)

	if !reflect.DeepEqual(expected, headers) {
		t.Errorf("Expected \n'%v'\nbut returned \n'%v'", expected, headers)
	}
}

func TestBuildAuthProxySetHeaders(t *testing.T) {
	proxySetHeaders := map[string]string{
		"header1": "value1",
		"header2": "value2",
	}
	expected := []string{
		"proxy_set_header 'header1' 'value1';",
		"proxy_set_header 'header2' 'value2';",
	}

	headers := buildAuthProxySetHeaders(proxySetHeaders)

	if !reflect.DeepEqual(expected, headers) {
		t.Errorf("Expected \n'%v'\nbut returned \n'%v'", expected, headers)
	}
}

func TestBuildAuthUpstreamName(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildAuthUpstreamName(invalidType, "")

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	loc := &ingress.Location{
		ExternalAuth: authreq.Config{
			URL: fooAuthHost,
		},
		Path: "/cat",
	}

	encodedAuthURL := strings.ReplaceAll(base64.URLEncoding.EncodeToString([]byte(loc.Path)), "=", "")
	externalAuthPath := fmt.Sprintf("external-auth-%v-default", encodedAuthURL)

	testCases := []struct {
		title    string
		host     string
		expected string
	}{
		{"valid host", "auth.my.site", fmt.Sprintf("%s-%s", "auth.my.site", externalAuthPath)},
		{"valid host", "your.auth.site", fmt.Sprintf("%s-%s", "your.auth.site", externalAuthPath)},
		{"empty host", "", ""},
	}

	for _, testCase := range testCases {
		str := buildAuthUpstreamName(loc, testCase.host)
		if str != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, str)
		}
	}
}

func TestShouldApplyAuthUpstream(t *testing.T) {
	authURL := fooAuthHost

	loc := &ingress.Location{
		ExternalAuth: authreq.Config{
			URL:                  authURL,
			KeepaliveConnections: 0,
		},
		Path: "/cat",
	}

	cfg := config.Configuration{
		UseHTTP2: false,
	}

	testCases := []struct {
		title                string
		authURL              string
		keepaliveConnections int
		expected             bool
	}{
		{"authURL, no keepalive", authURL, 0, false},
		{"authURL, keepalive", authURL, 10, true},
		{"empty, no keepalive", "", 0, false},
		{"empty, keepalive", "", 10, false},
	}

	for _, testCase := range testCases {
		loc.ExternalAuth.URL = testCase.authURL
		loc.ExternalAuth.KeepaliveConnections = testCase.keepaliveConnections

		result := shouldApplyAuthUpstream(loc, cfg)
		if result != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, result)
		}
	}

	// keepalive is not supported with UseHTTP2
	cfg.UseHTTP2 = true
	for _, testCase := range testCases {
		loc.ExternalAuth.URL = testCase.authURL
		loc.ExternalAuth.KeepaliveConnections = testCase.keepaliveConnections

		result := shouldApplyAuthUpstream(loc, cfg)
		if result != false {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, false, result)
		}
	}
}

func TestExtractHostPort(t *testing.T) {
	testCases := []struct {
		title    string
		url      string
		expected string
	}{
		{"full URL", "https://my.auth.site:5000/path", "my.auth.site:5000"},
		{"URL with no port", "http://my.auth.site/path", "my.auth.site"},
		{"URL with no path", "https://my.auth.site:5000", "my.auth.site:5000"},
		{"URL no port and path", "http://my.auth.site", "my.auth.site"},
		{"missing method", "my.auth.site/path", ""},
		{"all empty", "", ""},
	}

	for _, testCase := range testCases {
		result := extractHostPort(testCase.url)
		if result != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, result)
		}
	}
}

func TestChangeHostPort(t *testing.T) {
	testCases := []struct {
		title    string
		url      string
		host     string
		expected string
	}{
		{"full URL", "https://my.auth.site:5000/path", "your.domain", "https://your.domain/path"},
		{"URL with no port", "http://my.auth.site/path", "your.domain", "http://your.domain/path"},
		{"URL with no path", "http://my.auth.site:5000", "your.domain:8888", "http://your.domain:8888"},
		{"URL with no port and path", "https://my.auth.site", "your.domain:8888", "https://your.domain:8888"},
		{"invalid host", "my.auth.site/path", "", ""},
		{"missing method", "my.auth.site/path", "your.domain", ""},
		{"all empty", "", "", ""},
	}

	for _, testCase := range testCases {
		result := changeHostPort(testCase.url, testCase.host)
		if result != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, result)
		}
	}
}

func TestTemplateWithData(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	f, err := os.Open(path.Join(pwd, "../../../../test/data/config.json"))
	if err != nil {
		t.Errorf("unexpected error reading json file: %v", err)
	}
	defer f.Close()
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Error("unexpected error reading json file: ", err)
	}
	var dat config.TemplateConfig
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, &dat); err != nil {
		t.Errorf("unexpected error unmarshalling json: %v", err)
	}
	if dat.ListenPorts == nil {
		dat.ListenPorts = &config.ListenPorts{}
	}

	ngxTpl, err := NewTemplate(nginx.TemplatePath)
	if err != nil {
		t.Errorf("invalid NGINX template: %v", err)
	}

	dat.Cfg.DefaultSSLCertificate = &ingress.SSLCert{}

	rt, err := ngxTpl.Write(&dat)
	if err != nil {
		t.Errorf("invalid NGINX template: %v", err)
	}

	if !strings.Contains(string(rt), "listen [2001:db8:a0b:12f0::1]") {
		t.Errorf("invalid NGINX template, expected IPV6 listen address not present")
	}

	if !strings.Contains(string(rt), "listen [3731:54:65fe:2::a7]") {
		t.Errorf("invalid NGINX template, expected IPV6 listen address not present")
	}

	if !strings.Contains(string(rt), "listen 2.2.2.2") {
		t.Errorf("invalid NGINX template, expected IPV4 listen address not present")
	}
}

func BenchmarkTemplateWithData(b *testing.B) {
	pwd, err := os.Getwd()
	if err != nil {
		b.Errorf("unexpected error: %v", err)
	}
	f, err := os.Open(path.Join(pwd, "../../../../test/data/config.json"))
	if err != nil {
		b.Errorf("unexpected error reading json file: %v", err)
	}
	defer f.Close()
	data, err := os.ReadFile(f.Name())
	if err != nil {
		b.Error("unexpected error reading json file: ", err)
	}
	var dat config.TemplateConfig
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, &dat); err != nil {
		b.Errorf("unexpected error unmarshalling json: %v", err)
	}

	ngxTpl, err := NewTemplate(nginx.TemplatePath)
	if err != nil {
		b.Errorf("invalid NGINX template: %v", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := ngxTpl.Write(&dat); err != nil {
			b.Errorf("unexpected error writing template: %v", err)
		}
	}
}

func TestBuildDenyVariable(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildDenyVariable(invalidType)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	a := buildDenyVariable("host1.example.com_/.well-known/acme-challenge")
	b := buildDenyVariable("host1.example.com_/.well-known/acme-challenge")
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected '%v' but returned '%v'", a, b)
	}
}

func TestBuildByteSize(t *testing.T) {
	cases := []struct {
		value    interface{}
		isOffset bool
		expected bool
	}{
		{"1000", false, true},
		{"1000k", false, true},
		{"1m", false, true},
		{"10g", false, false},
		{" 1m ", false, true},
		{"1000kk", false, false},
		{"1000km", false, false},
		{"1mm", false, false},
		{nil, false, false},
		{"", false, false},
		{"    ", false, false},
		{"1G", true, true},
		{"1000kk", true, false},
		{"", true, false},
	}

	for _, tc := range cases {
		val := isValidByteSize(tc.value, tc.isOffset)
		if tc.expected != val {
			t.Errorf("Expected '%v' but returned '%v'", tc.expected, val)
		}
	}
}

func TestIsLocationAllowed(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := false
	actual := isLocationAllowed(invalidType)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	loc := ingress.Location{
		Denied: nil,
	}

	isAllowed := isLocationAllowed(&loc)
	if !isAllowed {
		t.Errorf("Expected '%v' but returned '%v'", true, isAllowed)
	}
}

func TestBuildForwardedFor(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildForwardedFor(invalidType)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	inputStr := "X-Forwarded-For"
	expected = "$http_x_forwarded_for"
	actual = buildForwardedFor(inputStr)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestBuildForwardedHost(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildForwardedHost(invalidType)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	inputStr := "X-Forwarded-Host"
	expected = "$http_x_forwarded_host"
	actual = buildForwardedHost(inputStr)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestBuildResolvers(t *testing.T) {
	ipOne := net.ParseIP("192.0.0.1")
	ipTwo := net.ParseIP("2001:db8:1234:0000:0000:0000:0000:0000")
	ipList := []net.IP{ipOne, ipTwo}

	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildResolvers(invalidType, false)

	// Invalid Type for []net.IP
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	actual = buildResolvers(ipList, invalidType)

	// Invalid Type for bool
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	validResolver := "resolver 192.0.0.1 [2001:db8:1234::] valid=30s;"
	resolver := buildResolvers(ipList, false)

	if resolver != validResolver {
		t.Errorf("Expected '%v' but returned '%v'", validResolver, resolver)
	}

	validResolver = "resolver 192.0.0.1 valid=30s ipv6=off;"
	resolver = buildResolvers(ipList, true)

	if resolver != validResolver {
		t.Errorf("Expected '%v' but returned '%v'", validResolver, resolver)
	}
}

func TestBuildNextUpstream(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildNextUpstream(invalidType, "")

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	cases := map[string]struct {
		NextUpstream  string
		NonIdempotent bool
		Output        string
	}{
		"default": {
			"timeout http_500 http_502",
			false,
			"timeout http_500 http_502",
		},
		"global": {
			"timeout http_500 http_502",
			true,
			"timeout http_500 http_502 non_idempotent",
		},
		"local": {
			"timeout http_500 http_502 non_idempotent",
			false,
			"timeout http_500 http_502 non_idempotent",
		},
	}

	for k, tc := range cases {
		nextUpstream := buildNextUpstream(tc.NextUpstream, tc.NonIdempotent)
		if nextUpstream != tc.Output {
			t.Errorf(
				"%s: called buildNextUpstream('%s', %v); expected '%v' but returned '%v'",
				k,
				tc.NextUpstream,
				tc.NonIdempotent,
				tc.Output,
				nextUpstream,
			)
		}
	}
}

func TestBuildRateLimit(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := []string{}
	actual := buildRateLimit(invalidType)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	loc := &ingress.Location{}

	loc.RateLimit.Connections.Name = "con"
	loc.RateLimit.Connections.Limit = 1

	loc.RateLimit.RPS.Name = "rps"
	loc.RateLimit.RPS.Limit = 1
	loc.RateLimit.RPS.Burst = 1

	loc.RateLimit.RPM.Name = "rpm"
	loc.RateLimit.RPM.Limit = 2
	loc.RateLimit.RPM.Burst = 2

	loc.RateLimit.LimitRateAfter = 1
	loc.RateLimit.LimitRate = 1

	validLimits := []string{
		"limit_conn con 1;",
		"limit_req zone=rps burst=1 nodelay;",
		"limit_req zone=rpm burst=2 nodelay;",
		"limit_rate_after 1k;",
		"limit_rate 1k;",
	}

	limits := buildRateLimit(loc)

	for i, limit := range limits {
		if limit != validLimits[i] {
			t.Errorf("Expected '%v' but returned '%v'", validLimits, limits)
		}
	}

	// Invalid limit
	limits = buildRateLimit(&ingress.Ingress{})
	if !reflect.DeepEqual(expected, limits) {
		t.Errorf("Expected '%v' but returned '%v'", expected, limits)
	}
}

// TODO: Needs more tests
func TestBuildRateLimitZones(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := []string{}
	actual := buildRateLimitZones(invalidType)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

// TODO: Needs more tests
func TestFilterRateLimits(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := []ratelimit.Config{}
	actual := filterRateLimits(invalidType)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestBuildAuthSignURL(t *testing.T) {
	cases := map[string]struct {
		Input, RedirectParam, Output string
	}{
		"default url and redirect":              {"http://google.com", "rd", "http://google.com?rd=$pass_access_scheme://$http_host$escaped_request_uri"},
		"default url and custom redirect":       {"http://google.com", "orig", "http://google.com?orig=$pass_access_scheme://$http_host$escaped_request_uri"},
		"with random field":                     {"http://google.com?cat=0", "rd", "http://google.com?cat=0&rd=$pass_access_scheme://$http_host$escaped_request_uri"},
		"with random field and custom redirect": {"http://google.com?cat=0", "orig", "http://google.com?cat=0&orig=$pass_access_scheme://$http_host$escaped_request_uri"},
		"with rd field":                         {"http://google.com?cat&rd=$request", "rd", "http://google.com?cat&rd=$request"},
		"with orig field":                       {"http://google.com?cat&orig=$request", "orig", "http://google.com?cat&orig=$request"},
	}
	for k, tc := range cases {
		res := buildAuthSignURL(tc.Input, tc.RedirectParam)
		if res != tc.Output {
			t.Errorf("%s: called buildAuthSignURL('%s','%s'); expected '%v' but returned '%v'", k, tc.Input, tc.RedirectParam, tc.Output, res)
		}
	}
}

func TestIsLocationInLocationList(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := false
	actual := isLocationInLocationList(invalidType, "")

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	testCases := []struct {
		location        *ingress.Location
		rawLocationList string
		expected        bool
	}{
		{&ingress.Location{Path: "/match"}, "/match", true},
		{&ingress.Location{Path: "/match"}, ",/match", true},
		{&ingress.Location{Path: "/match"}, "/dontmatch", false},
		{&ingress.Location{Path: "/match"}, ",/dontmatch", false},
		{&ingress.Location{Path: "/match"}, "/dontmatch,/match", true},
		{&ingress.Location{Path: "/match"}, "/dontmatch,/dontmatcheither", false},
	}

	for _, testCase := range testCases {
		result := isLocationInLocationList(testCase.location, testCase.rawLocationList)
		if result != testCase.expected {
			t.Errorf(" expected %v but return %v, path: '%s', rawLocation: '%s'", testCase.expected, result, testCase.location.Path, testCase.rawLocationList)
		}
	}
}

func TestBuildUpstreamName(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildUpstreamName(invalidType)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:             tc.Path,
			PathType:         &pathPrefix,
			Rewrite:          rewrite.Config{Target: tc.Target},
			Backend:          defaultBackend,
			XForwardedPrefix: tc.XForwardedPrefix,
		}

		if tc.SecureBackend {
			loc.BackendProtocol = httpsProtocol
		}

		backend := &ingress.Backend{
			Name: defaultBackend,
		}

		expected := defaultBackend

		if tc.Sticky {
			backend.SessionAffinity = ingress.SessionAffinityConfig{
				AffinityType: "cookie",
				CookieSessionAffinity: ingress.CookieSessionAffinity{
					Locations: map[string][]string{
						defaultHost: {tc.Path},
					},
				},
			}
		}

		pp := buildUpstreamName(loc)
		if !strings.EqualFold(expected, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, expected, pp)
		}
	}
}

func TestEscapeLiteralDollar(t *testing.T) {
	escapedPath := escapeLiteralDollar("/$")
	expected := "/${literal_dollar}"
	if escapedPath != expected {
		t.Errorf("Expected %v but returned %v", expected, escapedPath)
	}

	escapedPath = escapeLiteralDollar("/hello-$/world-$/")
	expected = "/hello-${literal_dollar}/world-${literal_dollar}/"
	if escapedPath != expected {
		t.Errorf("Expected %v but returned %v", expected, escapedPath)
	}

	leaveUnchanged := "/leave-me/unchanged"
	escapedPath = escapeLiteralDollar(leaveUnchanged)
	if escapedPath != leaveUnchanged {
		t.Errorf("Expected %v but returned %v", leaveUnchanged, escapedPath)
	}

	escapedPath = escapeLiteralDollar(false)
	expected = ""
	if escapedPath != expected {
		t.Errorf("Expected %v but returned %v", expected, escapedPath)
	}
}

func TestOpentelemetryPropagateContext(t *testing.T) {
	tests := map[*ingress.Location]string{
		{BackendProtocol: httpProtocol}:     "opentelemetry_propagate;",
		{BackendProtocol: httpsProtocol}:    "opentelemetry_propagate;",
		{BackendProtocol: autoHTTPProtocol}: "opentelemetry_propagate;",
		{BackendProtocol: grpcProtocol}:     "opentelemetry_propagate;",
		{BackendProtocol: grpcsProtocol}:    "opentelemetry_propagate;",
		{BackendProtocol: fcgiProtocol}:     "opentelemetry_propagate;",
		nil:                                 "",
	}

	for loc, expectedDirective := range tests {
		actualDirective := opentelemetryPropagateContext(loc)
		if actualDirective != expectedDirective {
			t.Errorf("Expected %v but returned %v", expectedDirective, actualDirective)
		}
	}
}

func TestGetIngressInformation(t *testing.T) {
	testcases := map[string]struct {
		Ingress  interface{}
		Host     string
		Path     interface{}
		Expected *ingressInformation
	}{
		"wrong ingress type": {
			"wrongtype",
			"host1",
			"/ok",
			&ingressInformation{},
		},
		"wrong path type": {
			&ingress.Ingress{},
			"host1",
			10,
			&ingressInformation{},
		},
		"valid ingress definition with name validIng in namespace default  using a service with name a-svc port number 8080": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "validIng",
						Namespace: apiv1.NamespaceDefault,
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						DefaultBackend: &networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "a-svc",
								Port: networking.ServiceBackendPort{
									Number: 8080,
								},
							},
						},
					},
				},
			},
			"host1",
			"",
			&ingressInformation{
				Namespace: "default",
				Rule:      "validIng",
				Path:      "/",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
				Service:     "a-svc",
				ServicePort: "8080",
			},
		},
		"valid ingress definition with name validIng in namespace default  using a service with name a-svc port name b-svc": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "validIng",
						Namespace: apiv1.NamespaceDefault,
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						DefaultBackend: &networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "a-svc",
								Port: networking.ServiceBackendPort{
									Name: "b-svc",
								},
							},
						},
					},
				},
			},
			"host1",
			"",
			&ingressInformation{
				Namespace: "default",
				Rule:      "validIng",
				Path:      "/",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
				Service:     "a-svc",
				ServicePort: "b-svc",
			},
		},
		"valid ingress definition with name validIng in namespace default": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "validIng",
						Namespace: apiv1.NamespaceDefault,
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						DefaultBackend: &networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "a-svc",
							},
						},
					},
				},
			},
			"host1",
			"",
			&ingressInformation{
				Namespace: "default",
				Rule:      "validIng",
				Path:      "/",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
				Service: "a-svc",
			},
		},
		"valid ingress definition with name demo in namespace something and path /ok using a service with name b-svc port 80": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "demo",
						Namespace: "something",
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/ok",
												PathType: &pathPrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "b-svc",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
							{},
						},
					},
				},
			},
			"foo.bar",
			"/ok",
			&ingressInformation{
				Namespace: "something",
				Rule:      "demo",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
				Service:     "b-svc",
				ServicePort: "80",
			},
		},
		"valid ingress definition with name demo in namespace something and path /ok using a service with name b-svc port name b-svc-80": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "demo",
						Namespace: "something",
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/ok",
												PathType: &pathPrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "b-svc",
														Port: networking.ServiceBackendPort{
															Name: "b-svc-80",
														},
													},
												},
											},
										},
									},
								},
							},
							{},
						},
					},
				},
			},
			"foo.bar",
			"/ok",
			&ingressInformation{
				Namespace: "something",
				Rule:      "demo",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
				Service:     "b-svc",
				ServicePort: "b-svc-80",
			},
		},
		"valid ingress definition with name demo in namespace something and path /ok with a nil backend service": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "demo",
						Namespace: "something",
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/ok",
												PathType: &pathPrefix,
												Backend: networking.IngressBackend{
													Service: nil,
												},
											},
										},
									},
								},
							},
							{},
						},
					},
				},
			},
			"foo.bar",
			"/ok",
			&ingressInformation{
				Namespace: "something",
				Rule:      "demo",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
			},
		},
		"valid ingress definition with name demo in namespace something and path /ok with both a nil service and a valid one": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "demo",
						Namespace: "something",
						Annotations: map[string]string{
							"ingress.annotation": "ok",
						},
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/ok",
												PathType: &pathPrefix,
												Backend: networking.IngressBackend{
													Service: nil,
												},
											},
											{
												Path:     "/oksvc",
												PathType: &pathPrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "b-svc",
														Port: networking.ServiceBackendPort{
															Name: "b-svc-80",
														},
													},
												},
											},
										},
									},
								},
							},
							{},
						},
					},
				},
			},
			"foo.bar",
			"/oksvc",
			&ingressInformation{
				Namespace: "something",
				Rule:      "demo",
				Annotations: map[string]string{
					"ingress.annotation": "ok",
				},
				Service:     "b-svc",
				ServicePort: "b-svc-80",
			},
		},
	}

	for title, testCase := range testcases {
		info := getIngressInformation(testCase.Ingress, testCase.Host, testCase.Path)

		if !info.Equal(testCase.Expected) {
			t.Fatalf("%s: expected '%v' but returned %v", title, testCase.Expected, info)
		}
	}
}

func TestBuildCustomErrorLocationsPerServer(t *testing.T) {
	testCases := []struct {
		server          interface{}
		expectedResults []errorLocation
	}{
		{ // Single ingress
			&ingress.Server{Locations: []*ingress.Location{
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-backend",
					CustomHTTPErrors:           []int{401, 402},
				},
			}},
			[]errorLocation{
				{
					UpstreamName: "custom-default-backend-test-backend",
					Codes:        []int{401, 402},
				},
			},
		},
		{ // Two ingresses, overlapping error codes, same backend
			&ingress.Server{Locations: []*ingress.Location{
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-backend",
					CustomHTTPErrors:           []int{401, 402},
				},
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-backend",
					CustomHTTPErrors:           []int{402, 403},
				},
			}},
			[]errorLocation{
				{
					UpstreamName: "custom-default-backend-test-backend",
					Codes:        []int{401, 402, 403},
				},
			},
		},
		{ // Two ingresses, overlapping error codes, different backends
			&ingress.Server{Locations: []*ingress.Location{
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-one",
					CustomHTTPErrors:           []int{401, 402},
				},
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-two",
					CustomHTTPErrors:           []int{402, 403},
				},
			}},
			[]errorLocation{
				{
					UpstreamName: "custom-default-backend-test-one",
					Codes:        []int{401, 402},
				},
				{
					UpstreamName: "custom-default-backend-test-two",
					Codes:        []int{402, 403},
				},
			},
		},
		{ // Many ingresses, overlapping error codes, different backends
			&ingress.Server{Locations: []*ingress.Location{
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-one",
					CustomHTTPErrors:           []int{401, 402},
				},
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-one",
					CustomHTTPErrors:           []int{501, 502},
				},
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-two",
					CustomHTTPErrors:           []int{409, 410},
				},
				{
					DefaultBackendUpstreamName: "custom-default-backend-test-two",
					CustomHTTPErrors:           []int{504, 505},
				},
			}},
			[]errorLocation{
				{
					UpstreamName: "custom-default-backend-test-one",
					Codes:        []int{401, 402, 501, 502},
				},
				{
					UpstreamName: "custom-default-backend-test-two",
					Codes:        []int{409, 410, 504, 505},
				},
			},
		},
	}

	for _, c := range testCases {
		response := buildCustomErrorLocationsPerServer(c.server)
		if !reflect.DeepEqual(c.expectedResults, response) {
			t.Errorf("Expected %+v but got %+v", c.expectedResults, response)
		}
	}
}

func TestProxySetHeader(t *testing.T) {
	tests := []struct {
		name     string
		loc      interface{}
		expected string
	}{
		{
			name:     "nil",
			loc:      nil,
			expected: "proxy_set_header",
		},
		{
			name:     "invalid type",
			loc:      &ingress.Ingress{},
			expected: "proxy_set_header",
		},
		{
			name:     "http backend",
			loc:      &ingress.Location{},
			expected: "proxy_set_header",
		},
		{
			name: "gRPC backend",
			loc: &ingress.Location{
				BackendProtocol: grpcProtocol,
			},
			expected: "grpc_set_header",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proxySetHeader(tt.loc); got != tt.expected {
				t.Errorf("proxySetHeader() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestBuildOpenTelemetry(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := ""
	actual := buildOpentelemetry(invalidType, []*ingress.Server{})

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	cfgNoHost := config.Configuration{
		EnableOpentelemetry: true,
	}
	expected = "\r\n"
	actual = buildOpentelemetry(cfgNoHost, []*ingress.Server{})

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	cfgOpenTelemetry := config.Configuration{
		EnableOpentelemetry:        true,
		OpentelemetryOperationName: "my-operation-name",
	}
	expected = "\r\n"
	expected += "opentelemetry_operation_name \"my-operation-name\";\n"
	actual = buildOpentelemetry(cfgOpenTelemetry, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestEnforceRegexModifier(t *testing.T) {
	invalidType := &ingress.Ingress{}
	expected := false
	actual := enforceRegexModifier(invalidType)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	locs := []*ingress.Location{
		{
			Rewrite: rewrite.Config{
				Target:   "/alright",
				UseRegex: true,
			},
			Path:     "/ok",
			PathType: &pathPrefix,
		},
	}
	expected = true
	actual = enforceRegexModifier(locs)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

//nolint:dupl // Ignore dupl errors for similar test case
func TestShouldLoadModSecurityModule(t *testing.T) {
	// ### Invalid argument type tests ###
	// The first tests should return false.
	expected := false

	invalidType := &ingress.Ingress{}
	actual := shouldLoadModSecurityModule(config.Configuration{}, invalidType)
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	actual = shouldLoadModSecurityModule(invalidType, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	// ### Functional tests ###
	actual = shouldLoadModSecurityModule(config.Configuration{}, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	// All further tests should return true.
	expected = true

	configuration := config.Configuration{EnableModsecurity: true}
	actual = shouldLoadModSecurityModule(configuration, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	servers := []*ingress.Server{
		{
			Locations: []*ingress.Location{
				{
					ModSecurity: modsecurity.Config{
						Enable: true,
					},
				},
			},
		},
	}
	actual = shouldLoadModSecurityModule(config.Configuration{}, servers)
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	actual = shouldLoadModSecurityModule(configuration, servers)
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestOpentelemetryForLocation(t *testing.T) {
	trueVal := true
	falseVal := false

	loadOT := `opentelemetry on;
opentelemetry_propagate;
opentelemetry_trust_incoming_spans on;`
	loadOTUntrustedSpan := `opentelemetry on;
opentelemetry_propagate;
opentelemetry_trust_incoming_spans off;`
	testCases := []struct {
		description     string
		globalOT        bool
		isSetInLoc      bool
		isOTInLoc       *bool
		globalTrust     bool
		isTrustSetInLoc bool
		isTrustInLoc    *bool
		expected        string
	}{
		{"globally enabled, without annotation", true, false, nil, true, false, nil, loadOT},
		{"globally enabled and enabled in location", true, true, &trueVal, true, false, nil, loadOT},
		{"globally disabled and not enabled in location", false, false, nil, true, false, nil, ""},
		{"globally disabled but enabled in location", false, true, &trueVal, true, false, nil, loadOT},
		{"globally trusted, not trusted in location", true, false, nil, true, true, &falseVal, loadOTUntrustedSpan},
		{"not globally trusted, trust set in location", true, false, nil, false, true, &trueVal, loadOT},
		{"not globally trusted, trust not set in location", true, false, nil, false, false, nil, loadOTUntrustedSpan},
	}

	for _, testCase := range testCases {
		il := &ingress.Location{
			Opentelemetry: opentelemetry.Config{Set: testCase.isSetInLoc, TrustSet: testCase.isTrustSetInLoc},
		}
		if il.Opentelemetry.Set {
			il.Opentelemetry.Enabled = *testCase.isOTInLoc
		}
		if il.Opentelemetry.TrustSet {
			il.Opentelemetry.TrustEnabled = *testCase.isTrustInLoc
		}

		actual := buildOpentelemetryForLocation(testCase.globalOT, testCase.globalTrust, il)

		if testCase.expected != actual {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.description, testCase.expected, actual)
		}
	}
}

//nolint:dupl // Ignore dupl errors for similar test case
func TestShouldLoadOpentelemetryModule(t *testing.T) {
	// ### Invalid argument type tests ###
	// The first tests should return false.
	expected := false

	invalidType := &ingress.Ingress{}
	actual := shouldLoadOpentelemetryModule(config.Configuration{}, invalidType)
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	actual = shouldLoadOpentelemetryModule(invalidType, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	// ### Functional tests ###
	actual = shouldLoadOpentelemetryModule(config.Configuration{}, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	// All further tests should return true.
	expected = true

	configuration := config.Configuration{EnableOpentelemetry: true}
	actual = shouldLoadOpentelemetryModule(configuration, []*ingress.Server{})
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	servers := []*ingress.Server{
		{
			Locations: []*ingress.Location{
				{
					Opentelemetry: opentelemetry.Config{
						Enabled: true,
					},
				},
			},
		},
	}
	actual = shouldLoadOpentelemetryModule(config.Configuration{}, servers)
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	actual = shouldLoadOpentelemetryModule(configuration, servers)
	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestModSecurityForLocation(t *testing.T) {
	loadModule := `modsecurity on;
`

	modSecCfg := `modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;
`

	modsecOff := "modsecurity off;"

	modsecRule := `modsecurity_rules '
#RULE#
';
`
	owaspRules := `modsecurity_rules_file /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf;
`

	transactionID := "0000000"
	transactionCfg := fmt.Sprintf(`modsecurity_transaction_id "%v";
`, transactionID)

	testRule := "#RULE#"

	testCases := []struct {
		description         string
		isEnabledInCM       bool
		isOwaspEnabledInCM  bool
		isEnabledInLoc      bool
		isEnableSetInLoc    bool
		isOwaspEnabledInLoc bool
		snippet             string
		transactionID       string
		expected            string
	}{
		{"configmap enabled, configmap OWASP disabled, without annotation, snippet or transaction ID", true, false, false, false, false, "", "", ""},
		{"configmap enabled, configmap OWASP disabled, without annotation, snippet and with transaction ID", true, false, false, false, false, "", transactionID, transactionCfg},
		{"configmap enabled, configmap OWASP enabled, without annotation, OWASP enabled", true, true, false, false, false, "", "", ""},
		{"configmap enabled, configmap OWASP enabled, without annotation, OWASP disabled, with snippet and no transaction ID", true, true, false, false, false, testRule, "", modsecRule},
		{"configmap enabled, configmap OWASP enabled, without annotation, OWASP disabled, with snippet and transaction ID", true, true, false, false, false, testRule, transactionID, fmt.Sprintf("%v%v", modsecRule, transactionCfg)},
		{"configmap enabled, configmap OWASP disabled, annotation enabled, OWASP disabled", true, false, true, true, false, "", "", ""},
		{"configmap enabled, configmap OWASP disabled, annotation enabled, OWASP enabled, no snippet and no transaction ID", true, false, true, true, true, "", "", owaspRules},
		{"configmap enabled, configmap OWASP disabled, annotation disabled, OWASP disabled, no snippet and no transaction ID", true, false, false, true, false, "", "", modsecOff},
		{"configmap enabled, configmap OWASP disabled, annotation enabled, OWASP enabled, with snippet and no transaction ID", true, false, true, true, true, "", "", owaspRules},
		{"configmap enabled, configmap OWASP disabled, annotation enabled, OWASP enabled, with snippet and transaction ID", true, false, true, true, true, "", transactionID, fmt.Sprintf("%v%v", transactionCfg, owaspRules)},
		{"configmap enabled, configmap OWASP enabled, annotation enabled, OWASP disabled", true, true, true, true, false, "", "", ""},
		{"configmap disabled, annotation enabled, OWASP disabled", false, false, true, true, false, "", "", fmt.Sprintf("%v%v", loadModule, modSecCfg)},
		{"configmap disabled, annotation disabled, OWASP disabled", false, false, false, true, false, "", "", ""},
		{"configmap disabled, annotation enabled, OWASP disabled", false, false, true, true, false, testRule, "", fmt.Sprintf("%v%v", loadModule, modsecRule)},
		{"configmap disabled, annotation enabled, OWASP enabled", false, false, true, true, false, testRule, "", fmt.Sprintf("%v%v", loadModule, modsecRule)},
	}

	for _, testCase := range testCases {
		il := &ingress.Location{
			ModSecurity: modsecurity.Config{
				Enable:        testCase.isEnabledInLoc,
				EnableSet:     testCase.isEnableSetInLoc,
				OWASPRules:    testCase.isOwaspEnabledInLoc,
				Snippet:       testCase.snippet,
				TransactionID: testCase.transactionID,
			},
		}

		actual := buildModSecurityForLocation(config.Configuration{
			EnableModsecurity:    testCase.isEnabledInCM,
			EnableOWASPCoreRules: testCase.isOwaspEnabledInCM,
		}, il)

		if testCase.expected != actual {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.description, testCase.expected, actual)
		}
	}
}

func TestBuildServerName(t *testing.T) {
	testCases := []struct {
		title    string
		hostname string
		expected string
	}{
		{"simple domain", "foo.bar", "foo.bar"},
		{"simple www domain", "www.foo.bar", "www.foo.bar"},
		{"wildcard domain", "*.foo.bar", "~^(?<subdomain>[\\w-]+)\\.foo\\.bar$"},
		{"wildcard two levels domain", "*.sub.foo.bar", "~^(?<subdomain>[\\w-]+)\\.sub\\.foo\\.bar$"},
	}

	for _, testCase := range testCases {
		result := buildServerName(testCase.hostname)
		if result != testCase.expected {
			t.Errorf("%v: expected '%v' but returned '%v'", testCase.title, testCase.expected, result)
		}
	}
}

func TestCleanConf(t *testing.T) {
	testDataDir, err := getTestDataDir()
	if err != nil {
		t.Error("unexpected error reading conf file: ", err)
	}
	actual := &bytes.Buffer{}
	{
		data, err := os.ReadFile(testDataDir + "/cleanConf.src.conf")
		if err != nil {
			t.Error("unexpected error reading conf file: ", err)
		}
		in := bytes.NewBuffer(data)
		err = cleanConf(in, actual)
		if err != nil {
			t.Error("cleanConf failed: ", err)
		}
	}

	expected, err := os.ReadFile(testDataDir + "/cleanConf.expected.conf")
	if err != nil {
		t.Error("unexpected error reading conf file: ", err)
	}
	if !bytes.Equal(expected, actual.Bytes()) {
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{A: strings.SplitAfter(string(expected), "\n"), B: strings.SplitAfter(actual.String(), "\n"), Context: 3})
		if err != nil {
			t.Error("failed to get diff for cleanConf", err)
		}
		t.Errorf("cleanConf result don't match with expected: %s", diff)
	}
}
