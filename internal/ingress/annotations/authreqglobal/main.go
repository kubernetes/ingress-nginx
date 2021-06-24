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

package authreqglobal

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type authReqGlobal struct {
	r resolver.Resolver
}

// NewParser creates a new authentication request annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return authReqGlobal{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to enable or disable global external authentication
func (a authReqGlobal) Parse(ing *networking.Ingress) (interface{}, error) {

	enableGlobalAuth, err := parser.GetBoolAnnotation("enable-global-auth", ing)
	if err != nil {
		enableGlobalAuth = true
	}

	return enableGlobalAuth, nil
}
