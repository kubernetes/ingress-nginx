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

package luarestywaf

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns lua-resty-waf configuration for an Ingress rule
type Config struct {
	Enabled bool `json:"enabled"`
	Debug   bool `json:"debug"`
}

// Equal tests for equality between two Config types
func (e1 *Config) Equal(e2 *Config) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.Enabled != e2.Enabled {
		return false
	}
	if e1.Debug != e2.Debug {
		return false
	}

	return true
}

type luarestywaf struct {
	r resolver.Resolver
}

// NewParser creates a new CORS annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return luarestywaf{r}
}

// Parse parses the annotations contained in the ingress rule
// used to indicate if the location/s contains a fragment of
// configuration to be included inside the paths of the rules
func (a luarestywaf) Parse(ing *extensions.Ingress) (interface{}, error) {
	enabled, err := parser.GetBoolAnnotation("lua-resty-waf", ing)
	if err != nil {
		return &Config{}, err
	}

	debug, _ := parser.GetBoolAnnotation("lua-resty-waf-debug", ing)

	return &Config{
		Enabled: enabled,
		Debug:   debug,
	}, nil
}
