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

package ipwhitelist

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/net/sets"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	whitelist = "ingress.kubernetes.io/whitelist-source-range"
)

// SourceRange returns the CIDR
type SourceRange struct {
	CIDR []string `json:"cidr"`
}

type ipwhitelist struct {
	backendResolver resolver.DefaultBackend
}

// NewParser creates a new whitelist annotation parser
func NewParser(br resolver.DefaultBackend) parser.IngressAnnotation {
	return ipwhitelist{br}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to limit access to certain client addresses or networks.
// Multiple ranges can specified using commas as separator
// e.g. `18.0.0.0/8,56.0.0.0/8`
func (a ipwhitelist) Parse(ing *extensions.Ingress) (interface{}, error) {
	defBackend := a.backendResolver.GetDefaultBackend()
	sort.Strings(defBackend.WhitelistSourceRange)

	val, err := parser.GetStringAnnotation(whitelist, ing)
	if err != nil {
		return &SourceRange{CIDR: defBackend.WhitelistSourceRange}, err
	}

	values := strings.Split(val, ",")
	ipnets, err := sets.ParseIPNets(values...)
	if err != nil {
		return &SourceRange{CIDR: defBackend.WhitelistSourceRange}, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "the annotation does not contains a valid IP address or network"),
		}
	}

	cidrs := []string{}
	for k := range ipnets {
		cidrs = append(cidrs, k)
	}

	sort.Strings(cidrs)

	return &SourceRange{cidrs}, nil
}
