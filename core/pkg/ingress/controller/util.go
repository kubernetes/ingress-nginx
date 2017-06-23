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

package controller

import (
	"strings"
	"unicode/utf8"

	"github.com/golang/glog"
	"github.com/imdario/mergo"

	"k8s.io/ingress/core/pkg/ingress"
)

// DeniedKeyName name of the key that contains the reason to deny a location
const DeniedKeyName = "Denied"

// newDefaultServer return an BackendServer to be use as default server that returns 503.
func newDefaultServer() ingress.Endpoint {
	return ingress.Endpoint{Address: "127.0.0.1", Port: "8181"}
}

// newUpstream creates an upstream without servers.
func newUpstream(name string) *ingress.Backend {
	return &ingress.Backend{
		Name:      name,
		Endpoints: []ingress.Endpoint{},
		SessionAffinity: ingress.SessionAffinityConfig{
			CookieSessionAffinity: ingress.CookieSessionAffinity{
				Locations: make(map[string][]string),
			},
		},
	}
}

func isHostValid(host string, cert *ingress.SSLCert) bool {
	if cert == nil {
		return false
	}

	lowered := toLowerCaseASCII(host)
	for _, cn := range cert.CN {
		if matchHostnames(toLowerCaseASCII(cn), lowered) {
			return true
		}
	}

	return false
}

func matchHostnames(pattern, host string) bool {
	host = strings.TrimSuffix(host, ".")
	pattern = strings.TrimSuffix(pattern, ".")

	if len(pattern) == 0 || len(host) == 0 {
		return false
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")

	if len(patternParts) != len(hostParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return false
		}
	}

	return true
}

// toLowerCaseASCII returns a lower-case version of in. See RFC 6125 6.4.1. We use
// an explicitly ASCII function to avoid any sharp corners resulting from
// performing Unicode operations on DNS labels.
func toLowerCaseASCII(in string) string {
	// If the string is already lower-case then there's nothing to do.
	isAlreadyLowerCase := true
	for _, c := range in {
		if c == utf8.RuneError {
			// If we get a UTF-8 error then there might be
			// upper-case ASCII bytes in the invalid sequence.
			isAlreadyLowerCase = false
			break
		}
		if 'A' <= c && c <= 'Z' {
			isAlreadyLowerCase = false
			break
		}
	}

	if isAlreadyLowerCase {
		return in
	}

	out := []byte(in)
	for i, c := range out {
		if 'A' <= c && c <= 'Z' {
			out[i] += 'a' - 'A'
		}
	}
	return string(out)
}

func mergeLocationAnnotations(loc *ingress.Location, anns map[string]interface{}) {
	if _, ok := anns[DeniedKeyName]; ok {
		loc.Denied = anns[DeniedKeyName].(error)
	}
	delete(anns, DeniedKeyName)
	err := mergo.Map(loc, anns)
	if err != nil {
		glog.Errorf("unexpected error merging extracted annotations in location type: %v", err)
	}
}
