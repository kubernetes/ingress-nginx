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

type log struct {
	r resolver.Resolver
}

// Config contains the configuration to be used in the Ingress
type Config struct {
	Access  bool `json:"accessLog"`
	Rewrite bool `json:"rewriteLog"`
}

// Equal tests for equality between two Config types
func (bd1 *Config) Equal(bd2 *Config) bool {
	if bd1.Access != bd2.Access {
		return false
	}

	if bd1.Rewrite != bd2.Rewrite {
		return false
	}

	return true
}

// NewParser creates a new log annotations parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return log{r}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should enable logs
func (l log) Parse(ing *extensions.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	config.Access, err = parser.GetBoolAnnotation("enable-access-log", ing)
	if err != nil {
		config.Access = true
	}

	config.Rewrite, err = parser.GetBoolAnnotation("enable-rewrite-log", ing)
	if err != nil {
		config.Rewrite = false
	}

	return config, nil
}
