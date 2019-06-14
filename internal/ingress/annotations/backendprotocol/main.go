/*
Copyright 2018 The Kubernetes Authors.

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

package backendprotocol

import (
	"regexp"
	"strings"

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// HTTP protocol
const HTTP = "HTTP"

var (
	validProtocols = regexp.MustCompile(`^(HTTP|HTTPS|AJP|GRPC|GRPCS)$`)
)

type backendProtocol struct {
	r resolver.Resolver
}

// NewParser creates a new backend protocol annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return backendProtocol{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to indicate the backend protocol.
func (a backendProtocol) Parse(ing *extensions.Ingress) (interface{}, error) {
	klog.Infof("Parsing backend protocol annotation")
	if ing.GetAnnotations() == nil {
		return HTTP, nil
	}

	// Proofpoint hack to make v0.24.1 compatible with deprecated "secure-backend" annotation
	// check backend-protocol first; if it exists, apply it, else check for secure-backends.

	proto, err := parser.GetStringAnnotation("backend-protocol", ing)
	if err == nil {
		proto = strings.TrimSpace(strings.ToUpper(proto))
		if !validProtocols.MatchString(proto) {
			klog.Warningf("Protocol %v is not a valid value for the backend-protocol annotation. Using HTTP as protocol", proto)
			return HTTP, nil
		}
		return proto, nil
	}

	secure, err := parser.GetBoolAnnotation("secure-backends", ing)
	if err == nil && secure == true {
		klog.Infof("Parsing backend protocol annotation: secure-backends is true")
		return "HTTPS", nil
	}
	return HTTP, nil
}
