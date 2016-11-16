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
	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

const (
	upsMaxFails    = "ingress.kubernetes.io/upstream-max-fails"
	upsFailTimeout = "ingress.kubernetes.io/upstream-fail-timeout"
)

// Upstream returns the URL and method to use check the status of
// the upstream server/s
type Upstream struct {
	MaxFails    int `json:"maxFails"`
	FailTimeout int `json:"failTimeout"`
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure upstream check parameters
func ParseAnnotations(cfg defaults.Backend, ing *extensions.Ingress) *Upstream {
	if ing.GetAnnotations() == nil {
		return &Upstream{cfg.UpstreamMaxFails, cfg.UpstreamFailTimeout}
	}

	mf, err := parser.GetIntAnnotation(upsMaxFails, ing)
	if err != nil {
		mf = cfg.UpstreamMaxFails
	}

	ft, err := parser.GetIntAnnotation(upsFailTimeout, ing)
	if err != nil {
		ft = cfg.UpstreamFailTimeout
	}

	return &Upstream{mf, ft}
}
