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

package httpport

import (
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config contains the HTTP and HTTPS ports used
// by the host to listen for incoming connection request
type Config struct {
	HTTPPort  int `json:"http-port"`
	HTTPSPort int `json:"https-port"`
}

// Equal tests for equality between two Config types
func (port1 *Config) Equal(port2 *Config) bool {
	if port1 == port2 {
		return true
	}
	if port1 == nil || port2 == nil {
		return false
	}
	if port1.HTTPPort != port2.HTTPPort {
		return false
	}
	if port1.HTTPSPort != port2.HTTPSPort {
		return false
	}
	return true
}

type httpPorts struct {
	r resolver.Resolver
}

// NewParser creates a new HTTP(S) port annotation parser
func NewParser(resolver resolver.Resolver) parser.IngressAnnotation {
	return httpPorts{resolver}
}

// Parse parses the annotations contained in the ingress
// rule used to configure HTTP(S) ports for the host
func (p httpPorts) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	config.HTTPPort, err = parser.GetIntAnnotation("http-port", ing)
	if err != nil {
		config.HTTPPort = 0
	}

	config.HTTPSPort, err = parser.GetIntAnnotation("https-port", ing)
	if err != nil {
		config.HTTPSPort = 0
	}

	return config, nil
}
