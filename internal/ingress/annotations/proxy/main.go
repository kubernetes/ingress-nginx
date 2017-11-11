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

package proxy

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns the proxy timeout to use in the upstream server/s
type Config struct {
	BodySize         string `json:"bodySize"`
	ConnectTimeout   int    `json:"connectTimeout"`
	SendTimeout      int    `json:"sendTimeout"`
	ReadTimeout      int    `json:"readTimeout"`
	BufferSize       string `json:"bufferSize"`
	CookieDomain     string `json:"cookieDomain"`
	CookiePath       string `json:"cookiePath"`
	NextUpstream     string `json:"nextUpstream"`
	PassParams       string `json:"passParams"`
	RequestBuffering string `json:"requestBuffering"`
}

// Equal tests for equality between two Configuration types
func (l1 *Config) Equal(l2 *Config) bool {
	if l1 == l2 {
		return true
	}
	if l1 == nil || l2 == nil {
		return false
	}
	if l1.BodySize != l2.BodySize {
		return false
	}
	if l1.ConnectTimeout != l2.ConnectTimeout {
		return false
	}
	if l1.SendTimeout != l2.SendTimeout {
		return false
	}
	if l1.ReadTimeout != l2.ReadTimeout {
		return false
	}
	if l1.BufferSize != l2.BufferSize {
		return false
	}
	if l1.CookieDomain != l2.CookieDomain {
		return false
	}
	if l1.CookiePath != l2.CookiePath {
		return false
	}
	if l1.NextUpstream != l2.NextUpstream {
		return false
	}
	if l1.PassParams != l2.PassParams {
		return false
	}

	if l1.RequestBuffering != l2.RequestBuffering {
		return false
	}

	return true
}

type proxy struct {
	r resolver.Resolver
}

// NewParser creates a new reverse proxy configuration annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return proxy{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure upstream check parameters
func (a proxy) Parse(ing *extensions.Ingress) (interface{}, error) {
	defBackend := a.r.GetDefaultBackend()
	ct, err := parser.GetIntAnnotation("proxy-connect-timeout", ing, a.r)
	if err != nil {
		ct = defBackend.ProxyConnectTimeout
	}

	st, err := parser.GetIntAnnotation("proxy-send-timeout", ing, a.r)
	if err != nil {
		st = defBackend.ProxySendTimeout
	}

	rt, err := parser.GetIntAnnotation("proxy-read-timeout", ing, a.r)
	if err != nil {
		rt = defBackend.ProxyReadTimeout
	}

	bufs, err := parser.GetStringAnnotation("proxy-buffer-size", ing, a.r)
	if err != nil || bufs == "" {
		bufs = defBackend.ProxyBufferSize
	}

	cp, err := parser.GetStringAnnotation("proxy-cookie-path", ing, a.r)
	if err != nil || cp == "" {
		cp = defBackend.ProxyCookiePath
	}

	cd, err := parser.GetStringAnnotation("proxy-cookie-domain", ing, a.r)
	if err != nil || cd == "" {
		cd = defBackend.ProxyCookieDomain
	}

	bs, err := parser.GetStringAnnotation("proxy-body-size", ing, a.r)
	if err != nil || bs == "" {
		bs = defBackend.ProxyBodySize
	}

	nu, err := parser.GetStringAnnotation("proxy-next-upstream", ing, a.r)
	if err != nil || nu == "" {
		nu = defBackend.ProxyNextUpstream
	}

	pp, err := parser.GetStringAnnotation("proxy-pass-params", ing, a.r)
	if err != nil || pp == "" {
		pp = defBackend.ProxyPassParams
	}

	rb, err := parser.GetStringAnnotation("proxy-request-buffering", ing, a.r)
	if err != nil || rb == "" {
		rb = defBackend.ProxyRequestBuffering
	}

	return &Config{bs, ct, st, rt, bufs, cd, cp, nu, pp, rb}, nil
}
