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

package upstreamhashby

import (
	"regexp"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	upstreamHashByAnnotation       = "upstream-hash-by"
	upstreamHashBySubsetAnnotation = "upstream-hash-by-subset"
	upstreamHashBySubsetSize       = "upstream-hash-by-subset-size"
)

var (
	specialChars = regexp.QuoteMeta("_${}")
	hashByRegex  = regexp.MustCompilePOSIX(`^[A-Za-z0-9\-` + specialChars + `]*$`)
)

var upstreamHashByAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		upstreamHashByAnnotation: {
			Validator: parser.ValidateRegex(*hashByRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskHigh, // High, this annotation allows accessing NGINX variables
			Documentation: `This annotation defines the nginx variable, text value or any combination thereof to use for consistent hashing. 
			For example: nginx.ingress.kubernetes.io/upstream-hash-by: "$request_uri" or nginx.ingress.kubernetes.io/upstream-hash-by: "$request_uri$host" or nginx.ingress.kubernetes.io/upstream-hash-by: "${request_uri}-text-value" to consistently hash upstream requests by the current request URI.`,
		},
		upstreamHashBySubsetAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation maps requests to subset of nodes instead of a single one.`,
		},
		upstreamHashBySubsetSize: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation determines the size of each subset (default 3)`,
		},
	},
}

type upstreamhashby struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config contains the Consistent hash configuration to be used in the Ingress
type Config struct {
	UpstreamHashBy           string `json:"upstream-hash-by,omitempty"`
	UpstreamHashBySubset     bool   `json:"upstream-hash-by-subset,omitempty"`
	UpstreamHashBySubsetSize int    `json:"upstream-hash-by-subset-size,omitempty"`
}

// NewParser creates a new UpstreamHashBy annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return upstreamhashby{
		r:                r,
		annotationConfig: upstreamHashByAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
func (a upstreamhashby) Parse(ing *networking.Ingress) (interface{}, error) {
	upstreamHashBy, err := parser.GetStringAnnotation(upstreamHashByAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}
	upstreamHashBySubset, err := parser.GetBoolAnnotation(upstreamHashBySubsetAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}
	upstreamHashbySubsetSize, err := parser.GetIntAnnotation(upstreamHashBySubsetSize, ing, a.annotationConfig.Annotations)
	if err != nil && !errors.IsMissingAnnotations(err) {
		return nil, err
	}

	if upstreamHashbySubsetSize == 0 {
		upstreamHashbySubsetSize = 3
	}

	return &Config{upstreamHashBy, upstreamHashBySubset, upstreamHashbySubsetSize}, nil
}

func (a upstreamhashby) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}
