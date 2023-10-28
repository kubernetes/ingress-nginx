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

package clientbodybuffersize

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	clientBodyBufferSizeAnnotation = "client-body-buffer-size"
)

var clientBodyBufferSizeConfig = parser.Annotation{
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

type clientBodyBufferSize struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new clientBodyBufferSize annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return clientBodyBufferSize{
		r:                r,
		annotationConfig: clientBodyBufferSizeConfig,
	}
}

func (cbbs clientBodyBufferSize) GetDocumentation() parser.AnnotationFields {
	return cbbs.annotationConfig.Annotations
}

// Parse parses the annotations contained in the ingress rule
// used to add an client-body-buffer-size to the provided locations
func (cbbs clientBodyBufferSize) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation(clientBodyBufferSizeAnnotation, ing, cbbs.annotationConfig.Annotations)
}

func (cbbs clientBodyBufferSize) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(cbbs.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, clientBodyBufferSizeConfig.Annotations)
}
