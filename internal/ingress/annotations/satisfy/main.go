/*
Copyright 2019 The Kubernetes Authors.

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

package satisfy

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type satisfy struct {
	r resolver.Resolver
}

// NewParser creates a new SATISFY annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return satisfy{r}
}

// Parse parses annotation contained in the ingress
func (s satisfy) Parse(ing *networking.Ingress) (interface{}, error) {
	satisfy, err := parser.GetStringAnnotation("satisfy", ing)

	if err != nil || (satisfy != "any" && satisfy != "all") {
		satisfy = ""
	}

	return satisfy, nil
}
