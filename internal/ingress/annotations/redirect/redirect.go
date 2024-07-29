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

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const defaultPermanentRedirectCode = http.StatusMovedPermanently

// Config returns the redirect configuration for an Ingress rule
type Config struct {
	URL       string `json:"url"`
	Code      int    `json:"code"`
	FromToWWW bool   `json:"fromToWWW"`
}

const (
	fromToWWWRedirAnnotation        = "from-to-www-redirect"
	temporalRedirectAnnotation      = "temporal-redirect"
	permanentRedirectAnnotation     = "permanent-redirect"
	permanentRedirectAnnotationCode = "permanent-redirect-code"
)

var redirectAnnotations = parser.Annotation{
	Group: "redirect",
	Annotations: parser.AnnotationFields{
		fromToWWWRedirAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `In some scenarios, it is required to redirect from www.domain.com to domain.com or vice versa, which way the redirect is performed depends on the configured host value in the Ingress object.`,
		},
		temporalRedirectAnnotation: {
			Validator: parser.ValidateRegex(parser.URLIsValidRegex, false),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskMedium, // Medium, as it allows arbitrary URLs that needs to be validated
			Documentation: `This annotation allows you to return a temporal redirect (Return Code 302) instead of sending data to the upstream. 
			For example setting this annotation to https://www.google.com would redirect everything to Google with a Return Code of 302 (Moved Temporarily).`,
		},
		permanentRedirectAnnotation: {
			Validator: parser.ValidateRegex(parser.URLIsValidRegex, false),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskMedium, // Medium, as it allows arbitrary URLs that needs to be validated
			Documentation: `This annotation allows to return a permanent redirect (Return Code 301) instead of sending data to the upstream. 
			For example setting this annotation https://www.google.com would redirect everything to Google with a code 301`,
		},
		permanentRedirectAnnotationCode: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `This annotation allows you to modify the status code used for permanent redirects.`,
		},
	},
}

type redirect struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new redirect annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return redirect{
		r:                r,
		annotationConfig: redirectAnnotations,
	}
}

// Parse parses the annotations contained in the ingress
// rule used to create a redirect in the paths defined in the rule.
// If the Ingress contains both annotations the execution order is
// temporal and then permanent
func (r redirect) Parse(ing *networking.Ingress) (interface{}, error) {
	r3w, err := parser.GetBoolAnnotation(fromToWWWRedirAnnotation, ing, r.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	tr, err := parser.GetStringAnnotation(temporalRedirectAnnotation, ing, r.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	if tr != "" {
		if err := isValidURL(tr); err != nil {
			return nil, err
		}

		return &Config{
			URL:       tr,
			Code:      http.StatusFound,
			FromToWWW: r3w,
		}, nil
	}

	pr, err := parser.GetStringAnnotation(permanentRedirectAnnotation, ing, r.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	prc, err := parser.GetIntAnnotation(permanentRedirectAnnotationCode, ing, r.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	if prc < http.StatusMultipleChoices || prc > http.StatusPermanentRedirect {
		prc = defaultPermanentRedirectCode
	}

	if pr != "" || r3w {
		return &Config{
			URL:       pr,
			Code:      prc,
			FromToWWW: r3w,
		}, nil
	}

	return nil, errors.ErrMissingAnnotations
}

// Equal tests for equality between two Redirect types
func (r1 *Config) Equal(r2 *Config) bool {
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

func (r redirect) GetDocumentation() parser.AnnotationFields {
	return r.annotationConfig.Annotations
}

func (r redirect) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(r.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, redirectAnnotations.Annotations)
}
