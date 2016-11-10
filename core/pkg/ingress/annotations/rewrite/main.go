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

package rewrite

import (
	"errors"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/defaults"

	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	rewriteTo   = "ingress.kubernetes.io/rewrite-target"
	addBaseURL  = "ingress.kubernetes.io/add-base-url"
	sslRedirect = "ingress.kubernetes.io/ssl-redirect"
)

// Redirect describes the per location redirect config
type Redirect struct {
	// Target URI where the traffic must be redirected
	Target string
	// AddBaseURL indicates if is required to add a base tag in the head
	// of the responses from the upstream servers
	AddBaseURL bool
	// SSLRedirect indicates if the location section is accessible SSL only
	SSLRedirect bool
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func ParseAnnotations(cfg defaults.Backend, ing *extensions.Ingress) (*Redirect, error) {
	if ing.GetAnnotations() == nil {
		return &Redirect{}, errors.New("no annotations present")
	}

	sslRe, err := parser.GetBoolAnnotation(sslRedirect, ing)
	if err != nil {
		sslRe = cfg.SSLRedirect
	}

	rt, _ := parser.GetStringAnnotation(rewriteTo, ing)
	abu, _ := parser.GetBoolAnnotation(addBaseURL, ing)
	return &Redirect{
		Target:      rt,
		AddBaseURL:  abu,
		SSLRedirect: sslRe,
	}, nil
}
