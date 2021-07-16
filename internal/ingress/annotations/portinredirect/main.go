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

package portinredirect

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type portInRedirect struct {
	r resolver.Resolver
}

// NewParser creates a new port in redirect annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return portInRedirect{r}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the redirects must
func (a portInRedirect) Parse(ing *networking.Ingress) (interface{}, error) {
	up, err := parser.GetBoolAnnotation("use-port-in-redirects", ing)
	if err != nil {
		return a.r.GetDefaultBackend().UsePortInRedirects, nil
	}

	return up, nil
}
