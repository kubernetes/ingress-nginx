/*
Copyright 2018 The Kubernetes Authors.

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

package log

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	enableAccessLogAnnotation  = "enable-access-log"
	enableRewriteLogAnnotation = "enable-rewrite-log"
)

var logAnnotations = parser.Annotation{
	Group: "log",
	Annotations: parser.AnnotationFields{
		enableAccessLogAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This configuration setting allows you to control if this location should generate an access_log`,
		},
		enableRewriteLogAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This configuration setting allows you to control if this location should generate logs from the rewrite feature usage`,
		},
	},
}

type log struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config contains the configuration to be used in the Ingress
type Config struct {
	Access  bool `json:"accessLog"`
	Rewrite bool `json:"rewriteLog"`
}

// Equal tests for equality between two Config types
func (bd1 *Config) Equal(bd2 *Config) bool {
	if bd1.Access != bd2.Access {
		return false
	}

	if bd1.Rewrite != bd2.Rewrite {
		return false
	}

	return true
}

// NewParser creates a new log annotations parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return log{
		r:                r,
		annotationConfig: logAnnotations,
	}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should enable logs
func (l log) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	config.Access, err = parser.GetBoolAnnotation(enableAccessLogAnnotation, ing, l.annotationConfig.Annotations)
	if err != nil {
		config.Access = true
	}

	config.Rewrite, err = parser.GetBoolAnnotation(enableRewriteLogAnnotation, ing, l.annotationConfig.Annotations)
	if err != nil {
		config.Rewrite = false
	}

	return config, nil
}

func (l log) GetDocumentation() parser.AnnotationFields {
	return l.annotationConfig.Annotations
}

func (l log) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(l.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, logAnnotations.Annotations)
}
