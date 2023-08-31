/*
Copyright 2016 The Kubernetes Authors.

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

package cors

import (
	"regexp"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	// Default values
	defaultCorsMethods = "GET, PUT, POST, DELETE, PATCH, OPTIONS"
	defaultCorsHeaders = "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"
	defaultCorsMaxAge  = 1728000
)

var (
	// Regex are defined here to prevent information leak, if user tries to set anything not valid
	// that could cause the Response to contain some internal value/variable (like returning $pid, $upstream_addr, etc)
	// Origin must contain a http/s Origin (including or not the port) or the value '*'
	// This Regex is composed of the following:
	// * Sets a group that can be (https?://)?*?.something.com:port?
	// * Allows this to be repeated as much as possible, and separated by comma
	// Otherwise it should be '*'
	corsOriginRegexValidator = regexp.MustCompile(`^((((https?://)?(\*\.)?[A-Za-z0-9\-.]*(:\d+)?,?)+)|\*)?$`)
	// corsOriginRegex defines the regex for validation inside Parse
	corsOriginRegex = regexp.MustCompile(`^(https?://(\*\.)?[A-Za-z0-9\-.]*(:\d+)?|\*)?$`)
	// Method must contain valid methods list (PUT, GET, POST, BLA)
	// May contain or not spaces between each verb
	corsMethodsRegex = regexp.MustCompile(`^([A-Za-z]+,?\s?)+$`)
	// Expose Headers must contain valid values only (*, X-HEADER12, X-ABC)
	// May contain or not spaces between each Header
	corsExposeHeadersRegex = regexp.MustCompile(`^(([A-Za-z0-9\-\_]+|\*),?\s?)+$`)
)

const (
	corsEnableAnnotation           = "enable-cors"
	corsAllowOriginAnnotation      = "cors-allow-origin"
	corsAllowHeadersAnnotation     = "cors-allow-headers"
	corsAllowMethodsAnnotation     = "cors-allow-methods"
	corsAllowCredentialsAnnotation = "cors-allow-credentials" //#nosec G101
	corsExposeHeadersAnnotation    = "cors-expose-headers"
	corsMaxAgeAnnotation           = "cors-max-age"
)

var corsAnnotation = parser.Annotation{
	Group: "cors",
	Annotations: parser.AnnotationFields{
		corsEnableAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables Cross-Origin Resource Sharing (CORS) in an Ingress rule`,
		},
		corsAllowOriginAnnotation: {
			Validator: parser.ValidateRegex(corsOriginRegexValidator, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation controls what's the accepted Origin for CORS.
			This is a multi-valued field, separated by ','. It must follow this format: http(s)://origin-site.com or http(s)://origin-site.com:port
			It also supports single level wildcard subdomains and follows this format: http(s)://*.foo.bar, http(s)://*.bar.foo:8080 or http(s)://*.abc.bar.foo:9000`,
		},
		corsAllowHeadersAnnotation: {
			Validator: parser.ValidateRegex(parser.HeadersVariable, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation controls which headers are accepted.
			This is a multi-valued field, separated by ',' and accepts letters, numbers, _ and -`,
		},
		corsAllowMethodsAnnotation: {
			Validator: parser.ValidateRegex(corsMethodsRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation controls which methods are accepted.
			This is a multi-valued field, separated by ',' and accepts only letters (upper and lower case)`,
		},
		corsAllowCredentialsAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation controls if credentials can be passed during CORS operations.`,
		},
		corsExposeHeadersAnnotation: {
			Validator: parser.ValidateRegex(corsExposeHeadersRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation controls which headers are exposed to response.
			This is a multi-valued field, separated by ',' and accepts letters, numbers, _, - and *.`,
		},
		corsMaxAgeAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation controls how long, in seconds, preflight requests can be cached.`,
		},
	},
}

type cors struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config contains the Cors configuration to be used in the Ingress
type Config struct {
	CorsEnabled          bool     `json:"corsEnabled"`
	CorsAllowOrigin      []string `json:"corsAllowOrigin"`
	CorsAllowMethods     string   `json:"corsAllowMethods"`
	CorsAllowHeaders     string   `json:"corsAllowHeaders"`
	CorsAllowCredentials bool     `json:"corsAllowCredentials"`
	CorsExposeHeaders    string   `json:"corsExposeHeaders"`
	CorsMaxAge           int      `json:"corsMaxAge"`
}

// NewParser creates a new CORS annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return cors{
		r:                r,
		annotationConfig: corsAnnotation,
	}
}

// Equal tests for equality between two External types
func (c1 *Config) Equal(c2 *Config) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	if c1.CorsMaxAge != c2.CorsMaxAge {
		return false
	}
	if c1.CorsExposeHeaders != c2.CorsExposeHeaders {
		return false
	}
	if c1.CorsAllowCredentials != c2.CorsAllowCredentials {
		return false
	}
	if c1.CorsAllowHeaders != c2.CorsAllowHeaders {
		return false
	}
	if c1.CorsAllowMethods != c2.CorsAllowMethods {
		return false
	}
	if c1.CorsEnabled != c2.CorsEnabled {
		return false
	}

	if len(c1.CorsAllowOrigin) != len(c2.CorsAllowOrigin) {
		return false
	}

	for i, v := range c1.CorsAllowOrigin {
		if v != c2.CorsAllowOrigin[i] {
			return false
		}
	}

	return true
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should allows CORS
func (c cors) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	config.CorsEnabled, err = parser.GetBoolAnnotation(corsEnableAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("enable-cors is invalid, defaulting to 'false'")
		}
		config.CorsEnabled = false
	}

	config.CorsAllowOrigin = []string{}
	unparsedOrigins, err := parser.GetStringAnnotation(corsAllowOriginAnnotation, ing, c.annotationConfig.Annotations)
	if err == nil {
		origins := strings.Split(unparsedOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin == "*" {
				config.CorsAllowOrigin = []string{"*"}
				break
			}

			if !corsOriginRegex.MatchString(origin) {
				klog.Errorf("Error parsing cors-allow-origin parameters. Supplied incorrect origin: %s. Skipping.", origin)
				continue
			}
			config.CorsAllowOrigin = append(config.CorsAllowOrigin, origin)
			klog.Infof("Current config.corsAllowOrigin %v", config.CorsAllowOrigin)
		}
	} else {
		if errors.IsValidationError(err) {
			klog.Warningf("cors-allow-origin is invalid, defaulting to '*'")
		}
		config.CorsAllowOrigin = []string{"*"}
	}

	config.CorsAllowHeaders, err = parser.GetStringAnnotation(corsAllowHeadersAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil || !parser.HeadersVariable.MatchString(config.CorsAllowHeaders) {
		config.CorsAllowHeaders = defaultCorsHeaders
	}

	config.CorsAllowMethods, err = parser.GetStringAnnotation(corsAllowMethodsAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil || !corsMethodsRegex.MatchString(config.CorsAllowMethods) {
		config.CorsAllowMethods = defaultCorsMethods
	}

	config.CorsAllowCredentials, err = parser.GetBoolAnnotation(corsAllowCredentialsAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			if errors.IsValidationError(err) {
				klog.Warningf("cors-allow-credentials is invalid, defaulting to 'true'")
			}
		}
		config.CorsAllowCredentials = true
	}

	config.CorsExposeHeaders, err = parser.GetStringAnnotation(corsExposeHeadersAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil || !corsExposeHeadersRegex.MatchString(config.CorsExposeHeaders) {
		config.CorsExposeHeaders = ""
	}

	config.CorsMaxAge, err = parser.GetIntAnnotation(corsMaxAgeAnnotation, ing, c.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("cors-max-age is invalid, defaulting to %d", defaultCorsMaxAge)
		}
		config.CorsMaxAge = defaultCorsMaxAge
	}

	return config, nil
}

func (c cors) GetDocumentation() parser.AnnotationFields {
	return c.annotationConfig.Annotations
}

func (c cors) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(c.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, corsAnnotation.Annotations)
}
