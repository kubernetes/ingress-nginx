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

package serviceupstream

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	serviceUpstreamAnnotation = "service-upstream"
)

var serviceUpstreamAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		serviceUpstreamAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow, // Critical, this annotation is not validated at all and allows arbitrary configutations
			Documentation: `This annotation makes NGINX use Service's Cluster IP and Port instead of Endpoints as the backend endpoints`,
		},
	},
}

type serviceUpstream struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new serviceUpstream annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return serviceUpstream{
		r:                r,
		annotationConfig: serviceUpstreamAnnotations,
	}
}

func (s serviceUpstream) Parse(ing *networking.Ingress) (interface{}, error) {
	defBackend := s.r.GetDefaultBackend()

	val, err := parser.GetBoolAnnotation(serviceUpstreamAnnotation, ing, s.annotationConfig.Annotations)
	// A missing annotation is not a problem, just use the default
	if err == errors.ErrMissingAnnotations {
		return defBackend.ServiceUpstream, nil
	}

	return val, nil
}

func (s serviceUpstream) GetDocumentation() parser.AnnotationFields {
	return s.annotationConfig.Annotations
}

func (s serviceUpstream) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(s.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, serviceUpstreamAnnotations.Annotations)
}
