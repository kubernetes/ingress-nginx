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

package opentracing

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	enableOpentracingAnnotation    = "enable-opentracing"
	opentracingTrustSpanAnnotation = "opentracing-trust-incoming-span"
)

var opentracingAnnotations = parser.Annotation{
	Group: "opentracing",
	Annotations: parser.AnnotationFields{
		enableOpentracingAnnotation: {
			Validator: parser.ValidateBool,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation defines if Opentracing collector should be enable for this location. Opentracing should 
			already be configured by Ingress administrator`,
		},
		opentracingTrustSpanAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables or disables using spans from incoming requests as parent for created ones`,
		},
	},
}

type opentracing struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config contains the configuration to be used in the Ingress
type Config struct {
	Enabled      bool `json:"enabled"`
	Set          bool `json:"set"`
	TrustEnabled bool `json:"trust-enabled"`
	TrustSet     bool `json:"trust-set"`
}

// Equal tests for equality between two Config types
func (bd1 *Config) Equal(bd2 *Config) bool {
	if bd1.Set != bd2.Set {
		return false
	}

	if bd1.Enabled != bd2.Enabled {
		return false
	}

	if bd1.TrustSet != bd2.TrustSet {
		return false
	}

	if bd1.TrustEnabled != bd2.TrustEnabled {
		return false
	}

	return true
}

// NewParser creates a new serviceUpstream annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return opentracing{
		r:                r,
		annotationConfig: opentracingAnnotations,
	}
}

func (o opentracing) Parse(ing *networking.Ingress) (interface{}, error) {
	enabled, err := parser.GetBoolAnnotation(enableOpentracingAnnotation, ing, o.annotationConfig.Annotations)
	if err != nil {
		return &Config{}, nil
	}

	trustSpan, err := parser.GetBoolAnnotation(opentracingTrustSpanAnnotation, ing, o.annotationConfig.Annotations)
	if err != nil {
		return &Config{Set: true, Enabled: enabled}, nil
	}

	return &Config{Set: true, Enabled: enabled, TrustSet: true, TrustEnabled: trustSpan}, nil
}

func (o opentracing) GetDocumentation() parser.AnnotationFields {
	return o.annotationConfig.Annotations
}

func (o opentracing) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(o.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, opentracingAnnotations.Annotations)
}
