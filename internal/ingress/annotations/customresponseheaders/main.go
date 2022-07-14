/*
Copyright 2023 The Kubernetes Authors.

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

package customresponseheaders

import (
	"reflect"
	"regexp"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"

	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var (
	headerRegexp = regexp.MustCompile(`^[a-zA-Z\d\-_]+$`)
	//Regex below requires review. Following regex has been picked up from PR: https://github.com/kubernetes/ingress-nginx/pull/9742
	headerValueRegexp    = regexp.MustCompile(`^[a-zA-Z\d_ :;.,\\/"'?!(){}\[\]@<>=\-+*#$&\x60|~^%]+$`)
	completeHeaderRegexp = regexp.MustCompile(`^[a-zA-Z\d_ :;.,\\/\011\012"'?!(){}\[\]@<>=\-+*#$&\x60|~^%]+$`)
)

const (
	customResponseHeadersAnnotation = "custom-response-headers"
)

var customResponseHeadersAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		customResponseHeadersAnnotation: {
			Validator:     parser.ValidateRegex(completeHeaderRegexp, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation will allows setting the custom response headers for the given ingress`,
		},
	},
}

// Config returns the custom response headers for an Ingress rule
type Config struct {
	ResponseHeaders map[string]string `json:"custom-response-headers,omitempty"`
}

type customresponseheaders struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new custom response headers annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return customresponseheaders{
		r:                r,
		annotationConfig: customResponseHeadersAnnotations,
	}
}

// Equal tests for equality between two Configuration types
func (l1 *Config) Equal(l2 *Config) bool {
	if l1 == l2 {
		return true
	}

	if l1 == nil || l2 == nil {
		return false
	}

	return reflect.DeepEqual(l1.ResponseHeaders, l2.ResponseHeaders)
}

// Parse parses the annotations contained in the ingress to use
// custom response headers
func (e customresponseheaders) Parse(ing *networking.Ingress) (interface{}, error) {
	headersMap := map[string]string{}
	responseHeader, err := parser.GetStringAnnotation(customResponseHeadersAnnotation, ing, e.annotationConfig.Annotations)
	if err != nil {
		return nil, err
	}

	headers := strings.Split(responseHeader, "\n")
	for i := 0; i < len(headers); i++ {
		if len(headers[i]) == 0 {
			continue
		}

		if !strings.Contains(headers[i], ":") {
			return nil, ing_errors.NewLocationDenied("Invalid header format")
		}

		headerSplit := strings.SplitN(headers[i], ":", 2)
		for j := range headerSplit {
			headerSplit[j] = strings.TrimSpace(headerSplit[j])
		}

		if len(headerSplit) < 2 {
			return nil, ing_errors.NewLocationDenied("Invalid header size")
		}

		if !ValidHeader(headerSplit[0]) {
			return nil, ing_errors.NewLocationDenied("Invalid header name")
		}

		if !ValidValue(headerSplit[1]) {
			return nil, ing_errors.NewLocationDenied("Invalid header value")
		}

		headersMap[strings.TrimSpace(headerSplit[0])] = strings.TrimSpace(headerSplit[1])
	}
	return &Config{headersMap}, nil
}

// ValidHeader checks is the provided string satisfies the header's name regex
func ValidHeader(header string) bool {
	return headerRegexp.Match([]byte(header))
}

// ValidValue checks if the provided string satisfies the header value regex
func ValidValue(header string) bool {
	return headerValueRegexp.MatchString(header)
}

func (e customresponseheaders) GetDocumentation() parser.AnnotationFields {
	return e.annotationConfig.Annotations
}

func (a customresponseheaders) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, customResponseHeadersAnnotations.Annotations)
}
