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

package canary

import (
	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	canaryAnnotation                = "canary"
	canaryWeightAnnotation          = "canary-weight"
	canaryWeightTotalAnnotation     = "canary-weight-total"
	canaryByHeaderAnnotation        = "canary-by-header"
	canaryByHeaderValueAnnotation   = "canary-by-header-value"
	canaryByHeaderPatternAnnotation = "canary-by-header-pattern"
	canaryByCookieAnnotation        = "canary-by-cookie"
)

var CanaryAnnotations = parser.Annotation{
	Group: "canary",
	Annotations: parser.AnnotationFields{
		canaryAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables the Ingress spec to act as an alternative service for requests to route to depending on the rules applied`,
		},
		canaryWeightAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines the integer based (0 - ) percent of random requests that should be routed to the service specified in the canary Ingress`,
		},
		canaryWeightTotalAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation The total weight of traffic. If unspecified, it defaults to 100`,
		},
		canaryByHeaderAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the header that should be used for notifying the Ingress to route the request to the service specified in the Canary Ingress.
			When the request header is set to 'always', it will be routed to the canary. When the header is set to 'never', it will never be routed to the canary.
			For any other value, the header will be ignored and the request compared against the other canary rules by precedence`,
		},
		canaryByHeaderValueAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the header value to match for notifying the Ingress to route the request to the service specified in the Canary Ingress. 
			When the request header is set to this value, it will be routed to the canary. For any other header value, the header will be ignored and the request compared against the other canary rules by precedence. 
			This annotation has to be used together with 'canary-by-header'. The annotation is an extension of the 'canary-by-header' to allow customizing the header value instead of using hardcoded values. 
			It doesn't have any effect if the 'canary-by-header' annotation is not defined`,
		},
		canaryByHeaderPatternAnnotation: {
			Validator: parser.ValidateRegex(parser.IsValidRegex, false),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation works the same way as canary-by-header-value except it does PCRE Regex matching. 
			Note that when 'canary-by-header-value' is set this annotation will be ignored. 
			When the given Regex causes error during request processing, the request will be considered as not matching.`,
		},
		canaryByCookieAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the cookie that should be used for notifying the Ingress to route the request to the service specified in the Canary Ingress.
			When the cookie is set to 'always', it will be routed to the canary. When the cookie is set to 'never', it will never be routed to the canary`,
		},
	},
}

type canary struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config returns the configuration rules for setting up the Canary
type Config struct {
	Enabled       bool
	Weight        int
	WeightTotal   int
	Header        string
	HeaderValue   string
	HeaderPattern string
	Cookie        string
}

// NewParser parses the ingress for canary related annotations
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return canary{
		r:                r,
		annotationConfig: CanaryAnnotations,
	}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the canary should be enabled and with what config
func (c canary) Parse(ing *networking.Ingress) (interface{}, error) {
	config := &Config{}
	var err error

	config.Enabled, err = parser.GetBoolAnnotation(canaryAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to 'false'", canaryAnnotation)
		}
		config.Enabled = false
	}

	config.Weight, err = parser.GetIntAnnotation(canaryWeightAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to '0'", canaryWeightAnnotation)
		}
		config.Weight = 0
	}

	config.WeightTotal, err = parser.GetIntAnnotation(canaryWeightTotalAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to '100'", canaryWeightTotalAnnotation)
		}
		config.WeightTotal = 100
	}

	config.Header, err = parser.GetStringAnnotation(canaryByHeaderAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to ''", canaryByHeaderAnnotation)
		}
		config.Header = ""
	}

	config.HeaderValue, err = parser.GetStringAnnotation(canaryByHeaderValueAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to ''", canaryByHeaderValueAnnotation)
		}
		config.HeaderValue = ""
	}

	config.HeaderPattern, err = parser.GetStringAnnotation(canaryByHeaderPatternAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to ''", canaryByHeaderPatternAnnotation)
		}
		config.HeaderPattern = ""
	}

	config.Cookie, err = parser.GetStringAnnotation(canaryByCookieAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to ''", canaryByCookieAnnotation)
		}
		config.Cookie = ""
	}

	if !config.Enabled && (config.Weight > 0 || config.Header != "" || config.HeaderValue != "" || config.Cookie != "" ||
		config.HeaderPattern != "") {
		return nil, errors.NewInvalidAnnotationConfiguration(canaryAnnotation, "configured but not enabled")
	}

	return config, nil
}

func (c canary) GetDocumentation() parser.AnnotationFields {
	return c.annotationConfig.Annotations
}

func (c canary) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(c.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, CanaryAnnotations.Annotations)
}
