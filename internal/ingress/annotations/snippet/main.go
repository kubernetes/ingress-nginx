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

package snippet

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	configurationSnippetAnnotation = "configuration-snippet"
)

var configurationSnippetAnnotations = parser.Annotation{
	Group: "snippets",
	Annotations: parser.AnnotationFields{
		configurationSnippetAnnotation: {
			Validator:     parser.ValidateNull,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskCritical, // Critical, this annotation is not validated at all and allows arbitrary configutations
			Documentation: `This annotation allows setting a custom NGINX configuration on a location block. This annotation does not contain any validation and it's usage is not recommended!`,
		},
	},
}

type snippet struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new CORS annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return snippet{
		r:                r,
		annotationConfig: configurationSnippetAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
// used to indicate if the location/s contains a fragment of
// configuration to be included inside the paths of the rules
func (a snippet) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation(configurationSnippetAnnotation, ing, a.annotationConfig.Annotations)
}

func (a snippet) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a snippet) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, configurationSnippetAnnotations.Annotations)
}
