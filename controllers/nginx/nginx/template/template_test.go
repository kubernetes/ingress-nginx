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
	"strings"
	"testing"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/ingress"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/rewrite"
)

var (
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
		"redirect /something to /": {"/something", "/", "~* /something", `
	rewrite /something/(.*) /$1 break;
	rewrite /something / break;
	proxy_pass http://upstream-name;
	`, false},
		"redirect /something-complex to /not-root": {"/something-complex", "/not-root", "~* /something-complex", `
	rewrite /something-complex/(.*) /not-root/$1 break;
	proxy_pass http://upstream-name;
	`, false},
		"redirect / to /jenkins and rewrite": {"/", "/jenkins", "~* /", `
	rewrite /(.*) /jenkins/$1 break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$server_name/jenkins/">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$server_name/jenkins/">' r;
	`, true},
		"redirect /something to / and rewrite": {"/something", "/", "~* /something", `
	rewrite /something/(.*) /$1 break;
	rewrite /something / break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$server_name/">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$server_name/">' r;
	`, true},
		"redirect /something-complex to /not-root and rewrite": {"/something-complex", "/not-root", "~* /something-complex", `
	rewrite /something-complex/(.*) /not-root/$1 break;
	proxy_pass http://upstream-name;
	subs_filter '<head(.*)>' '<head$1><base href="$scheme://$server_name/not-root/">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$server_name/not-root/">' r;
	`, true},
	}
)

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
			Upstream: ingress.Upstream{Name: "upstream-name"},
		}

		pp := buildProxyPass(loc)
		if !strings.EqualFold(tc.ProxyPass, pp) {
			t.Errorf("%s: expected \n'%v'\nbut returned \n'%v'", k, tc.ProxyPass, pp)
		}
	}
}
