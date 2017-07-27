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
	"regexp"
	"strings"

	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
)

const (
	// external URL that provides the authentication
	authURL       = "ingress.kubernetes.io/auth-url"
	authSigninURL = "ingress.kubernetes.io/auth-signin"
	authMethod    = "ingress.kubernetes.io/auth-method"
	authBody      = "ingress.kubernetes.io/auth-send-body"
	authHeaders   = "ingress.kubernetes.io/auth-response-headers"
)

// External returns external authentication configuration for an Ingress rule
type External struct {
	URL string `json:"url"`
	// Host contains the hostname defined in the URL
	Host            string   `json:"host"`
	SigninURL       string   `json:"signinUrl"`
	Method          string   `json:"method"`
	SendBody        bool     `json:"sendBody"`
	ResponseHeaders []string `json:"responseHeaders,omitEmpty"`
}

// Equal tests for equality between two External types
func (e1 *External) Equal(e2 *External) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.URL != e2.URL {
		return false
	}
	if e1.Host != e2.Host {
		return false
	}
	if e1.SigninURL != e2.SigninURL {
		return false
	}
	if e1.Method != e2.Method {
		return false
	}
	if e1.SendBody != e2.SendBody {
		return false
	}
	if e1.Method != e2.Method {
		return false
	}

	for _, ep1 := range e1.ResponseHeaders {
		found := false
		for _, ep2 := range e2.ResponseHeaders {
			if ep1 == ep2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

var (
	methods      = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
	headerRegexp = regexp.MustCompile(`^[a-zA-Z\d\-_]+$`)
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

func validHeader(header string) bool {
	return headerRegexp.Match([]byte(header))
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

	signin, _ := parser.GetStringAnnotation(authSigninURL, ing)

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

	h := []string{}
	hstr, _ := parser.GetStringAnnotation(authHeaders, ing)
	if len(hstr) != 0 {

		harr := strings.Split(hstr, ",")
		for _, header := range harr {
			header := strings.TrimSpace(header)
			if len(header) > 0 {
				if !validHeader(header) {
					return nil, ing_errors.NewLocationDenied("invalid headers list")
				}
				h = append(h, header)
			}
		}
	}

	sb, _ := parser.GetBoolAnnotation(authBody, ing)

	return &External{
		URL:             str,
		Host:            ur.Hostname(),
		SigninURL:       signin,
		Method:          m,
		SendBody:        sb,
		ResponseHeaders: h,
	}, nil
}
