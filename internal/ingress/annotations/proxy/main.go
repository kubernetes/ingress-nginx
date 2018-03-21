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
	BodySize          string `json:"bodySize"`
	ConnectTimeout    int    `json:"connectTimeout"`
	SendTimeout       int    `json:"sendTimeout"`
	ReadTimeout       int    `json:"readTimeout"`
	BufferSize        string `json:"bufferSize"`
	CookieDomain      string `json:"cookieDomain"`
	CookiePath        string `json:"cookiePath"`
	NextUpstream      string `json:"nextUpstream"`
	NextUpstreamTries int    `json:"nextUpstreamTries"`
	ProxyRedirectFrom string `json:"proxyRedirectFrom"`
	ProxyRedirectTo   string `json:"proxyRedirectTo"`
	RequestBuffering  string `json:"requestBuffering"`
	ProxyBuffering    string `json:"proxyBuffering"`
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
	if l1.NextUpstreamTries != l2.NextUpstreamTries {
		return false
	}
	if l1.RequestBuffering != l2.RequestBuffering {
		return false
	}
	if l1.ProxyRedirectFrom != l2.ProxyRedirectFrom {
		return false
	}
	if l1.ProxyRedirectTo != l2.ProxyRedirectTo {
		return false
	}
	if l1.ProxyBuffering != l2.ProxyBuffering {
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
	ct, err := parser.GetIntAnnotation("proxy-connect-timeout", ing)
	if err != nil {
		ct = defBackend.ProxyConnectTimeout
	}

	st, err := parser.GetIntAnnotation("proxy-send-timeout", ing)
	if err != nil {
		st = defBackend.ProxySendTimeout
	}

	rt, err := parser.GetIntAnnotation("proxy-read-timeout", ing)
	if err != nil {
		rt = defBackend.ProxyReadTimeout
	}

	bufs, err := parser.GetStringAnnotation("proxy-buffer-size", ing)
	if err != nil || bufs == "" {
		bufs = defBackend.ProxyBufferSize
	}

	cp, err := parser.GetStringAnnotation("proxy-cookie-path", ing)
	if err != nil || cp == "" {
		cp = defBackend.ProxyCookiePath
	}

	cd, err := parser.GetStringAnnotation("proxy-cookie-domain", ing)
	if err != nil || cd == "" {
		cd = defBackend.ProxyCookieDomain
	}

	bs, err := parser.GetStringAnnotation("proxy-body-size", ing)
	if err != nil || bs == "" {
		bs = defBackend.ProxyBodySize
	}

	nu, err := parser.GetStringAnnotation("proxy-next-upstream", ing)
	if err != nil || nu == "" {
		nu = defBackend.ProxyNextUpstream
	}

	nut, err := parser.GetIntAnnotation("proxy-next-upstream-tries", ing)
	if err != nil {
		nut = defBackend.ProxyNextUpstreamTries
	}

	rb, err := parser.GetStringAnnotation("proxy-request-buffering", ing)
	if err != nil || rb == "" {
		rb = defBackend.ProxyRequestBuffering
	}

	prf, err := parser.GetStringAnnotation("proxy-redirect-from", ing)
	if err != nil || prf == "" {
		prf = defBackend.ProxyRedirectFrom
	}

	prt, err := parser.GetStringAnnotation("proxy-redirect-to", ing)
	if err != nil || rb == "" {
		prt = defBackend.ProxyRedirectTo
	}

	pb, err := parser.GetStringAnnotation("proxy-buffering", ing)
	if err != nil || pb == "" {
		pb = defBackend.ProxyBuffering
	}

	return &Config{bs, ct, st, rt, bufs, cd, cp, nu, nut, prf, prt, rb, pb}, nil
}
