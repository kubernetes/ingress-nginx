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

package proxypass

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config describes the per location proxyPass config
type Config struct {
	// Address of the upstream proxy.
	Address string `json:"address"`
	// Port of the upstream proxy.
	Port string `json:"port"`
	// Replaces the address of the upstream proxy with the local node
	// address, as set by a "NODE_NAME" environment variable.
	ProxyToLocalNode bool `json:"proxyToLocalNode"`
}

// Equal tests for equality between two proxyPass config types
func (r1 *Config) Equal(r2 *Config) bool {
	if r1 == r2 {
		return true
	}
	if r1 == nil || r2 == nil {
		return false
	}
	if r1.Address != r2.Address {
		return false
	}
	if r1.Port != r2.Port {
		return false
	}
	if r1.ProxyToLocalNode != r2.ProxyToLocalNode {
		return false
	}

	return true
}

type proxypass struct {
	r resolver.Resolver
}

// NewParser creates a new proxyPass annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return proxypass{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to define the proxyPass value
func (a proxypass) Parse(ing *extensions.Ingress) (interface{}, error) {
	ppa, _ := parser.GetStringAnnotation("proxy-pass-address", ing)
	ppp, _ := parser.GetStringAnnotation("proxy-pass-port", ing)
	ptln, _ := parser.GetBoolAnnotation("proxy-to-local-node", ing)

	return &Config{
		Address:          ppa,
		Port:             ppp,
		ProxyToLocalNode: ptln,
	}, nil
}
