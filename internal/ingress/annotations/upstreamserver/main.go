/*
Copyright 2025 The Kubernetes Authors.

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

package upstreamserver

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	upstreamServerMaxConnsAnnotation = "upstream-server-max-conns"
)

var upstreamServerAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		upstreamServerMaxConnsAnnotation: {
			Validator:     parser.ValidateUint,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation allows setting the maximum number of simultaneous active connections to the proxied server. Default value is 0, which means no limit.`,
		},
	},
}

// Config contains the upstream server configuration
type Config struct {
	MaxConns uint `json:"maxConns,omitempty"`
}

type upstreamServer struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new serviceUpstream annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return upstreamServer{
		r:                r,
		annotationConfig: upstreamServerAnnotations,
	}
}

func (s upstreamServer) Parse(ing *networking.Ingress) (interface{}, error) {
	defBackend := s.r.GetDefaultBackend()

	val, err := parser.GetUintAnnotation(upstreamServerMaxConnsAnnotation, ing, s.annotationConfig.Annotations)
	// A missing annotation is not a problem, just use the default
	if err == errors.ErrMissingAnnotations {
		return &Config{MaxConns: defBackend.UpstreamServerMaxConns}, nil
	} else if err != nil {
		return &Config{MaxConns: 0}, err
	}

	return &Config{MaxConns: val}, nil
}

func (s upstreamServer) GetDocumentation() parser.AnnotationFields {
	return s.annotationConfig.Annotations
}

func (s upstreamServer) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(s.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, upstreamServerAnnotations.Annotations)
}
