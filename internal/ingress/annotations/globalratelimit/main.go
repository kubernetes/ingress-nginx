/*
Copyright 2020 The Kubernetes Authors.

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

package globalratelimit

import (
	"fmt"
	"strings"
	"time"

	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/pkg/util/sets"
)

const defaultKey = "$remote_addr"

const (
	globalRateLimitAnnotation             = "global-rate-limit"
	globalRateLimitWindowAnnotation       = "global-rate-limit-window"
	globalRateLimitKeyAnnotation          = "global-rate-limit-key"
	globalRateLimitIgnoredCidrsAnnotation = "global-rate-limit-ignored-cidrs"
)

var globalRateLimitAnnotationConfig = parser.Annotation{
	Group: "ratelimit",
	Annotations: parser.AnnotationFields{
		globalRateLimitAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation configures maximum allowed number of requests per window`,
		},
		globalRateLimitWindowAnnotation: {
			Validator:     parser.ValidateDuration,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `Configures a time window (i.e 1m) that the limit is applied`,
		},
		globalRateLimitKeyAnnotation: {
			Validator: parser.ValidateRegex(parser.NGINXVariable, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskHigh,
			Documentation: `This annotation Configures a key for counting the samples. Defaults to $remote_addr. 
			You can also combine multiple NGINX variables here, like ${remote_addr}-${http_x_api_client} which would mean the limit will be applied to 
			requests coming from the same API client (indicated by X-API-Client HTTP request header) with the same source IP address`,
		},
		globalRateLimitIgnoredCidrsAnnotation: {
			Validator: parser.ValidateCIDRs,
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation defines a comma separated list of IPs and CIDRs to match client IP against. 
			When there's a match request is not considered for rate limiting.`,
		},
	},
}

// Config encapsulates all global rate limit attributes
type Config struct {
	Namespace    string   `json:"namespace"`
	Limit        int      `json:"limit"`
	WindowSize   int      `json:"window-size"`
	Key          string   `json:"key"`
	IgnoredCIDRs []string `json:"ignored-cidrs"`
}

// Equal tests for equality between two Config types
func (l *Config) Equal(r *Config) bool {
	if l.Namespace != r.Namespace {
		return false
	}
	if l.Limit != r.Limit {
		return false
	}
	if l.WindowSize != r.WindowSize {
		return false
	}
	if l.Key != r.Key {
		return false
	}
	if len(l.IgnoredCIDRs) != len(r.IgnoredCIDRs) || !sets.StringElementsMatch(l.IgnoredCIDRs, r.IgnoredCIDRs) {
		return false
	}

	return true
}

type globalratelimit struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new globalratelimit annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return globalratelimit{
		r:                r,
		annotationConfig: globalRateLimitAnnotationConfig,
	}
}

// Parse extracts globalratelimit annotations from the given ingress
// and returns them structured as Config type
func (a globalratelimit) Parse(ing *networking.Ingress) (interface{}, error) {
	config := &Config{}

	limit, err := parser.GetIntAnnotation(globalRateLimitAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsInvalidContent(err) {
		return nil, err
	}
	rawWindowSize, err := parser.GetStringAnnotation(globalRateLimitWindowAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsValidationError(err) {
		return config, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("failed to parse 'global-rate-limit-window' value: %w", err),
		}
	}

	if limit == 0 || rawWindowSize == "" {
		return config, nil
	}

	windowSize, err := time.ParseDuration(rawWindowSize)
	if err != nil {
		return config, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("failed to parse 'global-rate-limit-window' value: %w", err),
		}
	}

	key, err := parser.GetStringAnnotation(globalRateLimitKeyAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.Warningf("invalid %s, defaulting to %s", globalRateLimitKeyAnnotation, defaultKey)
	}
	if key == "" {
		key = defaultKey
	}

	rawIgnoredCIDRs, err := parser.GetStringAnnotation(globalRateLimitIgnoredCidrsAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsInvalidContent(err) {
		return nil, err
	}
	ignoredCIDRs, err := net.ParseCIDRs(rawIgnoredCIDRs)
	if err != nil {
		return nil, err
	}

	config.Namespace = strings.ReplaceAll(string(ing.UID), "-", "")
	config.Limit = limit
	config.WindowSize = int(windowSize.Seconds())
	config.Key = key
	config.IgnoredCIDRs = ignoredCIDRs

	return config, nil
}

func (a globalratelimit) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a globalratelimit) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, globalRateLimitAnnotationConfig.Annotations)
}
