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
	"fmt"
	"sort"
	"strings"

	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/resolver"
	"k8s.io/ingress/core/pkg/net"
)

const (
	limitIP        = "ingress.kubernetes.io/limit-connections"
	limitRPS       = "ingress.kubernetes.io/limit-rps"
	limitRPM       = "ingress.kubernetes.io/limit-rpm"
	limitRATE      = "ingress.kubernetes.io/limit-rate"
	limitRATEAFTER = "ingress.kubernetes.io/limit-rate-after"
	limitWhitelist = "ingress.kubernetes.io/limit-whitelist"

	// allow 5 times the specified limit as burst
	defBurst = 5

	// 1MB -> 16 thousand 64-byte states or about 8 thousand 128-byte states
	// default is 5MB
	defSharedSize = 5
)

// RateLimit returns rate limit configuration for an Ingress rule limiting the
// number of connections per IP address and/or connections per second.
// If you both annotations are specified in a single Ingress rule, RPS limits
// takes precedence
type RateLimit struct {
	// Connections indicates a limit with the number of connections per IP address
	Connections Zone `json:"connections"`
	// RPS indicates a limit with the number of connections per second
	RPS Zone `json:"rps"`

	RPM Zone `json:"rpm"`

	LimitRate int `json:"limit-rate"`

	LimitRateAfter int `json:"limit-rate-after"`

	Name string `json:"name"`

	Whitelist []string `json:"whitelist"`
}

// Equal tests for equality between two RateLimit types
func (rt1 *RateLimit) Equal(rt2 *RateLimit) bool {
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
	if rt1.Name != rt2.Name {
		return false
	}

	for _, r1l := range rt1.Whitelist {
		found := false
		for _, rl2 := range rt2.Whitelist {
			if r1l == rl2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
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

type ratelimit struct {
	backendResolver resolver.DefaultBackend
}

// NewParser creates a new ratelimit annotation parser
func NewParser(br resolver.DefaultBackend) parser.IngressAnnotation {
	return ratelimit{br}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func (a ratelimit) Parse(ing *extensions.Ingress) (interface{}, error) {
	defBackend := a.backendResolver.GetDefaultBackend()
	lr, err := parser.GetIntAnnotation(limitRATE, ing)
	if err != nil {
		lr = defBackend.LimitRate
	}
	lra, err := parser.GetIntAnnotation(limitRATEAFTER, ing)
	if err != nil {
		lra = defBackend.LimitRateAfter
	}

	rpm, _ := parser.GetIntAnnotation(limitRPM, ing)
	rps, _ := parser.GetIntAnnotation(limitRPS, ing)
	conn, _ := parser.GetIntAnnotation(limitIP, ing)

	val, _ := parser.GetStringAnnotation(limitWhitelist, ing)

	cidrs, err := parseCIDRs(val)
	if err != nil {
		return nil, err
	}

	if rpm == 0 && rps == 0 && conn == 0 {
		return &RateLimit{
			Connections:    Zone{},
			RPS:            Zone{},
			RPM:            Zone{},
			LimitRate:      lr,
			LimitRateAfter: lra,
		}, nil
	}

	zoneName := fmt.Sprintf("%v_%v", ing.GetNamespace(), ing.GetName())

	return &RateLimit{
		Connections: Zone{
			Name:       fmt.Sprintf("%v_conn", zoneName),
			Limit:      conn,
			Burst:      conn * defBurst,
			SharedSize: defSharedSize,
		},
		RPS: Zone{
			Name:       fmt.Sprintf("%v_rps", zoneName),
			Limit:      rps,
			Burst:      rps * defBurst,
			SharedSize: defSharedSize,
		},
		RPM: Zone{
			Name:       fmt.Sprintf("%v_rpm", zoneName),
			Limit:      rpm,
			Burst:      rpm * defBurst,
			SharedSize: defSharedSize,
		},
		LimitRate:      lr,
		LimitRateAfter: lra,
		Name:           zoneName,
		Whitelist:      cidrs,
	}, nil
}

func parseCIDRs(s string) ([]string, error) {
	if s == "" {
		return []string{}, nil
	}

	values := strings.Split(s, ",")

	ipnets, ips, err := net.ParseIPNets(values...)
	if err != nil {
		return nil, err
	}

	cidrs := []string{}
	for k := range ipnets {
		cidrs = append(cidrs, k)
	}

	for k := range ips {
		cidrs = append(cidrs, k)
	}

	sort.Strings(cidrs)

	return cidrs, nil
}
