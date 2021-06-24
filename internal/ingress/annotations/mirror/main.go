/*
Copyright 2019 The Kubernetes Authors.

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

package mirror

import (
	"fmt"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns the mirror to use in a given location
type Config struct {
	Source      string `json:"source"`
	RequestBody string `json:"requestBody"`
	Target      string `json:"target"`
}

// Equal tests for equality between two Configuration types
func (m1 *Config) Equal(m2 *Config) bool {
	if m1 == m2 {
		return true
	}

	if m1 == nil || m2 == nil {
		return false
	}

	if m1.Source != m2.Source {
		return false
	}

	if m1.RequestBody != m2.RequestBody {
		return false
	}

	if m1.Target != m2.Target {
		return false
	}

	return true
}

type mirror struct {
	r resolver.Resolver
}

// NewParser creates a new mirror configuration annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return mirror{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure mirror
func (a mirror) Parse(ing *networking.Ingress) (interface{}, error) {
	config := &Config{
		Source: fmt.Sprintf("/_mirror-%v", ing.UID),
	}

	var err error
	config.RequestBody, err = parser.GetStringAnnotation("mirror-request-body", ing)
	if err != nil || config.RequestBody != "off" {
		config.RequestBody = "on"
	}

	config.Target, err = parser.GetStringAnnotation("mirror-target", ing)
	if err != nil {
		config.Target = ""
		config.Source = ""
	}

	return config, nil
}
