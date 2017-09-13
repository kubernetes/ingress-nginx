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

package vtsfilterkey

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
)

const (
	annotation = "ingress.kubernetes.io/vts-filter-key"
)

type vtsFilterKey struct {
}

// NewParser creates a new vts filter key annotation parser
func NewParser() parser.IngressAnnotation {
	return vtsFilterKey{}
}

// Parse parses the annotations contained in the ingress rule
// used to indicate if the location/s contains a fragment of
// configuration to be included inside the paths of the rules
func (a vtsFilterKey) Parse(ing *extensions.Ingress) (interface{}, error) {
	return parser.GetStringAnnotation(annotation, ing)
}
