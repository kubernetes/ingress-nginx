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
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	rewriteTo        = "ingress.kubernetes.io/rewrite-target"
	addBaseURL       = "ingress.kubernetes.io/add-base-url"
	baseURLScheme    = "ingress.kubernetes.io/base-url-scheme"
	sslRedirect      = "ingress.kubernetes.io/ssl-redirect"
	forceSSLRedirect = "ingress.kubernetes.io/force-ssl-redirect"
	appRoot          = "ingress.kubernetes.io/app-root"
)

// Redirect describes the per location redirect config
type Redirect struct {
	// Target URI where the traffic must be redirected
	Target string `json:"target"`
	// AddBaseURL indicates if is required to add a base tag in the head
	// of the responses from the upstream servers
	AddBaseURL bool `json:"addBaseUrl"`
	// BaseURLScheme override for the scheme passed to the base tag
	BaseURLScheme string `json:"baseUrlScheme"`
	// SSLRedirect indicates if the location section is accessible SSL only
	SSLRedirect bool `json:"sslRedirect"`
	// ForceSSLRedirect indicates if the location section is accessible SSL only
	ForceSSLRedirect bool `json:"forceSSLRedirect"`
	// AppRoot defines the Application Root that the Controller must redirect if it's not in '/' context
	AppRoot string `json:"appRoot"`
}

// Equal tests for equality between two Redirect types
func (r1 *Redirect) Equal(r2 *Redirect) bool {
	if r1 == r2 {
		return true
	}
	if r1 == nil || r2 == nil {
		return false
	}
	if r1.Target != r2.Target {
		return false
	}
	if r1.AddBaseURL != r2.AddBaseURL {
		return false
	}
	if r1.BaseURLScheme != r2.BaseURLScheme {
		return false
	}
	if r1.SSLRedirect != r2.SSLRedirect {
		return false
	}
	if r1.ForceSSLRedirect != r2.ForceSSLRedirect {
		return false
	}
	if r1.AppRoot != r2.AppRoot {
		return false
	}

	return true
}

type rewrite struct {
	backendResolver resolver.DefaultBackend
}

// NewParser creates a new reqrite annotation parser
func NewParser(br resolver.DefaultBackend) parser.IngressAnnotation {
	return rewrite{br}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func (a rewrite) Parse(ing *extensions.Ingress) (interface{}, error) {
	rt, _ := parser.GetStringAnnotation(rewriteTo, ing)
	sslRe, err := parser.GetBoolAnnotation(sslRedirect, ing)
	if err != nil {
		sslRe = a.backendResolver.GetDefaultBackend().SSLRedirect
	}
	fSslRe, err := parser.GetBoolAnnotation(forceSSLRedirect, ing)
	if err != nil {
		fSslRe = a.backendResolver.GetDefaultBackend().ForceSSLRedirect
	}
	abu, _ := parser.GetBoolAnnotation(addBaseURL, ing)
	bus, _ := parser.GetStringAnnotation(baseURLScheme, ing)
	ar, _ := parser.GetStringAnnotation(appRoot, ing)
	return &Redirect{
		Target:           rt,
		AddBaseURL:       abu,
		BaseURLScheme:    bus,
		SSLRedirect:      sslRe,
		ForceSSLRedirect: fSslRe,
		AppRoot:          ar,
	}, nil
}
