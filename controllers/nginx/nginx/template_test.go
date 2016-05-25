/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package nginx

import (
	"testing"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/rewrite"
)

var (
	tmplFuncTestcases = map[string]struct {
		Path      string
		To        string
		Location  string
		ProxyPass string
		Rewrite   bool
	}{
		"invalid redirect / to /": {"/", "/", "/", "proxy_pass http://upstream-name;", false},
		"redirect / to /jenkins": {"/", "/jenkins", "~* /",
			`rewrite /(.*) /jenkins/$1 break;
proxy_pass http://upstream-name;
`, false},
		"redirect /something to /": {"/something", "/", "~* /something/", `rewrite /something/(.*) /$1 break;
proxy_pass http://upstream-name;
`, false},
		"redirect /something-complex to /not-root": {"/something-complex", "/not-root", "~* /something-complex/", `rewrite /something-complex/(.*) /not-root/$1 break;
proxy_pass http://upstream-name;
`, false},
		"redirect / to /jenkins and rewrite": {"/", "/jenkins", "~* /",
			`rewrite /(.*) /jenkins/$1 break;
proxy_pass http://upstream-name;
sub_filter "//$host/" "//$host/jenkins";
sub_filter_once off;`, true},
		"redirect /something to / and rewrite": {"/something", "/", "~* /something/", `rewrite /something/(.*) /$1 break;
proxy_pass http://upstream-name;
sub_filter "//$host/something" "//$host/";
sub_filter_once off;`, true},
		"redirect /something-complex to /not-root and rewrite": {"/something-complex", "/not-root", "~* /something-complex/", `rewrite /something-complex/(.*) /not-root/$1 break;
proxy_pass http://upstream-name;
sub_filter "//$host/something-complex" "//$host/not-root";
sub_filter_once off;`, true},
	}
)

func TestBuildLocation(t *testing.T) {
	for k, tc := range tmplFuncTestcases {
		loc := &Location{
			Path:     tc.Path,
			Redirect: rewrite.Redirect{tc.To, tc.Rewrite},
		}

		newLoc := buildLocation(loc)
		if tc.Location != newLoc {
			t.Errorf("%s: expected %v but returned %v", k, tc.Location, newLoc)
		}
	}
}

func TestBuildProxyPass(t *testing.T) {
	for k, tc := range tmplFuncTestcases {
		loc := &Location{
			Path:     tc.Path,
			Redirect: rewrite.Redirect{tc.To, tc.Rewrite},
			Upstream: Upstream{Name: "upstream-name"},
		}

		pp := buildProxyPass(loc)
		if tc.ProxyPass != pp {
			t.Errorf("%s: expected \n%v \nbut returned \n%v", k, tc.ProxyPass, pp)
		}
	}
}
