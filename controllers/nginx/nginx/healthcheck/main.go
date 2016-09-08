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
	"errors"
	"strconv"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
)

const (
	upsMaxFails    = "ingress.kubernetes.io/upstream-max-fails"
	upsFailTimeout = "ingress.kubernetes.io/upstream-fail-timeout"
)

var (
	// ErrMissingMaxFails returned error when the ingress does not contains the
	// max-fails annotation
	ErrMissingMaxFails = errors.New("max-fails annotations is missing")

	// ErrMissingFailTimeout returned error when the ingress does not contains
	// the fail-timeout annotation
	ErrMissingFailTimeout = errors.New("fail-timeout annotations is missing")

	// ErrInvalidNumber returned
	ErrInvalidNumber = errors.New("the annotation does not contains a number")
)

// Upstream returns the URL and method to use check the status of
// the upstream server/s
type Upstream struct {
	MaxFails    int
	FailTimeout int
}

type ingAnnotations map[string]string

func (a ingAnnotations) maxFails() (int, error) {
	val, ok := a[upsMaxFails]
	if !ok {
		return 0, ErrMissingMaxFails
	}

	mf, err := strconv.Atoi(val)
	if err != nil {
		return 0, ErrInvalidNumber
	}

	return mf, nil
}

func (a ingAnnotations) failTimeout() (int, error) {
	val, ok := a[upsFailTimeout]
	if !ok {
		return 0, ErrMissingFailTimeout
	}

	ft, err := strconv.Atoi(val)
	if err != nil {
		return 0, ErrInvalidNumber
	}

	return ft, nil
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure upstream check parameters
func ParseAnnotations(cfg config.Configuration, ing *extensions.Ingress) *Upstream {
	if ing.GetAnnotations() == nil {
		return &Upstream{cfg.UpstreamMaxFails, cfg.UpstreamFailTimeout}
	}

	mf, err := ingAnnotations(ing.GetAnnotations()).maxFails()
	if err != nil {
		mf = cfg.UpstreamMaxFails
	}

	ft, err := ingAnnotations(ing.GetAnnotations()).failTimeout()
	if err != nil {
		ft = cfg.UpstreamFailTimeout
	}

	return &Upstream{mf, ft}
}
