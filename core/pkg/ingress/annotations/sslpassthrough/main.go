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

package sslpassthrough

import (
	"fmt"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

const (
	passthrough = "ingress.kubernetes.io/ssl-passthrough"
)

// ParseAnnotations parses the annotations contained in the ingress
// rule used to indicate if is required to configure
func ParseAnnotations(cfg defaults.Backend, ing *extensions.Ingress) (bool, error) {

	if ing.GetAnnotations() == nil {
		return false, parser.ErrMissingAnnotations
	}

	if len(ing.Spec.TLS) == 0 {
		return false, fmt.Errorf("ingres rule %v/%v does not contains a TLS section", ing.Name, ing.Namespace)
	}

	return parser.GetBoolAnnotation(passthrough, ing)
}
