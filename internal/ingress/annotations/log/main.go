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
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type cors struct {
	r resolver.Resolver
}

// Config contains the configuration to be used in the Ingress
type Config struct {
	Access bool `json:"accessLog"`
}

// Equal tests for equality between two Config types
func (bd1 *Config) Equal(bd2 *Config) bool {
	if bd1.Access == bd2.Access {
		return true
	}

	return false
}

// NewParser creates a new access log annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return cors{r}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should enable logs
func (c cors) Parse(ing *extensions.Ingress) (interface{}, error) {
	accessEnabled, err := parser.GetBoolAnnotation("enable-access-log", ing)
	if err != nil {
		accessEnabled = true
	}

	return &Config{accessEnabled}, nil
}
