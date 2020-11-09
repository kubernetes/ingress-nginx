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
	"k8s.io/ingress-nginx/internal/k8s"
)

const (
	// IngressKey picks a specific "class" for the Ingress.
	// The controller only processes Ingresses with this annotation either
	// unset, or set to either the configured value or the empty string.
	IngressKey = "kubernetes.io/ingress.class"
)

var (
	// DefaultClass defines the default class used in the nginx ingress controller
	DefaultClass = "nginx"

	// IngressClass sets the runtime ingress class to use
	// An empty string means accept all ingresses without
	// annotation and the ones configured with class nginx
	IngressClass = "nginx"
)

// IsValid returns true if the given Ingress specify the ingress.class
// annotation or IngressClassName resource for Kubernetes >= v1.18
func IsValid(ing *networking.Ingress) bool {
	// 1. with annotation or IngressClass
	ingress, ok := ing.GetAnnotations()[IngressKey]
	if !ok && ing.Spec.IngressClassName != nil {
		ingress = *ing.Spec.IngressClassName
	}

	// empty ingress and IngressClass equal default
	if len(ingress) == 0 && IngressClass == DefaultClass {
		return true
	}

	// k8s > v1.18.
	// Processing may be redundant because k8s.IngressClass is obtained by IngressClass
	// 3. without annotation and IngressClass. Check IngressClass
	if k8s.IngressClass != nil {
		return ingress == k8s.IngressClass.Name
	}

	// 4. with IngressClass
	return ingress == IngressClass
}
