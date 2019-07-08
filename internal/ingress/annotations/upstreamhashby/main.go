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
	networking "k8s.io/api/networking/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type upstreamhashby struct {
	r resolver.Resolver
}

// Config contains the Consistent hash configuration to be used in the Ingress
type Config struct {
	UpstreamHashBy           string `json:"upstream-hash-by,omitempty"`
	UpstreamHashBySubset     bool   `json:"upstream-hash-by-subset,omitempty"`
	UpstreamHashBySubsetSize int    `json:"upstream-hash-by-subset-size,omitempty"`
}

// NewParser creates a new UpstreamHashBy annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return upstreamhashby{r}
}

// Parse parses the annotations contained in the ingress rule
func (a upstreamhashby) Parse(ing *networking.Ingress) (interface{}, error) {
	upstreamHashBy, _ := parser.GetStringAnnotation("upstream-hash-by", ing)
	upstreamHashBySubset, _ := parser.GetBoolAnnotation("upstream-hash-by-subset", ing)
	upstreamHashbySubsetSize, _ := parser.GetIntAnnotation("upstream-hash-by-subset-size", ing)

	if upstreamHashbySubsetSize == 0 {
		upstreamHashbySubsetSize = 3
	}

	return &Config{upstreamHashBy, upstreamHashBySubset, upstreamHashbySubsetSize}, nil
}
