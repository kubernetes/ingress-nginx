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

package publishservice

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type publishService struct {
	r resolver.Resolver
}

// NewParser creates a new publish service annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return publishService{r}
}

// Parse parses the annotations contained in the ingress rule
// used to identify the ingress service that is actually connecting this ingrees
// this overwrites the automatically detected address of the controller pod
// and also the service specified globally with the publishservice parameter
// on a per ingress level
func (a publishService) Parse(ing *extensions.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation("publish-service", ing)
}
