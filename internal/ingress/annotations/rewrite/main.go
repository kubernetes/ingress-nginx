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

package rewrite

import (
	"net/url"

	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	rewriteTargetAnnotation         = "rewrite-target"
	sslRedirectAnnotation           = "ssl-redirect"
	preserveTrailingSlashAnnotation = "preserve-trailing-slash"
	forceSSLRedirectAnnotation      = "force-ssl-redirect"
	useRegexAnnotation              = "use-regex"
	appRootAnnotation               = "app-root"
)

var rewriteAnnotations = parser.Annotation{
	Group: "rewrite",
	Annotations: parser.AnnotationFields{
		rewriteTargetAnnotation: {
			Validator: parser.ValidateRegex(parser.RegexPathWithCapture, false),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation allows to specify the target URI where the traffic must be redirected. It can contain regular characters and captured 
			groups specified as '$1', '$2', etc.`,
		},
		sslRedirectAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines if the location section is only accessible via SSL`,
		},
		preserveTrailingSlashAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation defines if the trailing slash should be preserved in the URI with 'ssl-redirect'`,
		},
		forceSSLRedirectAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation forces the redirection to HTTPS even if the Ingress is not TLS Enabled`,
		},
		useRegexAnnotation: {
			Validator: parser.ValidateBool,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation defines if the paths defined on an Ingress use regular expressions. To use regex on path
			the pathType should also be defined as 'ImplementationSpecific'.`,
		},
		appRootAnnotation: {
			Validator:     parser.ValidateRegex(parser.RegexPathWithCapture, false),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the Application Root that the Controller must redirect if it's in / context`,
		},
	},
}

// Config describes the per location redirect config
type Config struct {
	// Target URI where the traffic must be redirected
	Target string `json:"target"`
	// SSLRedirect indicates if the location section is accessible SSL only
	SSLRedirect bool `json:"sslRedirect"`
	// ForceSSLRedirect indicates if the location section is accessible SSL only
	ForceSSLRedirect bool `json:"forceSSLRedirect"`
	// PreserveTrailingSlash indicates if the trailing slash should be kept during a tls redirect
	PreserveTrailingSlash bool `json:"preserveTrailingSlash"`
	// AppRoot defines the Application Root that the Controller must redirect if it's in '/' context
	AppRoot string `json:"appRoot"`
	// UseRegex indicates whether or not the locations use regex paths
	UseRegex bool `json:"useRegex"`
}

// Equal tests for equality between two Redirect types
func (r1 *Config) Equal(r2 *Config) bool {
	if r1 == r2 {
		return true
	}
	if r1 == nil || r2 == nil {
		return false
	}
	if r1.Target != r2.Target {
		return false
	}
	if r1.SSLRedirect != r2.SSLRedirect {
		return false
	}
	if r1.ForceSSLRedirect != r2.ForceSSLRedirect {
		return false
	}
	if r1.AppRoot != r2.AppRoot {
		return false
	}
	if r1.UseRegex != r2.UseRegex {
		return false
	}

	return true
}

type rewrite struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new rewrite annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return rewrite{
		r:                r,
		annotationConfig: rewriteAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func (a rewrite) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	config.Target, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%sis invalid, defaulting to empty", rewriteTargetAnnotation)
		}
		config.Target = ""
	}
	config.SSLRedirect, err = parser.GetBoolAnnotation(sslRedirectAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%sis invalid, defaulting to '%s'", sslRedirectAnnotation, a.r.GetDefaultBackend().SSLRedirect)
		}
		config.SSLRedirect = a.r.GetDefaultBackend().SSLRedirect
	}
	config.PreserveTrailingSlash, err = parser.GetBoolAnnotation(preserveTrailingSlashAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%sis invalid, defaulting to '%s'", preserveTrailingSlashAnnotation, a.r.GetDefaultBackend().PreserveTrailingSlash)
		}
		config.PreserveTrailingSlash = a.r.GetDefaultBackend().PreserveTrailingSlash
	}

	config.ForceSSLRedirect, err = parser.GetBoolAnnotation(forceSSLRedirectAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%sis invalid, defaulting to '%s'", forceSSLRedirectAnnotation, a.r.GetDefaultBackend().ForceSSLRedirect)
		}
		config.ForceSSLRedirect = a.r.GetDefaultBackend().ForceSSLRedirect
	}

	config.UseRegex, err = parser.GetBoolAnnotation(useRegexAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%sis invalid, defaulting to 'false'", useRegexAnnotation)
		}
		config.UseRegex = false
	}

	config.AppRoot, err = parser.GetStringAnnotation(appRootAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if !errors.IsMissingAnnotations(err) && !errors.IsInvalidContent(err) {
			klog.Warningf("Annotation app-root contains an invalid value: %v", err)
		}

		return config, nil
	}

	u, err := url.ParseRequestURI(config.AppRoot)
	if err != nil {
		klog.Warningf("Annotation app-root contains an invalid value: %v", err)
		config.AppRoot = ""
		return config, nil
	}

	if u.IsAbs() {
		klog.Warningf("Annotation app-root only allows absolute paths (%v)", config.AppRoot)
		config.AppRoot = ""
		return config, nil
	}

	return config, nil
}

func (a rewrite) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a rewrite) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, rewriteAnnotations.Annotations)
}
