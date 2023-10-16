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

package ratelimit

import (
	"encoding/base64"
	"fmt"
	"strings"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/pkg/util/sets"
)

const (
	// allow 5 times the specified limit as burst
	defBurst = 5

	// 1MB -> 16 thousand 64-byte states or about 8 thousand 128-byte states
	// default is 5MB
	defSharedSize = 5
)

// Config returns rate limit configuration for an Ingress rule limiting the
// number of connections per IP address and/or connections per second.
// If you both annotations are specified in a single Ingress rule, RPS limits
// takes precedence
type Config struct {
	// Connections indicates a limit with the number of connections per IP address
	Connections Zone `json:"connections"`
	// RPS indicates a limit with the number of connections per second
	RPS Zone `json:"rps"`

	RPM Zone `json:"rpm"`

	LimitRate int `json:"limit-rate"`

	LimitRateAfter int `json:"limit-rate-after"`

	Name string `json:"name"`

	ID string `json:"id"`

	Allowlist []string `json:"allowlist"`
}

// Equal tests for equality between two RateLimit types
func (rt1 *Config) Equal(rt2 *Config) bool {
	if rt1 == rt2 {
		return true
	}
	if rt1 == nil || rt2 == nil {
		return false
	}
	if !(&rt1.Connections).Equal(&rt2.Connections) {
		return false
	}
	if !(&rt1.RPM).Equal(&rt2.RPM) {
		return false
	}
	if !(&rt1.RPS).Equal(&rt2.RPS) {
		return false
	}
	if rt1.LimitRate != rt2.LimitRate {
		return false
	}
	if rt1.LimitRateAfter != rt2.LimitRateAfter {
		return false
	}
	if rt1.ID != rt2.ID {
		return false
	}
	if rt1.Name != rt2.Name {
		return false
	}
	if len(rt1.Allowlist) != len(rt2.Allowlist) {
		return false
	}

	return sets.StringElementsMatch(rt1.Allowlist, rt2.Allowlist)
}

// Zone returns information about the NGINX rate limit (limit_req_zone)
// http://nginx.org/en/docs/http/ngx_http_limit_req_module.html#limit_req_zone
type Zone struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Burst int    `json:"burst"`
	// SharedSize amount of shared memory for the zone
	SharedSize int `json:"sharedSize"`
}

// Equal tests for equality between two Zone types
func (z1 *Zone) Equal(z2 *Zone) bool {
	if z1 == z2 {
		return true
	}
	if z1 == nil || z2 == nil {
		return false
	}
	if z1.Name != z2.Name {
		return false
	}
	if z1.Limit != z2.Limit {
		return false
	}
	if z1.Burst != z2.Burst {
		return false
	}
	if z1.SharedSize != z2.SharedSize {
		return false
	}

	return true
}

const (
	limitRateAnnotation                = "limit-rate"
	limitRateAfterAnnotation           = "limit-rate-after"
	limitRateRPMAnnotation             = "limit-rpm"
	limitRateRPSAnnotation             = "limit-rps"
	limitRateConnectionsAnnotation     = "limit-connections"
	limitRateBurstMultiplierAnnotation = "limit-burst-multiplier"
	limitWhitelistAnnotation           = "limit-whitelist" // This annotation is an alias for limit-allowlist
	limitAllowlistAnnotation           = "limit-allowlist"
)

var rateLimitAnnotations = parser.Annotation{
	Group: "rate-limit",
	Annotations: parser.AnnotationFields{
		limitRateAnnotation: {
			Validator: parser.ValidateInt,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Limits the rate of response transmission to a client. The rate is specified in bytes per second. 
			The zero value disables rate limiting. The limit is set per a request, and so if a client simultaneously opens two connections, the overall rate will be twice as much as the specified limit.
			References: https://nginx.org/en/docs/http/ngx_http_core_module.html#limit_rate`,
		},
		limitRateAfterAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Sets the initial amount after which the further transmission of a response to a client will be rate limited.`,
		},
		limitRateRPMAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Requests per minute that will be allowed.`,
		},
		limitRateRPSAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Requests per second that will be allowed.`,
		},
		limitRateConnectionsAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Number of connections that will be allowed`,
		},
		limitRateBurstMultiplierAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Burst multiplier for a limit-rate enabled location.`,
		},
		limitAllowlistAnnotation: {
			Validator:         parser.ValidateCIDRs,
			Scope:             parser.AnnotationScopeLocation,
			Risk:              parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation:     `List of CIDR/IP addresses that will not be rate-limited.`,
			AnnotationAliases: []string{limitWhitelistAnnotation},
		},
	},
}

type ratelimit struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new ratelimit annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return ratelimit{
		r:                r,
		annotationConfig: rateLimitAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func (a ratelimit) Parse(ing *networking.Ingress) (interface{}, error) {
	defBackend := a.r.GetDefaultBackend()
	lr, err := parser.GetIntAnnotation(limitRateAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		lr = defBackend.LimitRate
	}
	lra, err := parser.GetIntAnnotation(limitRateAfterAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		lra = defBackend.LimitRateAfter
	}

	rpm, err := parser.GetIntAnnotation(limitRateRPMAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && errors.IsValidationError(err) {
		return nil, err
	}
	rps, err := parser.GetIntAnnotation(limitRateRPSAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && errors.IsValidationError(err) {
		return nil, err
	}
	conn, err := parser.GetIntAnnotation(limitRateConnectionsAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && errors.IsValidationError(err) {
		return nil, err
	}
	burstMultiplier, err := parser.GetIntAnnotation(limitRateBurstMultiplierAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		burstMultiplier = defBurst
	}

	val, err := parser.GetStringAnnotation(limitAllowlistAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && errors.IsValidationError(err) {
		return nil, err
	}

	cidrs, errCidr := net.ParseCIDRs(val)
	if errCidr != nil {
		return nil, errCidr
	}

	if rpm == 0 && rps == 0 && conn == 0 {
		return &Config{
			Connections:    Zone{},
			RPS:            Zone{},
			RPM:            Zone{},
			LimitRate:      lr,
			LimitRateAfter: lra,
		}, nil
	}

	zoneName := fmt.Sprintf("%v_%v_%v", ing.GetNamespace(), ing.GetName(), ing.UID)

	return &Config{
		Connections: Zone{
			Name:       fmt.Sprintf("%v_conn", zoneName),
			Limit:      conn,
			Burst:      conn * burstMultiplier,
			SharedSize: defSharedSize,
		},
		RPS: Zone{
			Name:       fmt.Sprintf("%v_rps", zoneName),
			Limit:      rps,
			Burst:      rps * burstMultiplier,
			SharedSize: defSharedSize,
		},
		RPM: Zone{
			Name:       fmt.Sprintf("%v_rpm", zoneName),
			Limit:      rpm,
			Burst:      rpm * burstMultiplier,
			SharedSize: defSharedSize,
		},
		LimitRate:      lr,
		LimitRateAfter: lra,
		Name:           zoneName,
		ID:             encode(zoneName),
		Allowlist:      cidrs,
	}, nil
}

func encode(s string) string {
	str := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.ReplaceAll(str, "=", "")
}

func (a ratelimit) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a ratelimit) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, rateLimitAnnotations.Annotations)
}
