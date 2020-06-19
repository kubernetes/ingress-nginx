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
	// 1. with annotation
	ingress, ok := ing.GetAnnotations()[IngressKey]
	if ok {
		// empty annotation and same annotation on ingress
		if ingress == "" && IngressClass == DefaultClass {
			return true
		}

		return ingress == IngressClass
	}

	// 2. k8s < v1.18. Check default annotation
	if !k8s.IsIngressV1Ready {
		return IngressClass == DefaultClass
	}

	// 3. without annotation and IngressClass. Check default annotation
	if k8s.IngressClass == nil {
		return IngressClass == DefaultClass
	}

	// 4. with IngressClass
	return k8s.IngressClass.Name == *ing.Spec.IngressClassName
}
