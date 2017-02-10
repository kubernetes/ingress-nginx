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

package stickysession

import (
	"regexp"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
)

const (
	stickyEnabled     = "ingress.kubernetes.io/sticky-enabled"
	stickyName        = "ingress.kubernetes.io/sticky-name"
	stickyHash        = "ingress.kubernetes.io/sticky-hash"
	defaultStickyHash = "md5"
	defaultStickyName = "route"
)

var (
	stickyHashRegex = regexp.MustCompile(`index|md5|sha1`)
)

// StickyConfig describes the per ingress sticky session config
type StickyConfig struct {
	// The name of the cookie that will be used as stickness router.
	Name string `json:"name"`
	// If sticky must or must not be enabled
	Enabled bool `json:"enabled"`
	// The hash that will be used to encode the cookie
	Hash string `json:"hash"`
}

type sticky struct {
}

// NewParser creates a new Sticky annotation parser
func NewParser() parser.IngressAnnotation {
	return sticky{}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure the sticky directives
func (a sticky) Parse(ing *extensions.Ingress) (interface{}, error) {
	// Check if the sticky is enabled
	se, err := parser.GetBoolAnnotation(stickyEnabled, ing)
	if err != nil {
		return nil, err
	}

	// Get the Sticky Cookie Name
	sn, err := parser.GetStringAnnotation(stickyName, ing)
	if err != nil {
		return nil, err
	}

	if sn == "" {
		sn = defaultStickyName
	}

	sh, err := parser.GetStringAnnotation(stickyHash, ing)

	if err != nil {
		return nil, err
	}

	if sh == "" {
		sh = defaultStickyHash
	}

	if !stickyHashRegex.MatchString(sh) {
		return nil, ing_errors.NewInvalidAnnotationContent(stickyHash, sh)
	}

	return &StickyConfig{
		Name:    sn,
		Enabled: se,
		Hash:    sh,
	}, nil
}
