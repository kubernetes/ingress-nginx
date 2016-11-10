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
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	// external URL that provides the authentication
	authURL    = "ingress.kubernetes.io/auth-url"
	authMethod = "ingress.kubernetes.io/auth-method"
	authBody   = "ingress.kubernetes.io/auth-send-body"
)

var (
	// ErrMissingAnnotations is returned when the ingress rule
	// does not contain annotations related with authentication
	ErrMissingAnnotations = errors.New("missing authentication annotations")
)

// Auth returns external authentication configuration for an Ingress rule
type Auth struct {
	URL      string
	Method   string
	SendBody bool
}

type ingAnnotations map[string]string

func (a ingAnnotations) url() (string, error) {
	val, ok := a[authURL]
	if !ok {
		return "", ErrMissingAnnotations
	}

	return val, nil
}

func (a ingAnnotations) method() string {
	val, ok := a[authMethod]
	if !ok {
		return ""
	}

	return val
}

func (a ingAnnotations) sendBody() bool {
	val, ok := a[authBody]
	if ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return false
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

// ParseAnnotations parses the annotations contained in the ingress
// rule used to use an external URL as source for authentication
func ParseAnnotations(ing *extensions.Ingress) (Auth, error) {
	if ing.GetAnnotations() == nil {
		return Auth{}, ErrMissingAnnotations
	}

	str, err := ingAnnotations(ing.GetAnnotations()).url()
	if err != nil {
		return Auth{}, err
	}
	if str == "" {
		return Auth{}, fmt.Errorf("an empty string is not a valid URL")
	}

	ur, err := url.Parse(str)
	if err != nil {
		return Auth{}, err
	}
	if ur.Scheme == "" {
		return Auth{}, fmt.Errorf("url scheme is empty")
	}
	if ur.Host == "" {
		return Auth{}, fmt.Errorf("url host is empty")
	}

	if strings.Index(ur.Host, "..") != -1 {
		return Auth{}, fmt.Errorf("invalid url host")
	}

	m := ingAnnotations(ing.GetAnnotations()).method()
	if len(m) != 0 && !validMethod(m) {
		return Auth{}, fmt.Errorf("invalid HTTP method")
	}

	sb := ingAnnotations(ing.GetAnnotations()).sendBody()

	return Auth{
		URL:      str,
		Method:   m,
		SendBody: sb,
	}, nil
}
