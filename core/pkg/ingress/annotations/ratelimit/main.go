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

package ratelimit

import (
	"fmt"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
)

const (
	limitIP  = "ingress.kubernetes.io/limit-connections"
	limitRPS = "ingress.kubernetes.io/limit-rps"

	// allow 5 times the specified limit as burst
	defBurst = 5

	// 1MB -> 16 thousand 64-byte states or about 8 thousand 128-byte states
	// default is 5MB
	defSharedSize = 5
)

// RateLimit returns rate limit configuration for an Ingress rule
// Is possible to limit the number of connections per IP address or
// connections per second.
// Note: Is possible to specify both limits
type RateLimit struct {
	// Connections indicates a limit with the number of connections per IP address
	Connections Zone `json:"connections"`
	// RPS indicates a limit with the number of connections per second
	RPS Zone `json:"rps"`
}

// Zone returns information about the NGINX rate limit (limit_req_zone)
// http://nginx.org/en/docs/http/ngx_http_limit_req_module.html#limit_req_zone
type Zone struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Burst int    `json:"burst"`
	// SharedSize amount of shared memory for the zone
	SharedSize int `json:"sharedSize"`
}

type ratelimit struct {
}

// NewParser creates a new ratelimit annotation parser
func NewParser() parser.IngressAnnotation {
	return ratelimit{}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to rewrite the defined paths
func (a ratelimit) Parse(ing *extensions.Ingress) (interface{}, error) {

	rps, _ := parser.GetIntAnnotation(limitRPS, ing)
	conn, _ := parser.GetIntAnnotation(limitIP, ing)

	if rps == 0 && conn == 0 {
		return &RateLimit{
			Connections: Zone{},
			RPS:         Zone{},
		}, nil
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
