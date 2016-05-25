/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"strconv"

	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	rewrite = "ingress-nginx.kubernetes.io/rewrite-to"
	fixUrls = "ingress-nginx.kubernetes.io/fix-urls"
)

// ErrMissingAnnotations is returned when the ingress rule
// does not contains annotations related with redirect or strip prefix
type ErrMissingAnnotations struct {
	msg string
}

func (e ErrMissingAnnotations) Error() string {
	return e.msg
}

// Redirect returns authentication configuration for an Ingress rule
type Redirect struct {
	// To URI where the traffic must be redirected
	To string
	// Rewrite indicates if is required to change the
	// links in the response from the upstream servers
	Rewrite bool
}

type ingAnnotations map[string]string

func (a ingAnnotations) rewrite() string {
	val, ok := a[rewrite]
	if ok {
		return val
	}

	return ""
}

func (a ingAnnotations) fixUrls() bool {
	val, ok := a[fixUrls]
	if ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}

	return false
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func ParseAnnotations(ing *extensions.Ingress) (*Redirect, error) {
	if ing.GetAnnotations() == nil {
		return &Redirect{}, ErrMissingAnnotations{"no annotations present"}
	}

	rt := ingAnnotations(ing.GetAnnotations()).rewrite()
	rw := ingAnnotations(ing.GetAnnotations()).fixUrls()
	return &Redirect{
		To:      rt,
		Rewrite: rw,
	}, nil
}
