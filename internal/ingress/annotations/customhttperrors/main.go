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

package customhttperrors

import (
	"regexp"
	"strconv"
	"strings"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	customHTTPErrorsAnnotation = "custom-http-errors"
)

// We accept anything between 400 and 599, on a comma separated.
var arrayOfHTTPErrors = regexp.MustCompile(`^(?:[4,5]\d{2},?)*$`)

var customHTTPErrorsAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		customHTTPErrorsAnnotation: {
			Validator: parser.ValidateRegex(arrayOfHTTPErrors, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `If a default backend annotation is specified on the ingress, the errors code specified on this annotation 
			will be routed to that annotation's default backend service. Otherwise they will be routed to the global default backend.
			A comma-separated list of error codes is accepted (anything between 400 and 599, like 403, 503)`,
		},
	},
}

type customhttperrors struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new custom http errors annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return customhttperrors{
		r:                r,
		annotationConfig: customHTTPErrorsAnnotations,
	}
}

// Parse parses the annotations contained in the ingress to use
// custom http errors
func (e customhttperrors) Parse(ing *networking.Ingress) (interface{}, error) {
	c, err := parser.GetStringAnnotation(customHTTPErrorsAnnotation, ing, e.annotationConfig.Annotations)
	if err != nil {
		return nil, err
	}

	cSplit := strings.Split(c, ",")
	codes := make([]int, 0, len(cSplit))
	for _, i := range cSplit {
		num, err := strconv.Atoi(i)
		if err != nil {
			return nil, err
		}
		codes = append(codes, num)
	}

	return codes, nil
}

func (e customhttperrors) GetDocumentation() parser.AnnotationFields {
	return e.annotationConfig.Annotations
}

func (e customhttperrors) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(e.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, customHTTPErrorsAnnotations.Annotations)
}
