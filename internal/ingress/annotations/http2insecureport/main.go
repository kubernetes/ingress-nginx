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

package http2insecureport

import (
	networking "k8s.io/api/networking/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type http2InsecurePort struct {
}

// NewParser creates a new default backend annotation parser
func NewParser(_ resolver.Resolver) parser.IngressAnnotation {
	return http2InsecurePort{}
}

// Parse parses the annotations contained in the ingress to use
// a custom default backend
func (h http2InsecurePort) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetBoolAnnotation("http2-insecure-port", ing)
}
