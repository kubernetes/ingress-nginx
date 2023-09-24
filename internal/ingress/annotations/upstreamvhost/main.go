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

package upstreamvhost

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	upstreamVhostAnnotation = "upstream-vhost"
)

var upstreamVhostAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		upstreamVhostAnnotation: {
			Validator: parser.ValidateServerName,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow, // Low, as it allows regexes but on a very limited set
			Documentation: `This configuration setting allows you to control the value for host in the following statement: proxy_set_header Host $host, which forms part of the location block. 
			This is useful if you need to call the upstream server by something other than $host`,
		},
	},
}

type upstreamVhost struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new upstream VHost annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return upstreamVhost{
		r:                r,
		annotationConfig: upstreamVhostAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
// used to indicate if the location/s contains a fragment of
// configuration to be included inside the paths of the rules
func (a upstreamVhost) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation(upstreamVhostAnnotation, ing, a.annotationConfig.Annotations)
}

func (a upstreamVhost) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a upstreamVhost) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, upstreamVhostAnnotations.Annotations)
}
