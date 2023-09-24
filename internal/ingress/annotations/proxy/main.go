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
	"regexp"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	proxyConnectTimeoutAnnotation      = "proxy-connect-timeout"
	proxySendTimeoutAnnotation         = "proxy-send-timeout"
	proxyReadTimeoutAnnotation         = "proxy-read-timeout"
	proxyBuffersNumberAnnotation       = "proxy-buffers-number"
	proxyBufferSizeAnnotation          = "proxy-buffer-size"
	proxyCookiePathAnnotation          = "proxy-cookie-path"
	proxyCookieDomainAnnotation        = "proxy-cookie-domain"
	proxyBodySizeAnnotation            = "proxy-body-size"
	proxyNextUpstreamAnnotation        = "proxy-next-upstream"
	proxyNextUpstreamTimeoutAnnotation = "proxy-next-upstream-timeout"
	proxyNextUpstreamTriesAnnotation   = "proxy-next-upstream-tries"
	proxyRequestBufferingAnnotation    = "proxy-request-buffering"
	proxyRedirectFromAnnotation        = "proxy-redirect-from"
	proxyRedirectToAnnotation          = "proxy-redirect-to"
	proxyBufferingAnnotation           = "proxy-buffering"
	proxyHTTPVersionAnnotation         = "proxy-http-version"
	proxyMaxTempFileSizeAnnotation     = "proxy-max-temp-file-size" //#nosec G101
)

var validUpstreamAnnotation = regexp.MustCompile(`^((error|timeout|invalid_header|http_500|http_502|http_503|http_504|http_403|http_404|http_429|non_idempotent|off)\s?)+$`)

var proxyAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		proxyConnectTimeoutAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation allows setting the timeout in seconds of the connect operation to the backend.`,
		},
		proxySendTimeoutAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation allows setting the timeout in seconds of the send operation to the backend.`,
		},
		proxyReadTimeoutAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation allows setting the timeout in seconds of the read operation to the backend.`,
		},
		proxyBuffersNumberAnnotation: {
			Validator: parser.ValidateInt,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation sets the number of the buffers in proxy_buffers used for reading the first part of the response received from the proxied server. 
			By default proxy buffers number is set as 4`,
		},
		proxyBufferSizeAnnotation: {
			Validator: parser.ValidateRegex(parser.SizeRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation sets the size of the buffer proxy_buffer_size used for reading the first part of the response received from the proxied server. 
			By default proxy buffer size is set as "4k".`,
		},
		proxyCookiePathAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLIsValidRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation sets a text that should be changed in the path attribute of the "Set-Cookie" header fields of a proxied server response.`,
		},
		proxyCookieDomainAnnotation: {
			Validator:     parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation ets a text that should be changed in the domain attribute of the "Set-Cookie" header fields of a proxied server response.`,
		},
		proxyBodySizeAnnotation: {
			Validator:     parser.ValidateRegex(parser.SizeRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation allows setting the maximum allowed size of a client request body.`,
		},
		proxyNextUpstreamAnnotation: {
			Validator: parser.ValidateRegex(validUpstreamAnnotation, false),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation defines when the next upstream should be used. 
			This annotation reflect the directive https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream 
			and only the allowed values on upstream are allowed here.`,
		},
		proxyNextUpstreamTimeoutAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation limits the time during which a request can be passed to the next server`,
		},
		proxyNextUpstreamTriesAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation limits the number of possible tries for passing a request to the next server`,
		},
		proxyRequestBufferingAnnotation: {
			Validator:     parser.ValidateOptions([]string{"on", "off"}, true, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables or disables buffering of a client request body.`,
		},
		proxyRedirectFromAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLIsValidRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `The annotations proxy-redirect-from and proxy-redirect-to will set the first and second parameters of NGINX's proxy_redirect directive respectively`,
		},
		proxyRedirectToAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLIsValidRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `The annotations proxy-redirect-from and proxy-redirect-to will set the first and second parameters of NGINX's proxy_redirect directive respectively`,
		},
		proxyBufferingAnnotation: {
			Validator:     parser.ValidateOptions([]string{"on", "off"}, true, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables or disables buffering of responses from the proxied server. It can be "on" or "off"`,
		},
		proxyHTTPVersionAnnotation: {
			Validator:     parser.ValidateOptions([]string{"1.0", "1.1"}, true, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotations sets the HTTP protocol version for proxying. Can be "1.0" or "1.1".`,
		},
		proxyMaxTempFileSizeAnnotation: {
			Validator:     parser.ValidateRegex(parser.SizeRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines the maximum size of a temporary file when buffering responses.`,
		},
	},
}

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
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new reverse proxy configuration annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return proxy{
		r:                r,
		annotationConfig: proxyAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure upstream check parameters
func (a proxy) Parse(ing *networking.Ingress) (interface{}, error) {
	defBackend := a.r.GetDefaultBackend()
	config := &Config{}

	var err error

	config.ConnectTimeout, err = parser.GetIntAnnotation(proxyConnectTimeoutAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ConnectTimeout = defBackend.ProxyConnectTimeout
	}

	config.SendTimeout, err = parser.GetIntAnnotation(proxySendTimeoutAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.SendTimeout = defBackend.ProxySendTimeout
	}

	config.ReadTimeout, err = parser.GetIntAnnotation(proxyReadTimeoutAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ReadTimeout = defBackend.ProxyReadTimeout
	}

	config.BuffersNumber, err = parser.GetIntAnnotation(proxyBuffersNumberAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.BuffersNumber = defBackend.ProxyBuffersNumber
	}

	config.BufferSize, err = parser.GetStringAnnotation(proxyBufferSizeAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.BufferSize = defBackend.ProxyBufferSize
	}

	config.CookiePath, err = parser.GetStringAnnotation(proxyCookiePathAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.CookiePath = defBackend.ProxyCookiePath
	}

	config.CookieDomain, err = parser.GetStringAnnotation(proxyCookieDomainAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.CookieDomain = defBackend.ProxyCookieDomain
	}

	config.BodySize, err = parser.GetStringAnnotation(proxyBodySizeAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.BodySize = defBackend.ProxyBodySize
	}

	config.NextUpstream, err = parser.GetStringAnnotation(proxyNextUpstreamAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.NextUpstream = defBackend.ProxyNextUpstream
	}

	config.NextUpstreamTimeout, err = parser.GetIntAnnotation(proxyNextUpstreamTimeoutAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.NextUpstreamTimeout = defBackend.ProxyNextUpstreamTimeout
	}

	config.NextUpstreamTries, err = parser.GetIntAnnotation(proxyNextUpstreamTriesAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.NextUpstreamTries = defBackend.ProxyNextUpstreamTries
	}

	config.RequestBuffering, err = parser.GetStringAnnotation(proxyRequestBufferingAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.RequestBuffering = defBackend.ProxyRequestBuffering
	}

	config.ProxyRedirectFrom, err = parser.GetStringAnnotation(proxyRedirectFromAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ProxyRedirectFrom = defBackend.ProxyRedirectFrom
	}

	config.ProxyRedirectTo, err = parser.GetStringAnnotation(proxyRedirectToAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ProxyRedirectTo = defBackend.ProxyRedirectTo
	}

	config.ProxyBuffering, err = parser.GetStringAnnotation(proxyBufferingAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ProxyBuffering = defBackend.ProxyBuffering
	}

	config.ProxyHTTPVersion, err = parser.GetStringAnnotation(proxyHTTPVersionAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ProxyHTTPVersion = defBackend.ProxyHTTPVersion
	}

	config.ProxyMaxTempFileSize, err = parser.GetStringAnnotation(proxyMaxTempFileSizeAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		config.ProxyMaxTempFileSize = defBackend.ProxyMaxTempFileSize
	}

	return config, nil
}

func (a proxy) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a proxy) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, proxyAnnotations.Annotations)
}
