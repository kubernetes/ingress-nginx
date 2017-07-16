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
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	annotation = "ingress.kubernetes.io/use-port-in-redirects"
)

type portInRedirect struct {
	backendResolver resolver.DefaultBackend
}

// NewParser creates a new port in redirect annotation parser
func NewParser(db resolver.DefaultBackend) parser.IngressAnnotation {
	return portInRedirect{db}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the redirects must
func (a portInRedirect) Parse(ing *extensions.Ingress) (interface{}, error) {
	up, err := parser.GetBoolAnnotation(annotation, ing)
	if err != nil {
		return a.backendResolver.GetDefaultBackend().UsePortInRedirects, nil
	}

	return up, nil
}
