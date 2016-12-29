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
	"net/url"
	"strings"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
)

const (
	// external URL that provides the authentication
	authURL    = "ingress.kubernetes.io/auth-url"
	authMethod = "ingress.kubernetes.io/auth-method"
	authBody   = "ingress.kubernetes.io/auth-send-body"
)

// External returns external authentication configuration for an Ingress rule
type External struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	SendBody bool   `json:"sendBody"`
}

var (
	methods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
)

func validMethod(method string) bool {
	if len(method) == 0 {
		return false
	}

	for _, m := range methods {
		if method == m {
			return true
		}
	}
	return false
}

type authReq struct {
}

// NewParser creates a new authentication request annotation parser
func NewParser() parser.IngressAnnotation {
	return authReq{}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to use an external URL as source for authentication
func (a authReq) Parse(ing *extensions.Ingress) (interface{}, error) {
	str, err := parser.GetStringAnnotation(authURL, ing)
	if err != nil {
		return nil, err
	}

	if str == "" {
		return nil, ing_errors.NewLocationDenied("an empty string is not a valid URL")
	}

	ur, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	if ur.Scheme == "" {
		return nil, ing_errors.NewLocationDenied("url scheme is empty")
	}
	if ur.Host == "" {
		return nil, ing_errors.NewLocationDenied("url host is empty")
	}

	if strings.Contains(ur.Host, "..") {
		return nil, ing_errors.NewLocationDenied("invalid url host")
	}

	m, _ := parser.GetStringAnnotation(authMethod, ing)
	if len(m) != 0 && !validMethod(m) {
		return nil, ing_errors.NewLocationDenied("invalid HTTP method")
	}

	sb, _ := parser.GetBoolAnnotation(authBody, ing)

	return &External{
		URL:      str,
		Method:   m,
		SendBody: sb,
	}, nil
}
