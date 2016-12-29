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

package cors

import (
	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
)

const (
	annotation = "ingress.kubernetes.io/enable-cors"
)

type cors struct {
}

// NewParser creates a new CORS annotation parser
func NewParser() parser.IngressAnnotation {
	return cors{}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should allows CORS
func (a cors) Parse(ing *extensions.Ingress) (interface{}, error) {
	return parser.GetBoolAnnotation(annotation, ing)
}
