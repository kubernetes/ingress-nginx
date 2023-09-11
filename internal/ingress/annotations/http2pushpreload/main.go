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

package http2pushpreload

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	http2PushPreloadAnnotation = "http2-push-preload"
)

var http2PushPreloadAnnotations = parser.Annotation{
	Group: "http2",
	Annotations: parser.AnnotationFields{
		http2PushPreloadAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `Enables automatic conversion of preload links specified in the “Link” response header fields into push requests`,
		},
	},
}

type http2PushPreload struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new http2PushPreload annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return http2PushPreload{
		r:                r,
		annotationConfig: http2PushPreloadAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
// used to add http2 push preload to the server
func (h2pp http2PushPreload) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetBoolAnnotation(http2PushPreloadAnnotation, ing, h2pp.annotationConfig.Annotations)
}

func (h2pp http2PushPreload) GetDocumentation() parser.AnnotationFields {
	return h2pp.annotationConfig.Annotations
}

func (h2pp http2PushPreload) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(h2pp.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, http2PushPreloadAnnotations.Annotations)
}
