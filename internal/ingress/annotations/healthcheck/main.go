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

package healthcheck

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns the URL and method to use check the status of
// the upstream server/s
type Config struct {
	MaxFails    int `json:"maxFails"`
	FailTimeout int `json:"failTimeout"`
}

type healthCheck struct {
	r resolver.Resolver
}

// NewParser creates a new health check annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return healthCheck{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure upstream check parameters
func (hc healthCheck) Parse(ing *extensions.Ingress) (interface{}, error) {
	defBackend := hc.r.GetDefaultBackend()
	if ing.GetAnnotations() == nil {
		return &Config{defBackend.UpstreamMaxFails, defBackend.UpstreamFailTimeout}, nil
	}

	mf, err := parser.GetIntAnnotation("upstream-max-fails", ing)
	if err != nil {
		mf = defBackend.UpstreamMaxFails
	}

	ft, err := parser.GetIntAnnotation("upstream-fail-timeout", ing)
	if err != nil {
		ft = defBackend.UpstreamFailTimeout
	}

	return &Config{mf, ft}, nil
}
