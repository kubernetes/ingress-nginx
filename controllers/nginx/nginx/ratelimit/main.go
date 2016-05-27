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

package ratelimit

import (
	"errors"
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	limitIp  = "ingress-nginx.kubernetes.io/limit-connections"
	limitRps = "ingress-nginx.kubernetes.io/limit-rps"

	// allow 5 times the specified limit as burst
	defBurst = 5

	// 1MB -> 16 thousand 64-byte states or about 8 thousand 128-byte states
	// default is 5MB
	defSharedSize = 5
)

var (
	// ErrInvalidRateLimit is returned when the annotation caontains invalid values
	ErrInvalidRateLimit = errors.New("invalid rate limit value. Must be > 0")
)

// ErrMissingAnnotations is returned when the ingress rule
// does not contains annotations related with rate limit
type ErrMissingAnnotations struct {
	msg string
}

func (e ErrMissingAnnotations) Error() string {
	return e.msg
}

// RateLimit returns rate limit configuration for an Ingress rule
// Is possible to limit the number of connections per IP address or
// connections per second.
// Note: Is possible to specify both limits
type RateLimit struct {
	// Connections indicates a limit with the number of connections per IP address
	Connections Zone
	// RPS indicates a limit with the number of connections per second
	RPS Zone
}

// Zone returns information about the rate limit
type Zone struct {
	Name  string
	Limit int
	Burst int
	// SharedSize amount of shared memory for the zone
	SharedSize int
}

type ingAnnotations map[string]string

func (a ingAnnotations) limitIp() int {
	val, ok := a[limitIp]
	if ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return -1
}

func (a ingAnnotations) limitRps() int {
	val, ok := a[limitRps]
	if ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return -1
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func ParseAnnotations(ing *extensions.Ingress) (*RateLimit, error) {
	if ing.GetAnnotations() == nil {
		return &RateLimit{}, ErrMissingAnnotations{"no annotations present"}
	}

	rps := ingAnnotations(ing.GetAnnotations()).limitRps()
	conn := ingAnnotations(ing.GetAnnotations()).limitIp()

	if rps == 0 && conn == 0 {
		return &RateLimit{
			Connections: Zone{"", -1, -1, 1},
			RPS:         Zone{"", -1, -1, 1},
		}, ErrInvalidRateLimit
	}

	zoneName := fmt.Sprintf("%v_%v", ing.GetNamespace(), ing.GetName())

	return &RateLimit{
		Connections: Zone{
			Name:       fmt.Sprintf("%v_conn", zoneName),
			Limit:      conn,
			Burst:      conn * defBurst,
			SharedSize: defSharedSize,
		},
		RPS: Zone{
			Name:       fmt.Sprintf("%v_rps", zoneName),
			Limit:      rps,
			Burst:      conn * defBurst,
			SharedSize: defSharedSize,
		},
	}, nil
}
