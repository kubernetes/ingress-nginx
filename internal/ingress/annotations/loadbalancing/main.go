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

package loadbalancing

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// LB Alghorithms are defined in https://github.com/kubernetes/ingress-nginx/blob/d3e75b056f77be54e01bdb18675f1bb46caece31/rootfs/etc/nginx/lua/balancer.lua#L28

const (
	loadBalanceAlghoritmAnnotation = "load-balance"
)

var loadBalanceAlghoritms = []string{"round_robin", "chash", "chashsubset", "sticky_balanced", "sticky_persistent", "ewma"}

var loadBalanceAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		loadBalanceAlghoritmAnnotation: {
			Validator: parser.ValidateOptions(loadBalanceAlghoritms, true, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation allows setting the load balancing alghorithm that should be used. If none is specified, defaults to 
			the default configured by Ingress admin, otherwise to round_robin`,
		},
	},
}

type loadbalancing struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new Load Balancer annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return loadbalancing{
		r:                r,
		annotationConfig: loadBalanceAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
// used to indicate if the location/s contains a fragment of
// configuration to be included inside the paths of the rules
func (a loadbalancing) Parse(ing *networking.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation(loadBalanceAlghoritmAnnotation, ing, a.annotationConfig.Annotations)
}

func (a loadbalancing) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a loadbalancing) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, loadBalanceAnnotations.Annotations)
}
