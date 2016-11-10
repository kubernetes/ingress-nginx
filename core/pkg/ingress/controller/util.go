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

package controller

import (
	"strings"

	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/annotations/parser"

	"k8s.io/kubernetes/pkg/apis/extensions"
)

// newDefaultServer return an UpstreamServer to be use as default server that returns 503.
func newDefaultServer() ingress.UpstreamServer {
	return ingress.UpstreamServer{Address: "127.0.0.1", Port: "8181"}
}

// newUpstream creates an upstream without servers.
func newUpstream(name string) *ingress.Upstream {
	return &ingress.Upstream{
		Name:     name,
		Backends: []ingress.UpstreamServer{},
	}
}

func isHostValid(host string, cert *ingress.SSLCert) bool {
	if cert == nil {
		return false
	}
	for _, cn := range cert.CN {
		if matchHostnames(cn, host) {
			return true
		}
	}

	return false
}

func matchHostnames(pattern, host string) bool {
	host = strings.TrimSuffix(host, ".")
	pattern = strings.TrimSuffix(pattern, ".")

	if len(pattern) == 0 || len(host) == 0 {
		return false
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")

	if len(patternParts) != len(hostParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return false
		}
	}

	return true
}

// IsValidClass returns true if the given Ingress either doesn't specify
// the ingress.class annotation, or it's set to the configured in the
// ingress controller.
func IsValidClass(ing *extensions.Ingress, class string) bool {
	if class == "" {
		return true
	}

	cc, _ := parser.GetStringAnnotation(ingressClassKey, ing)
	if cc == "" {
		return true
	}

	return cc == class
}
