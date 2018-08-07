/*
Copyright 2017 The Kubernetes Authors.

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

package trustxforwardedfor

import (
	"regexp"

	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var trustTypeRegex = regexp.MustCompile(`first|last`)

type trustXForwardedfor struct {
	r resolver.Resolver
}

// NewParser creates a new trustXForwardedfor annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return trustXForwardedfor{r}
}

// Parse parses the annotations contained in the ingress rule
// used to decide if to trust x-forwarded-for to whitelist rules and which of
// the IP in the list (first or last)
func (a trustXForwardedfor) Parse(ing *extensions.Ingress) (interface{}, error) {
	t, err := parser.GetStringAnnotation("trust-x-forwarded-for", ing)
	if err != nil {
		return nil, err
	}

	if !trustTypeRegex.MatchString(t) {
		return nil, ing_errors.NewLocationDenied("invalid x-forwarded-for trust type")
	}

	return t, err
}
