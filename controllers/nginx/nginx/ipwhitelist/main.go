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

package ipwhitelist

import (
	"errors"

	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/net/sets"
)

const (
	whitelist = "ingress.kubernetes.io/whitelist"
)

var (
	// ErrMissingWhitelist returned error when the ingress does not contains the
	// whitelist annotation
	ErrMissingWhitelist = errors.New("whitelist annotation is missing")

	// ErrInvalidCIDR returned error when the whitelist annotation does not
	// contains a valid IP or network address
	ErrInvalidCIDR = errors.New("the annotation does not contains a valid IP address or network")
)

// Whitelist returns the CIDR
type Whitelist struct {
	CIDR []string
}

type ingAnnotations map[string]string

func (a ingAnnotations) whitelist() ([]string, error) {
	val, ok := a[whitelist]
	if !ok {
		return nil, ErrMissingWhitelist
	}

	ipnet, err := sets.ParseIPNets(val)
	if err != nil {
		return nil, ErrInvalidCIDR
	}

	nets := make([]string, 0)
	for k := range ipnet {
		nets = append(nets, k)
	}

	return nets, nil
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure upstream check parameters
func ParseAnnotations(whiteList []string, ing *extensions.Ingress) (*Whitelist, error) {
	if ing.GetAnnotations() == nil {
		return &Whitelist{whiteList}, ErrMissingWhitelist
	}

	wl, err := ingAnnotations(ing.GetAnnotations()).whitelist()
	if err != nil {
		wl = whiteList
	}

	return &Whitelist{wl}, err
}
