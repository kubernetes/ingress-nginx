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

package defaultbackend

import (
	"fmt"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	defaultBackendAnnotation = "default-backend"
)

var defaultBackendAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		defaultBackendAnnotation: {
			Validator: parser.ValidateServiceName,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This service will be used to handle the response when the configured service in the Ingress rule does not have any active endpoints. 
			It will also be used to handle the error responses if both this annotation and the custom-http-errors annotation are set.`,
		},
	},
}

type backend struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new default backend annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return backend{
		r:                r,
		annotationConfig: defaultBackendAnnotations,
	}
}

// Parse parses the annotations contained in the ingress to use
// a custom default backend
func (db backend) Parse(ing *networking.Ingress) (interface{}, error) {
	s, err := parser.GetStringAnnotation(defaultBackendAnnotation, ing, db.annotationConfig.Annotations)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%v/%v", ing.Namespace, s)
	svc, err := db.r.GetService(name)
	if err != nil {
		return nil, fmt.Errorf("unexpected error reading service %s: %w", name, err)
	}

	return svc, nil
}

func (db backend) GetDocumentation() parser.AnnotationFields {
	return db.annotationConfig.Annotations
}

func (a backend) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRisk)
	return parser.CheckAnnotationRisk(anns, maxrisk, defaultBackendAnnotations.Annotations)
}
