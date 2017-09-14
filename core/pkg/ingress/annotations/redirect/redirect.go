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

package redirect

import (
	"net/http"
	"net/url"
	"strings"

	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/errors"
)

const (
	permanent = "ingress.kubernetes.io/permanent-redirect"
	temporal  = "ingress.kubernetes.io/temporal-redirect"
	www       = "ingress.kubernetes.io/from-to-www-redirect"
)

// Redirect returns the redirect configuration for an Ingress rule
type Redirect struct {
	URL       string `json:"url"`
	Code      int    `json:"code"`
	FromToWWW bool   `json:"fromToWWW"`
}

type redirect struct{}

// NewParser creates a new redirect annotation parser
func NewParser() parser.IngressAnnotation {
	return redirect{}
}

// Parse parses the annotations contained in the ingress
// rule used to create a redirect in the paths defined in the rule.
// If the Ingress containes both annotations the execution order is
// temporal and then permanent
func (a redirect) Parse(ing *extensions.Ingress) (interface{}, error) {
	r3w, _ := parser.GetBoolAnnotation(www, ing)

	tr, err := parser.GetStringAnnotation(temporal, ing)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	if tr != "" {
		if err := isValidURL(tr); err != nil {
			return nil, err
		}

		return &Redirect{
			URL:       tr,
			Code:      http.StatusFound,
			FromToWWW: r3w,
		}, nil
	}

	pr, err := parser.GetStringAnnotation(permanent, ing)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	if pr != "" {
		if err := isValidURL(pr); err != nil {
			return nil, err
		}

		return &Redirect{
			URL:       pr,
			Code:      http.StatusMovedPermanently,
			FromToWWW: r3w,
		}, nil
	}

	if r3w {
		return &Redirect{
			FromToWWW: r3w,
		}, nil
	}

	return nil, errors.ErrMissingAnnotations
}

// Equal tests for equality between two Redirect types
func (r1 *Redirect) Equal(r2 *Redirect) bool {
	if r1 == r2 {
		return true
	}
	if r1 == nil || r2 == nil {
		return false
	}
	if r1.URL != r2.URL {
		return false
	}
	if r1.Code != r2.Code {
		return false
	}
	if r1.FromToWWW != r2.FromToWWW {
		return false
	}
	return true
}

func isValidURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(u.Scheme, "http") {
		return errors.Errorf("only http and https are valid protocols (%v)", u.Scheme)
	}

	return nil
}
