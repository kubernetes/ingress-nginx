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
	"strconv"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
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
	// Should indicates if the location section should be accessible SSL only
	SSLRedirect bool
}

var (
	// ErrMissingSSLRedirect returned error when the ingress does not contains the
	// ssl-redirect annotation
	ErrMissingSSLRedirect = errors.New("ssl-redirect annotations is missing")

	// ErrInvalidBool gets returned when the str value is not convertible to a bool
	ErrInvalidBool = errors.New("ssl-redirect annotations has invalid value")
)

type ingAnnotations map[string]string

func (a ingAnnotations) addBaseURL() bool {
	val, ok := a[addBaseURL]
	if ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return false
}

func (a ingAnnotations) rewriteTo() string {
	val, ok := a[rewriteTo]
	if ok {
		return val
	}
	return ""
}

func (a ingAnnotations) sslRedirect() (bool, error) {
	val, ok := a[sslRedirect]
	if !ok {
		return false, ErrMissingSSLRedirect
	}

	sr, err := strconv.ParseBool(val)
	if err != nil {
		return false, ErrInvalidBool
	}

	return sr, nil
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func ParseAnnotations(cfg config.Configuration, ing *extensions.Ingress) (*Redirect, error) {
	if ing.GetAnnotations() == nil {
		return &Redirect{}, errors.New("no annotations present")
	}

	annotations := ingAnnotations(ing.GetAnnotations())

	sslRe, err := annotations.sslRedirect()
	if err != nil {
		sslRe = cfg.SSLRedirect
	}

	rt := annotations.rewriteTo()
	abu := annotations.addBaseURL()
	return &Redirect{
		Target:      rt,
		AddBaseURL:  abu,
		SSLRedirect: sslRe,
	}, nil
}
