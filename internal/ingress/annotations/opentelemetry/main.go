/*
Copyright 2022 The Kubernetes Authors.

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

package opentelemetry

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type opentelemetry struct {
	r resolver.Resolver
}

// Config contains the configuration to be used in the Ingress
type Config struct {
	Enabled       bool   `json:"enabled"`
	Set           bool   `json:"set"`
	TrustEnabled  bool   `json:"trust-enabled"`
	TrustSet      bool   `json:"trust-set"`
	OperationName string `json:"operation-name"`
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

	if bd1.OperationName != bd2.OperationName {
		return false
	}

	return true
}

// NewParser creates a new serviceUpstream annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return opentelemetry{r}
}

// Parse parses the annotations to look for opentelemetry configurations
func (c opentelemetry) Parse(ing *networking.Ingress) (interface{}, error) {
	cfg := Config{}
	enabled, err := parser.GetBoolAnnotation("enable-opentelemetry", ing)
	if err != nil {
		return &cfg, nil
	}
	cfg.Set = true
	cfg.Enabled = enabled
	if !enabled {
		return &cfg, nil
	}

	trustEnabled, err := parser.GetBoolAnnotation("opentelemetry-trust-incoming-span", ing)
	if err != nil {
		operationName, err := parser.GetStringAnnotation("opentelemetry-operation-name", ing)
		if err != nil {
			return &cfg, nil
		}
		cfg.OperationName = operationName
		return &cfg, nil
	}

	cfg.TrustSet = true
	cfg.TrustEnabled = trustEnabled
	operationName, err := parser.GetStringAnnotation("opentelemetry-operation-name", ing)
	if err != nil {
		return &cfg, nil
	}
	cfg.OperationName = operationName
	return &cfg, nil
}
