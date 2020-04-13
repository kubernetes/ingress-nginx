/*
Copyright 2015 The Kubernetes Authors.

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

package class

import (
	networking "k8s.io/api/networking/v1beta1"
)

const (
	// IngressKey picks a specific "class" for the Ingress.
	// The controller only processes Ingresses with this annotation either
	// unset, or set to either the configured value or the empty string.
	IngressKey = "kubernetes.io/ingress.class"
)

var (
	// IngressClass sets the runtime ingress class to use.
	// An empty string means accept all ingresses regardless of their Ingress class.
	// The value of this is set based on `ingress-class` command line argument when controller starts.
	IngressClass string
)

// IsValid decides whether given Ingress should be processed by the running instance of controller.
func IsValid(ing *networking.Ingress) bool {
	if IngressClass == "" {
		return true
	}

	className := ing.Spec.IngressClassName
	if className == nil {
		return true
	}

	return *className == IngressClass
}
