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

const (
	enableGlobalAuthAnnotation = "enable-global-auth"
)

var globalAuthAnnotations = parser.Annotation{
	Group: "authentication",
	Annotations: parser.AnnotationFields{
		enableGlobalAuthAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `Defines if the global external authentication should be enabled.`,
		},
	},
}

type authReqGlobal struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new authentication request annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return authReqGlobal{
		r:                r,
		annotationConfig: globalAuthAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to enable or disable global external authentication
func (a authReqGlobal) Parse(ing *networking.Ingress) (interface{}, error) {
	enableGlobalAuth, err := parser.GetBoolAnnotation(enableGlobalAuthAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		enableGlobalAuth = true
	}

	return enableGlobalAuth, nil
}

func (a authReqGlobal) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a authReqGlobal) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, globalAuthAnnotations.Annotations)
}
