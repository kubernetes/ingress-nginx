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
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"io/ioutil"

	"k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/annotations/authreq"
	"k8s.io/ingress/core/pkg/ingress/annotations/rewrite"
)

var (
	// TODO: add tests for secure endpoints
	tmplFuncTestcases = map[string]struct {
		Path       string
		Target     string
		Location   string
		ProxyPass  string
		AddBaseURL bool
	}{
		"invalid redirect / to /": {"/", "/", "/", "proxy_pass http://upstream-name;", false},
		"redirect / to /jenkins": {"/", "/jenkins", "~* /",
			`
	rewrite /(.*) /jenkins/$1 break;
	proxy_pass http://upstream-name;
	`, false},
		"redirect /something to /": {"/something", "/", `~* ^/something\/?(?<baseuri>.*)`, `
	rewrite /something/(.*) /$1 break;
	rewrite /something / break;
	proxy_pass http://upstream-name;
	`, false},
		"redirect /end-with-slash/ to /not-root": {"/end-with-slash/", "/not-root", "~* ^/end-with-slash/(?<baseuri>.*)", `
	rewrite /end-with-slash/(.*) /not-root/$1 break;
	proxy_pass http://upstream-name;
	`, false},
		"redirect /something-complex to /not-root": {"/something-complex", "/not-root", `~* ^/something-complex\/?(?<baseuri>.*)`, `
	rewrite /something-complex/(.*) /not-root/$1 break;
	proxy_pass http://upstream-name;
	`, false},
		"redirect / to /jenkins and rewrite": {"/", "/jenkins", "~* /", `
	rewrite /(.*) /jenkins/$1 break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$http_host/$baseuri">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$http_host/$baseuri">' r;
	`, true},
		"redirect /something to / and rewrite": {"/something", "/", `~* ^/something\/?(?<baseuri>.*)`, `
	rewrite /something/(.*) /$1 break;
	rewrite /something / break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$http_host/something/$baseuri">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$http_host/something/$baseuri">' r;
	`, true},
		"redirect /end-with-slash/ to /not-root and rewrite": {"/end-with-slash/", "/not-root", `~* ^/end-with-slash/(?<baseuri>.*)`, `
	rewrite /end-with-slash/(.*) /not-root/$1 break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$http_host/end-with-slash/$baseuri">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$http_host/end-with-slash/$baseuri">' r;
	`, true},
		"redirect /something-complex to /not-root and rewrite": {"/something-complex", "/not-root", `~* ^/something-complex\/?(?<baseuri>.*)`, `
	rewrite /something-complex/(.*) /not-root/$1 break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$http_host/something-complex/$baseuri">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$http_host/something-complex/$baseuri">' r;
	`, true},
	}
)

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
			Path:     tc.Path,
			Redirect: rewrite.Redirect{Target: tc.Target, AddBaseURL: tc.AddBaseURL},
		}

		newLoc := buildLocation(loc)
		if tc.Location != newLoc {
			t.Errorf("%s: expected '%v' but returned %v", k, tc.Location, newLoc)
		}
	}
}

func TestBuildProxyPass(t *testing.T) {
	for k, tc := range tmplFuncTestcases {
		loc := &ingress.Location{
			Path:     tc.Path,
			Redirect: rewrite.Redirect{Target: tc.Target, AddBaseURL: tc.AddBaseURL},
			Backend:  "upstream-name",
		}

		pp := buildProxyPass("", []*ingress.Backend{}, loc)
		if !strings.EqualFold(tc.ProxyPass, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, tc.ProxyPass, pp)
		}
	}
}

func TestBuildAuthResponseHeaders(t *testing.T) {
	loc := &ingress.Location{
		ExternalAuth: authreq.External{ResponseHeaders: []string{"h1", "H-With-Caps-And-Dashes"}},
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
	f, err := os.Open(path.Join(pwd, "../../test/data/config.json"))
	if err != nil {
		t.Errorf("unexpected error reading json file: %v", err)
	}
	defer f.Close()
	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Error("unexpected error reading json file: ", err)
	}
	var dat config.TemplateConfig
	if err := json.Unmarshal(data, &dat); err != nil {
		t.Errorf("unexpected error unmarshalling json: %v", err)
	}

	tf, err := os.Open(path.Join(pwd, "../../rootfs/etc/nginx/template/nginx.tmpl"))
	if err != nil {
		t.Errorf("unexpected error reading json file: %v", err)
	}
	defer tf.Close()

	ngxTpl, err := NewTemplate(tf.Name(), func() {})
	if err != nil {
		t.Errorf("invalid NGINX template: %v", err)
	}

	_, err = ngxTpl.Write(dat)
	if err != nil {
		t.Errorf("invalid NGINX template: %v", err)
	}
}

func BenchmarkTemplateWithData(b *testing.B) {
	pwd, _ := os.Getwd()
	f, err := os.Open(path.Join(pwd, "../../test/data/config.json"))
	if err != nil {
		b.Errorf("unexpected error reading json file: %v", err)
	}
	defer f.Close()
	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		b.Error("unexpected error reading json file: ", err)
	}
	var dat config.TemplateConfig
	if err := json.Unmarshal(data, &dat); err != nil {
		b.Errorf("unexpected error unmarshalling json: %v", err)
	}

	tf, err := os.Open(path.Join(pwd, "../../rootfs/etc/nginx/template/nginx.tmpl"))
	if err != nil {
		b.Errorf("unexpected error reading json file: %v", err)
	}
	defer tf.Close()

	ngxTpl, err := NewTemplate(tf.Name(), func() {})
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
