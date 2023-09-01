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

package modsecurity

import (
	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/klog/v2"
)

const (
	modsecEnableAnnotation          = "enable-modsecurity"
	modsecEnableOwaspCoreAnnotation = "enable-owasp-core-rules"
	modesecTransactionIDAnnotation  = "modsecurity-transaction-id"
	modsecSnippetAnnotation         = "modsecurity-snippet"
)

var modsecurityAnnotation = parser.Annotation{
	Group: "modsecurity",
	Annotations: parser.AnnotationFields{
		modsecEnableAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables ModSecurity`,
		},
		modsecEnableOwaspCoreAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables the OWASP Core Rule Set`,
		},
		modesecTransactionIDAnnotation: {
			Validator:     parser.ValidateRegex(parser.NGINXVariable, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation enables passing an NGINX variable to ModSecurity.`,
		},
		modsecSnippetAnnotation: {
			Validator:     parser.ValidateNull,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskCritical,
			Documentation: `This annotation enables adding a specific snippet configuration for ModSecurity`,
		},
	},
}

// Config contains ModSecurity Configuration items
type Config struct {
	Enable        bool   `json:"enable-modsecurity"`
	EnableSet     bool   `json:"enable-modsecurity-set"`
	OWASPRules    bool   `json:"enable-owasp-core-rules"`
	TransactionID string `json:"modsecurity-transaction-id"`
	Snippet       string `json:"modsecurity-snippet"`
}

// Equal tests for equality between two Config types
func (modsec1 *Config) Equal(modsec2 *Config) bool {
	if modsec1 == modsec2 {
		return true
	}
	if modsec1 == nil || modsec2 == nil {
		return false
	}
	if modsec1.Enable != modsec2.Enable {
		return false
	}
	if modsec1.EnableSet != modsec2.EnableSet {
		return false
	}
	if modsec1.OWASPRules != modsec2.OWASPRules {
		return false
	}
	if modsec1.TransactionID != modsec2.TransactionID {
		return false
	}
	if modsec1.Snippet != modsec2.Snippet {
		return false
	}

	return true
}

// NewParser creates a new ModSecurity annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return modSecurity{
		r:                r,
		annotationConfig: modsecurityAnnotation,
	}
}

type modSecurity struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Parse parses the annotations contained in the ingress
// rule used to enable ModSecurity in a particular location
func (a modSecurity) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	config.EnableSet = true
	config.Enable, err = parser.GetBoolAnnotation(modsecEnableAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsInvalidContent(err) {
			klog.Warningf("annotation %s contains invalid directive, defaulting to false", modsecEnableAnnotation)
		}
		config.Enable = false
		config.EnableSet = false
	}

	config.OWASPRules, err = parser.GetBoolAnnotation(modsecEnableOwaspCoreAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsInvalidContent(err) {
			klog.Warningf("annotation %s contains invalid directive, defaulting to false", modsecEnableOwaspCoreAnnotation)
		}
		config.OWASPRules = false
	}

	config.TransactionID, err = parser.GetStringAnnotation(modesecTransactionIDAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsInvalidContent(err) {
			klog.Warningf("annotation %s contains invalid directive, defaulting", modesecTransactionIDAnnotation)
		}
		config.TransactionID = ""
	}

	config.Snippet, err = parser.GetStringAnnotation("modsecurity-snippet", ing, a.annotationConfig.Annotations)
	if err != nil {
		config.Snippet = ""
	}

	return config, nil
}

func (a modSecurity) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a modSecurity) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, modsecurityAnnotation.Annotations)
}
