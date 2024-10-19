/*
Copyright 2017 The Kubernetes Authors.

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

package client

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	clientBodyBufferSizeAnnotation = "client-body-buffer-size"
)

var clientAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		clientBodyBufferSizeAnnotation: {
			Validator: parser.ValidateRegex(parser.SizeRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow, // Low, as it allows just a set of options
			Documentation: `Sets buffer size for reading client request body per location. 
			In case the request body is larger than the buffer, the whole body or only its part is written to a temporary file. 
			By default, buffer size is equal to two memory pages. This is 8K on x86, other 32-bit platforms, and x86-64. 
			It is usually 16K on other 64-bit platforms. This annotation is applied to each location provided in the ingress rule.`,
		},
	},
}

type Config struct {
	BodyBufferSize string `json:"bodyBufferSize"`
}

// Equal tests for equality between two Configuration types
func (l1 *Config) Equal(l2 *Config) bool {
	if l1 == l2 {
		return true
	}
	if l1 == nil || l2 == nil {
		return false
	}
	if l1.BodyBufferSize != l2.BodyBufferSize {
		return false
	}

	return true
}

type client struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new client annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return client{
		r:                r,
		annotationConfig: clientAnnotations,
	}
}

func (c client) GetDocumentation() parser.AnnotationFields {
	return c.annotationConfig.Annotations
}

// Parse parses the annotations contained in the ingress rule
// used to add an client related configuration to the provided locations.
func (c client) Parse(ing *networking.Ingress) (interface{}, error) {
	config := &Config{}

	var err error
	config.BodyBufferSize, err = parser.GetStringAnnotation(clientBodyBufferSizeAnnotation, ing, c.annotationConfig.Annotations)

	return config, err
}

func (c client) Validate(annotations map[string]string) error {
	maxRisk := parser.StringRiskToRisk(c.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(annotations, maxRisk, clientAnnotations.Annotations)
}
