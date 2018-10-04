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
	"io/ioutil"
	"net"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"encoding/base64"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authreq"
	"k8s.io/ingress-nginx/internal/ingress/annotations/luarestywaf"
	"k8s.io/ingress-nginx/internal/ingress/annotations/rewrite"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

var (
	// TODO: add tests for SSLPassthrough
	tmplFuncTestcases = map[string]struct {
		Path                        string
		Target                      string
		Location                    string
		ProxyPass                   string
		AddBaseURL                  bool
		BaseURLScheme               string
		Sticky                      bool
		XForwardedPrefix            bool
		DynamicConfigurationEnabled bool
		SecureBackend               bool
		enforceRegex                bool
	}{
		"when secure backend enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://upstream-name;",
			false,
			"",
			false,
			false,
			false,
			true,
			false},
		"when secure backend and stickeness enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://sticky-upstream-name;",
			false,
			"",
			true,
			false,
			false,
			true,
			false},
		"when secure backend and dynamic config enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://upstream_balancer;",
			false,
			"",
			false,
			false,
			true,
			true,
			false},
		"when secure backend, stickeness and dynamic config enabled": {
			"/",
			"/",
			"/",
			"proxy_pass https://upstream_balancer;",
			false,
			"",
			true,
			false,
			true,
			true,
			false},
		"invalid redirect / to / with dynamic config enabled": {
			"/",
			"/",
			"/",
			"proxy_pass http://upstream_balancer;",
			false,
			"",
			false,
			false,
			true,
			false,
			false},
		"invalid redirect / to /": {
			"/",
			"/",
			"/",
			"proxy_pass http://upstream-name;",
			false,
			"",
			false,
			false,
			false,
			false,
			false},
		"redirect / to /jenkins": {
			"/",
			"/jenkins",
			"~* ^/",
			`
rewrite "(?i)/(.*)" /jenkins/$1 break;
rewrite "(?i)/$" /jenkins/ break;
proxy_pass http://upstream-name;
`,
			false,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /something to /": {
			"/something",
			"/",
			`~* "^/something\/?(?<baseuri>.*)"`,
			`
rewrite "(?i)/something/(.*)" /$1 break;
rewrite "(?i)/something$" / break;
proxy_pass http://upstream-name;
`,
			false,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /end-with-slash/ to /not-root": {
			"/end-with-slash/",
			"/not-root",
			`~* "^/end-with-slash/(?<baseuri>.*)"`,
			`
rewrite "(?i)/end-with-slash/(.*)" /not-root/$1 break;
rewrite "(?i)/end-with-slash/$" /not-root/ break;
proxy_pass http://upstream-name;
`,
			false,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /something-complex to /not-root": {
			"/something-complex",
			"/not-root",
			`~* "^/something-complex\/?(?<baseuri>.*)"`,
			`
rewrite "(?i)/something-complex/(.*)" /not-root/$1 break;
rewrite "(?i)/something-complex$" /not-root/ break;
proxy_pass http://upstream-name;
`,
			false,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect / to /jenkins and rewrite": {
			"/",
			"/jenkins",
			"~* ^/",
			`
rewrite "(?i)/(.*)" /jenkins/$1 break;
rewrite "(?i)/$" /jenkins/ break;
proxy_pass http://upstream-name;

set_escape_uri $escaped_base_uri $baseuri;
subs_filter '(<(?:H|h)(?:E|e)(?:A|a)(?:D|d)(?:[^">]|"[^"]*")*>)' '$1<base href="$scheme://$http_host/$escaped_base_uri">' ro;
`,
			true,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /something to / and rewrite": {
			"/something",
			"/",
			`~* "^/something\/?(?<baseuri>.*)"`,
			`
rewrite "(?i)/something/(.*)" /$1 break;
rewrite "(?i)/something$" / break;
proxy_pass http://upstream-name;

set_escape_uri $escaped_base_uri $baseuri;
subs_filter '(<(?:H|h)(?:E|e)(?:A|a)(?:D|d)(?:[^">]|"[^"]*")*>)' '$1<base href="$scheme://$http_host/something/$escaped_base_uri">' ro;
`,
			true,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /end-with-slash/ to /not-root and rewrite": {
			"/end-with-slash/",
			"/not-root",
			`~* "^/end-with-slash/(?<baseuri>.*)"`,
			`
rewrite "(?i)/end-with-slash/(.*)" /not-root/$1 break;
rewrite "(?i)/end-with-slash/$" /not-root/ break;
proxy_pass http://upstream-name;

set_escape_uri $escaped_base_uri $baseuri;
subs_filter '(<(?:H|h)(?:E|e)(?:A|a)(?:D|d)(?:[^">]|"[^"]*")*>)' '$1<base href="$scheme://$http_host/end-with-slash/$escaped_base_uri">' ro;
`,
			true,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /something-complex to /not-root and rewrite": {
			"/something-complex",
			"/not-root",
			`~* "^/something-complex\/?(?<baseuri>.*)"`,
			`
rewrite "(?i)/something-complex/(.*)" /not-root/$1 break;
rewrite "(?i)/something-complex$" /not-root/ break;
proxy_pass http://upstream-name;

set_escape_uri $escaped_base_uri $baseuri;
subs_filter '(<(?:H|h)(?:E|e)(?:A|a)(?:D|d)(?:[^">]|"[^"]*")*>)' '$1<base href="$scheme://$http_host/something-complex/$escaped_base_uri">' ro;
`,
			true,
			"",
			false,
			false,
			false,
			false,
			true},
		"redirect /something to / and rewrite with specific scheme": {
			"/something",
			"/",
			`~* "^/something\/?(?<baseuri>.*)"`,
			`
rewrite "(?i)/something/(.*)" /$1 break;
rewrite "(?i)/something$" / break;
proxy_pass http://upstream-name;

set_escape_uri $escaped_base_uri $baseuri;
subs_filter '(<(?:H|h)(?:E|e)(?:A|a)(?:D|d)(?:[^">]|"[^"]*")*>)' '$1<base href="http://$http_host/something/$escaped_base_uri">' ro;
`,
			true,
			"http",
			false,
			false,
			false,
			false,
			true},
		"redirect / to /something with sticky enabled": {
			"/",
			"/something",
			`~* ^/`,
			`
rewrite "(?i)/(.*)" /something/$1 break;
rewrite "(?i)/$" /something/ break;
proxy_pass http://sticky-upstream-name;
`,
			false,
			"http",
			true,
			false,
			false,
			false,
			true},
		"redirect / to /something with sticky and dynamic config enabled": {
			"/",
			"/something",
			`~* ^/`,
			`
rewrite "(?i)/(.*)" /something/$1 break;
rewrite "(?i)/$" /something/ break;
proxy_pass http://upstream_balancer;
`,
			false,
			"http",
			true,
			false,
			true,
			false,
			true},
		"add the X-Forwarded-Prefix header": {
			"/there",
			"/something",
			`~* "^/there\/?(?<baseuri>.*)"`,
			`
rewrite "(?i)/there/(.*)" /something/$1 break;
rewrite "(?i)/there$" /something/ break;
proxy_set_header X-Forwarded-Prefix "/there/";
proxy_pass http://sticky-upstream-name;
`,
			false,
			"http",
			true,
			true,
			false,
			false,
			true},
		"use ~* location modifier when ingress does not use rewrite/regex target but at least one other ingress does": {
			"/something",
			"/something",
			`~* "^/something"`,
			"proxy_pass http://upstream-name;",
			false,
			"",
			false,
			false,
			false,
			false,
			true},
	}
)

func TestBuildLuaSharedDictionaries(t *testing.T) {
	servers := []*ingress.Server{
		{
			Hostname:  "foo.bar",
			Locations: []*ingress.Location{{Path: "/", LuaRestyWAF: luarestywaf.Config{}}},
		},
		{
			Hostname:  "another.host",
			Locations: []*ingress.Location{{Path: "/", LuaRestyWAF: luarestywaf.Config{}}},
		},
	}

	config := buildLuaSharedDictionaries(servers, false, false)
	if config != "" {
		t.Errorf("expected to not configure any lua shared dictionary, but generated %s", config)
	}
	config = buildLuaSharedDictionaries(servers, true, false)
	if !strings.Contains(config, "lua_shared_dict configuration_data") {
		t.Errorf("expected to include 'configuration_data' but got %s", config)
	}
	if strings.Contains(config, "waf_storage") {
		t.Errorf("expected to not include 'waf_storage' but got %s", config)
	}

	servers[1].Locations[0].LuaRestyWAF = luarestywaf.Config{Mode: "ACTIVE"}
	config = buildLuaSharedDictionaries(servers, false, false)
	if !strings.Contains(config, "lua_shared_dict waf_storage") {
		t.Errorf("expected to configure 'waf_storage', but got %s", config)
	}
	config = buildLuaSharedDictionaries(servers, true, false)
	if !strings.Contains(config, "lua_shared_dict waf_storage") {
		t.Errorf("expected to configure 'waf_storage', but got %s", config)
	}

	config = buildLuaSharedDictionaries(servers, false, true)
	if config != "" {
		t.Errorf("expected to not configure any lua shared dictionary, but generated %s", config)
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

func TestBuildLocation(t *testing.T) {
	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:    tc.Path,
			Rewrite: rewrite.Config{Target: tc.Target, AddBaseURL: tc.AddBaseURL},
		}

		newLoc := buildLocation(loc, tc.enforceRegex)
		if tc.Location != newLoc {
			t.Errorf("%s: expected '%v' but returned %v", k, tc.Location, newLoc)
		}
	}
}

func TestBuildProxyPass(t *testing.T) {
	defaultBackend := "upstream-name"
	defaultHost := "example.com"

	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:             tc.Path,
			Rewrite:          rewrite.Config{Target: tc.Target, AddBaseURL: tc.AddBaseURL, BaseURLScheme: tc.BaseURLScheme},
			Backend:          defaultBackend,
			XForwardedPrefix: tc.XForwardedPrefix,
		}

		backend := &ingress.Backend{
			Name:   defaultBackend,
			Secure: tc.SecureBackend,
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

		pp := buildProxyPass(defaultHost, backends, loc, tc.DynamicConfigurationEnabled)
		if !strings.EqualFold(tc.ProxyPass, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, tc.ProxyPass, pp)
		}
	}
}

func TestBuildAuthLocation(t *testing.T) {
	authURL := "foo.com/auth"

	loc := &ingress.Location{
		ExternalAuth: authreq.Config{
			URL: authURL,
		},
		Path: "/cat",
	}

	str := buildAuthLocation(loc)

	encodedAuthURL := strings.Replace(base64.URLEncoding.EncodeToString([]byte(loc.Path)), "=", "", -1)
	expected := fmt.Sprintf("/_external-auth-%v", encodedAuthURL)

	if str != expected {
		t.Errorf("Expected \n'%v'\nbut returned \n'%v'", expected, str)
	}
}

func TestBuildAuthResponseHeaders(t *testing.T) {
	loc := &ingress.Location{
		ExternalAuth: authreq.Config{ResponseHeaders: []string{"h1", "H-With-Caps-And-Dashes"}},
	}
	headers := buildAuthResponseHeaders(loc)
	expected := []string{
		"auth_request_set $authHeader0 $upstream_http_h1;",
		"proxy_set_header 'h1' $authHeader0;",
		"auth_request_set $authHeader1 $upstream_http_h_with_caps_and_dashes;",
		"proxy_set_header 'H-With-Caps-And-Dashes' $authHeader1;",
	}

	if !reflect.DeepEqual(expected, headers) {
		t.Errorf("Expected \n'%v'\nbut returned \n'%v'", expected, headers)
	}
}

func TestTemplateWithData(t *testing.T) {
	pwd, _ := os.Getwd()
	f, err := os.Open(path.Join(pwd, "../../../../test/data/config.json"))
	if err != nil {
		t.Errorf("unexpected error reading json file: %v", err)
	}
	defer f.Close()
	data, err := ioutil.ReadFile(f.Name())
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

	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ngxTpl, err := NewTemplate("/etc/nginx/template/nginx.tmpl", fs)
	if err != nil {
		t.Errorf("invalid NGINX template: %v", err)
	}

	rt, err := ngxTpl.Write(dat)
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
	pwd, _ := os.Getwd()
	f, err := os.Open(path.Join(pwd, "../../../../test/data/config.json"))
	if err != nil {
		b.Errorf("unexpected error reading json file: %v", err)
	}
	defer f.Close()
	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		b.Error("unexpected error reading json file: ", err)
	}
	var dat config.TemplateConfig
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, &dat); err != nil {
		b.Errorf("unexpected error unmarshalling json: %v", err)
	}

	fs, err := file.NewFakeFS()
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	ngxTpl, err := NewTemplate("/etc/nginx/template/nginx.tmpl", fs)
	if err != nil {
		b.Errorf("invalid NGINX template: %v", err)
	}

	for i := 0; i < b.N; i++ {
		ngxTpl.Write(dat)
	}
}

func TestBuildDenyVariable(t *testing.T) {
	a := buildDenyVariable("host1.example.com_/.well-known/acme-challenge")
	b := buildDenyVariable("host1.example.com_/.well-known/acme-challenge")
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected '%v' but returned '%v'", a, b)
	}
}

func TestBuildClientBodyBufferSize(t *testing.T) {
	a := isValidClientBodyBufferSize("1000")
	if !a {
		t.Errorf("Expected '%v' but returned '%v'", true, a)
	}
	b := isValidClientBodyBufferSize("1000k")
	if !b {
		t.Errorf("Expected '%v' but returned '%v'", true, b)
	}
	c := isValidClientBodyBufferSize("1000m")
	if !c {
		t.Errorf("Expected '%v' but returned '%v'", true, c)
	}
	d := isValidClientBodyBufferSize("1000km")
	if d {
		t.Errorf("Expected '%v' but returned '%v'", false, d)
	}
	e := isValidClientBodyBufferSize("1000mk")
	if e {
		t.Errorf("Expected '%v' but returned '%v'", false, e)
	}
	f := isValidClientBodyBufferSize("1000kk")
	if f {
		t.Errorf("Expected '%v' but returned '%v'", false, f)
	}
	g := isValidClientBodyBufferSize("1000mm")
	if g {
		t.Errorf("Expected '%v' but returned '%v'", false, g)
	}
	h := isValidClientBodyBufferSize(nil)
	if h {
		t.Errorf("Expected '%v' but returned '%v'", false, h)
	}
	i := isValidClientBodyBufferSize("")
	if i {
		t.Errorf("Expected '%v' but returned '%v'", false, i)
	}
}

func TestIsLocationAllowed(t *testing.T) {
	loc := ingress.Location{
		Denied: nil,
	}

	isAllowed := isLocationAllowed(&loc)
	if !isAllowed {
		t.Errorf("Expected '%v' but returned '%v'", true, isAllowed)
	}
}

func TestBuildForwardedFor(t *testing.T) {
	inputStr := "X-Forwarded-For"
	outputStr := buildForwardedFor(inputStr)

	validStr := "$http_x_forwarded_for"

	if outputStr != validStr {
		t.Errorf("Expected '%v' but returned '%v'", validStr, outputStr)
	}
}

func TestBuildResolversForLua(t *testing.T) {
	ipOne := net.ParseIP("192.0.0.1")
	ipTwo := net.ParseIP("2001:db8:1234:0000:0000:0000:0000:0000")
	ipList := []net.IP{ipOne, ipTwo}

	expected := "\"192.0.0.1\", \"2001:db8:1234::\""
	actual := buildResolversForLua(ipList, false)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}

	expected = "\"192.0.0.1\""
	actual = buildResolversForLua(ipList, true)

	if expected != actual {
		t.Errorf("Expected '%v' but returned '%v'", expected, actual)
	}
}

func TestBuildResolvers(t *testing.T) {
	ipOne := net.ParseIP("192.0.0.1")
	ipTwo := net.ParseIP("2001:db8:1234:0000:0000:0000:0000:0000")
	ipList := []net.IP{ipOne, ipTwo}

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
}

func TestBuildAuthSignURL(t *testing.T) {
	cases := map[string]struct {
		Input, Output string
	}{
		"default url":       {"http://google.com", "http://google.com?rd=$pass_access_scheme://$http_host$escaped_request_uri"},
		"with random field": {"http://google.com?cat=0", "http://google.com?cat=0&rd=$pass_access_scheme://$http_host$escaped_request_uri"},
		"with rd field":     {"http://google.com?cat&rd=$request", "http://google.com?cat&rd=$request"},
	}
	for k, tc := range cases {
		res := buildAuthSignURL(tc.Input)
		if res != tc.Output {
			t.Errorf("%s: called buildAuthSignURL('%s'); expected '%v' but returned '%v'", k, tc.Input, tc.Output, res)
		}
	}
}

func TestIsLocationInLocationList(t *testing.T) {

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
	defaultBackend := "upstream-name"
	defaultHost := "example.com"

	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:             tc.Path,
			Rewrite:          rewrite.Config{Target: tc.Target, AddBaseURL: tc.AddBaseURL, BaseURLScheme: tc.BaseURLScheme},
			Backend:          defaultBackend,
			XForwardedPrefix: tc.XForwardedPrefix,
		}

		backend := &ingress.Backend{
			Name:   defaultBackend,
			Secure: tc.SecureBackend,
		}

		expected := defaultBackend

		if tc.Sticky {
			if !tc.DynamicConfigurationEnabled {
				expected = fmt.Sprintf("sticky-" + expected)
			}

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

		pp := buildUpstreamName(defaultHost, backends, loc, tc.DynamicConfigurationEnabled)
		if !strings.EqualFold(expected, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, expected, pp)
		}
	}
}
