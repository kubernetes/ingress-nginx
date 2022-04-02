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
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns the proxy timeout to use in the upstream server/s
type Config struct {
	BodySize             string `json:"bodySize"`
	ConnectTimeout       int    `json:"connectTimeout"`
	SendTimeout          int    `json:"sendTimeout"`
	ReadTimeout          int    `json:"readTimeout"`
	BuffersNumber        int    `json:"buffersNumber"`
	BufferSize           string `json:"bufferSize"`
	CookieDomain         string `json:"cookieDomain"`
	CookiePath           string `json:"cookiePath"`
	NextUpstream         string `json:"nextUpstream"`
	NextUpstreamTimeout  int    `json:"nextUpstreamTimeout"`
	NextUpstreamTries    int    `json:"nextUpstreamTries"`
	ProxyRedirectFrom    string `json:"proxyRedirectFrom"`
	ProxyRedirectTo      string `json:"proxyRedirectTo"`
	RequestBuffering     string `json:"requestBuffering"`
	ProxyBuffering       string `json:"proxyBuffering"`
	ProxyHTTPVersion     string `json:"proxyHTTPVersion"`
	ProxyMaxTempFileSize string `json:"proxyMaxTempFileSize"`
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
	if l1.BuffersNumber != l2.BuffersNumber {
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
	if l1.NextUpstreamTimeout != l2.NextUpstreamTimeout {
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
	if l1.ProxyHTTPVersion != l2.ProxyHTTPVersion {
		return false
	}

	if l1.ProxyMaxTempFileSize != l2.ProxyMaxTempFileSize {
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
func (a proxy) Parse(ing *networking.Ingress) (interface{}, error) {
	defBackend := a.r.GetDefaultBackend()
	config := &Config{}

	var err error

	config.ConnectTimeout, err = parser.GetIntAnnotation("proxy-connect-timeout", ing)
	if err != nil {
		config.ConnectTimeout = defBackend.ProxyConnectTimeout
	}

	config.SendTimeout, err = parser.GetIntAnnotation("proxy-send-timeout", ing)
	if err != nil {
		config.SendTimeout = defBackend.ProxySendTimeout
	}

	config.ReadTimeout, err = parser.GetIntAnnotation("proxy-read-timeout", ing)
	if err != nil {
		config.ReadTimeout = defBackend.ProxyReadTimeout
	}

	config.BuffersNumber, err = parser.GetIntAnnotation("proxy-buffers-number", ing)
	if err != nil {
		config.BuffersNumber = defBackend.ProxyBuffersNumber
	}

	config.BufferSize, err = parser.GetStringAnnotation("proxy-buffer-size", ing)
	if err != nil {
		config.BufferSize = defBackend.ProxyBufferSize
	}

	config.CookiePath, err = parser.GetStringAnnotation("proxy-cookie-path", ing)
	if err != nil {
		config.CookiePath = defBackend.ProxyCookiePath
	}

	config.CookieDomain, err = parser.GetStringAnnotation("proxy-cookie-domain", ing)
	if err != nil {
		config.CookieDomain = defBackend.ProxyCookieDomain
	}

	config.BodySize, err = parser.GetStringAnnotation("proxy-body-size", ing)
	if err != nil {
		config.BodySize = defBackend.ProxyBodySize
	}

	config.NextUpstream, err = parser.GetStringAnnotation("proxy-next-upstream", ing)
	if err != nil {
		config.NextUpstream = defBackend.ProxyNextUpstream
	}

	config.NextUpstreamTimeout, err = parser.GetIntAnnotation("proxy-next-upstream-timeout", ing)
	if err != nil {
		config.NextUpstreamTimeout = defBackend.ProxyNextUpstreamTimeout
	}

	config.NextUpstreamTries, err = parser.GetIntAnnotation("proxy-next-upstream-tries", ing)
	if err != nil {
		config.NextUpstreamTries = defBackend.ProxyNextUpstreamTries
	}

	config.RequestBuffering, err = parser.GetStringAnnotation("proxy-request-buffering", ing)
	if err != nil {
		config.RequestBuffering = defBackend.ProxyRequestBuffering
	}

	config.ProxyRedirectFrom, err = parser.GetStringAnnotation("proxy-redirect-from", ing)
	if err != nil {
		config.ProxyRedirectFrom = defBackend.ProxyRedirectFrom
	}

	config.ProxyRedirectTo, err = parser.GetStringAnnotation("proxy-redirect-to", ing)
	if err != nil {
		config.ProxyRedirectTo = defBackend.ProxyRedirectTo
	}

	config.ProxyBuffering, err = parser.GetStringAnnotation("proxy-buffering", ing)
	if err != nil {
		config.ProxyBuffering = defBackend.ProxyBuffering
	}

	config.ProxyHTTPVersion, err = parser.GetStringAnnotation("proxy-http-version", ing)
	if err != nil {
		config.ProxyHTTPVersion = defBackend.ProxyHTTPVersion
	}

	config.ProxyMaxTempFileSize, err = parser.GetStringAnnotation("proxy-max-temp-file-size", ing)
	if err != nil {
		config.ProxyMaxTempFileSize = defBackend.ProxyMaxTempFileSize
	}

	return config, nil
}
