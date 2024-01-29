/*
Copyright 2019 The Kubernetes Authors.

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

package satisfy

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	satisfyAnnotation = "satisfy"
)

var satisfyAnnotations = parser.Annotation{
	Group: "authentication",
	Annotations: parser.AnnotationFields{
		satisfyAnnotation: {
			Validator: parser.ValidateOptions([]string{"any", "all"}, true, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `By default, a request would need to satisfy all authentication requirements in order to be allowed. 
			By using this annotation, requests that satisfy either any or all authentication requirements are allowed, based on the configuration value.
			Valid options are "all" and "any"`,
		},
	},
}

type satisfy struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new SATISFY annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return satisfy{
		r:                r,
		annotationConfig: satisfyAnnotations,
	}
}

// Parse parses annotation contained in the ingress
func (s satisfy) Parse(ing *networking.Ingress) (interface{}, error) {
	satisfy, err := parser.GetStringAnnotation(satisfyAnnotation, ing, s.annotationConfig.Annotations)

	if err != nil || (satisfy != "any" && satisfy != "all") {
		satisfy = ""
	}

	return satisfy, nil
}

func (s satisfy) GetDocumentation() parser.AnnotationFields {
	return s.annotationConfig.Annotations
}

func (s satisfy) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(s.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, satisfyAnnotations.Annotations)
}
