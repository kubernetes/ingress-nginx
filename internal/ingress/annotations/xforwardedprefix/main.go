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

package xforwardedprefix

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	xForwardedForPrefixAnnotation = "x-forwarded-prefix"
)

var xForwardedForAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		xForwardedForPrefixAnnotation: {
			Validator:     parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows regexes but on a very limited set
			Documentation: `This annotation can be used to add the non-standard X-Forwarded-Prefix header to the upstream request with a string value`,
		},
	},
}

type xforwardedprefix struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new xforwardedprefix annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return xforwardedprefix{
		r:                r,
		annotationConfig: xForwardedForAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
// used to add an x-forwarded-prefix header to the request
func (x xforwardedprefix) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation(xForwardedForPrefixAnnotation, ing, x.annotationConfig.Annotations)
}

func (x xforwardedprefix) GetDocumentation() parser.AnnotationFields {
	return x.annotationConfig.Annotations
}

func (x xforwardedprefix) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(x.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, xForwardedForAnnotations.Annotations)
}
