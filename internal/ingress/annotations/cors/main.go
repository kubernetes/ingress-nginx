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
	corsOriginRegex = regexp.MustCompile(`^(https?://(\*\.)?[A-Za-z0-9\-\.]*(:[0-9]+)?|\*)?$`)
	// Method must contain valid methods list (PUT, GET, POST, BLA)
	// May contain or not spaces between each verb
	corsMethodsRegex = regexp.MustCompile(`^([A-Za-z]+,?\s?)+$`)
	// Headers must contain valid values only (X-HEADER12, X-ABC)
	// May contain or not spaces between each Header
	corsHeadersRegex = regexp.MustCompile(`^([A-Za-z0-9\-\_]+,?\s?)+$`)
	// Expose Headers must contain valid values only (*, X-HEADER12, X-ABC)
	// May contain or not spaces between each Header
	corsExposeHeadersRegex = regexp.MustCompile(`^(([A-Za-z0-9\-\_]+|\*),?\s?)+$`)
)

type cors struct {
	r resolver.Resolver
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
	return cors{r}
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

	config.CorsEnabled, err = parser.GetBoolAnnotation("enable-cors", ing)
	if err != nil {
		config.CorsEnabled = false
	}

	config.CorsAllowOrigin = []string{}
	unparsedOrigins, err := parser.GetStringAnnotation("cors-allow-origin", ing)
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
		config.CorsAllowOrigin = []string{"*"}
	}

	config.CorsAllowHeaders, err = parser.GetStringAnnotation("cors-allow-headers", ing)
	if err != nil || !corsHeadersRegex.MatchString(config.CorsAllowHeaders) {
		config.CorsAllowHeaders = defaultCorsHeaders
	}

	config.CorsAllowMethods, err = parser.GetStringAnnotation("cors-allow-methods", ing)
	if err != nil || !corsMethodsRegex.MatchString(config.CorsAllowMethods) {
		config.CorsAllowMethods = defaultCorsMethods
	}

	config.CorsAllowCredentials, err = parser.GetBoolAnnotation("cors-allow-credentials", ing)
	if err != nil {
		config.CorsAllowCredentials = true
	}

	config.CorsExposeHeaders, err = parser.GetStringAnnotation("cors-expose-headers", ing)
	if err != nil || !corsExposeHeadersRegex.MatchString(config.CorsExposeHeaders) {
		config.CorsExposeHeaders = ""
	}

	config.CorsMaxAge, err = parser.GetIntAnnotation("cors-max-age", ing)
	if err != nil {
		config.CorsMaxAge = defaultCorsMaxAge
	}

	return config, nil
}
