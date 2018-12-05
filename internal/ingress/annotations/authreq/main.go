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

	"k8s.io/klog"

	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns external authentication configuration for an Ingress rule
type Config struct {
	URL string `json:"url"`
	// Host contains the hostname defined in the URL
	Host            string   `json:"host"`
	SigninURL       string   `json:"signinUrl"`
	Method          string   `json:"method"`
	ResponseHeaders []string `json:"responseHeaders,omitempty"`
	RequestRedirect string   `json:"requestRedirect"`
	AuthSnippet     string   `json:"authSnippet"`
}

// Equal tests for equality between two Config types
func (e1 *Config) Equal(e2 *Config) bool {
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
	if e1.RequestRedirect != e2.RequestRedirect {
		return false
	}
	if e1.AuthSnippet != e2.AuthSnippet {
		return false
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
	r resolver.Resolver
}

// NewParser creates a new authentication request annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return authReq{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to use an Config URL as source for authentication
func (a authReq) Parse(ing *extensions.Ingress) (interface{}, error) {
	// Required Parameters
	urlString, err := parser.GetStringAnnotation("auth-url", ing)
	if err != nil {
		return nil, err
	}

	authURL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	if authURL.Scheme == "" {
		return nil, ing_errors.NewLocationDenied("url scheme is empty")
	}
	if authURL.Host == "" {
		return nil, ing_errors.NewLocationDenied("url host is empty")
	}
	if strings.Contains(authURL.Host, "..") {
		return nil, ing_errors.NewLocationDenied("invalid url host")
	}

	authMethod, _ := parser.GetStringAnnotation("auth-method", ing)
	if len(authMethod) != 0 && !validMethod(authMethod) {
		return nil, ing_errors.NewLocationDenied("invalid HTTP method")
	}

	// Optional Parameters
	signIn, err := parser.GetStringAnnotation("auth-signin", ing)
	if err != nil {
		klog.Warning("auth-signin annotation is undefined and will not be set")
	}

	authSnippet, err := parser.GetStringAnnotation("auth-snippet", ing)
	if err != nil {
		klog.Warning("auth-snippet annotation is undefined and will not be set")
	}

	responseHeaders := []string{}
	hstr, _ := parser.GetStringAnnotation("auth-response-headers", ing)
	if len(hstr) != 0 {
		harr := strings.Split(hstr, ",")
		for _, header := range harr {
			header = strings.TrimSpace(header)
			if len(header) > 0 {
				if !validHeader(header) {
					return nil, ing_errors.NewLocationDenied("invalid headers list")
				}
				responseHeaders = append(responseHeaders, header)
			}
		}
	}

	requestRedirect, _ := parser.GetStringAnnotation("auth-request-redirect", ing)

	return &Config{
		URL:             urlString,
		Host:            authURL.Hostname(),
		SigninURL:       signIn,
		Method:          authMethod,
		ResponseHeaders: responseHeaders,
		RequestRedirect: requestRedirect,
		AuthSnippet:     authSnippet,
	}, nil
}
