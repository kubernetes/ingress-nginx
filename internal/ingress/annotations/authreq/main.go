/*
Copyright 2015 The Kubernetes Authors.

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

package authreq

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/klog/v2"

	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/pkg/util/sets"
)

const (
	authReqURLAnnotation                = "auth-url"
	authReqMethodAnnotation             = "auth-method"
	authReqSigninAnnotation             = "auth-signin"
	authReqSigninRedirParamAnnotation   = "auth-signin-redirect-param"
	authReqSnippetAnnotation            = "auth-snippet"
	authReqCacheKeyAnnotation           = "auth-cache-key"
	authReqKeepaliveAnnotation          = "auth-keepalive"
	authReqKeepaliveShareVarsAnnotation = "auth-keepalive-share-vars"
	authReqKeepaliveRequestsAnnotation  = "auth-keepalive-requests"
	authReqKeepaliveTimeout             = "auth-keepalive-timeout"
	authReqCacheDuration                = "auth-cache-duration"
	authReqResponseHeadersAnnotation    = "auth-response-headers"
	authReqProxySetHeadersAnnotation    = "auth-proxy-set-headers"
	authReqRequestRedirectAnnotation    = "auth-request-redirect"
	authReqAlwaysSetCookieAnnotation    = "auth-always-set-cookie"

	// This should be exported as it is imported by other packages
	AuthSecretAnnotation = "auth-secret"
)

var authReqAnnotations = parser.Annotation{
	Group: "authentication",
	Annotations: parser.AnnotationFields{
		authReqURLAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLWithNginxVariableRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation allows to indicate the URL where the HTTP request should be sent`,
		},
		authReqMethodAnnotation: {
			Validator:     parser.ValidateRegex(methodsRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation allows to specify the HTTP method to use`,
		},
		authReqSigninAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLWithNginxVariableRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation allows to specify the location of the error page`,
		},
		authReqSigninRedirParamAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLIsValidRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation allows to specify the URL parameter in the error page which should contain the original URL for a failed signin request`,
		},
		authReqSnippetAnnotation: {
			Validator:     parser.ValidateNull,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskCritical,
			Documentation: `This annotation allows to specify a custom snippet to use with external authentication`,
		},
		authReqCacheKeyAnnotation: {
			Validator:     parser.ValidateRegex(parser.NGINXVariable, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation enables caching for auth requests.`,
		},
		authReqKeepaliveAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation specifies the maximum number of keepalive connections to auth-url. Only takes effect when no variables are used in the host part of the URL`,
		},
		authReqKeepaliveShareVarsAnnotation: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation specifies whether to share Nginx variables among the current request and the auth request`,
		},
		authReqKeepaliveRequestsAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines the maximum number of requests that can be served through one keepalive connection`,
		},
		authReqKeepaliveTimeout: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation specifies a duration in seconds which an idle keepalive connection to an upstream server will stay open`,
		},
		authReqCacheDuration: {
			Validator:     parser.ValidateRegex(parser.ExtendedCharsRegex, false),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation allows to specify a caching time for auth responses based on their response codes, e.g. 200 202 30m`,
		},
		authReqResponseHeadersAnnotation: {
			Validator:     parser.ValidateRegex(parser.HeadersVariable, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation sets the headers to pass to backend once authentication request completes. They should be separated by comma.`,
		},
		authReqProxySetHeadersAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation sets the name of a ConfigMap that specifies headers to pass to the authentication service.
			Only ConfigMaps on the same namespace are allowed`,
		},
		authReqRequestRedirectAnnotation: {
			Validator:     parser.ValidateRegex(parser.URLIsValidRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation allows to specify the X-Auth-Request-Redirect header value`,
		},
		authReqAlwaysSetCookieAnnotation: {
			Validator: parser.ValidateBool,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation enables setting a cookie returned by auth request. 
			By default, the cookie will be set only if an upstream reports with the code 200, 201, 204, 206, 301, 302, 303, 304, 307, or 308`,
		},
	},
}

// Config returns external authentication configuration for an Ingress rule
type Config struct {
	URL string `json:"url"`
	// Host contains the hostname defined in the URL
	Host                   string            `json:"host"`
	SigninURL              string            `json:"signinUrl"`
	SigninURLRedirectParam string            `json:"signinUrlRedirectParam,omitempty"`
	Method                 string            `json:"method"`
	ResponseHeaders        []string          `json:"responseHeaders,omitempty"`
	RequestRedirect        string            `json:"requestRedirect"`
	AuthSnippet            string            `json:"authSnippet"`
	AuthCacheKey           string            `json:"authCacheKey"`
	AuthCacheDuration      []string          `json:"authCacheDuration"`
	KeepaliveConnections   int               `json:"keepaliveConnections"`
	KeepaliveShareVars     bool              `json:"keepaliveShareVars"`
	KeepaliveRequests      int               `json:"keepaliveRequests"`
	KeepaliveTimeout       int               `json:"keepaliveTimeout"`
	ProxySetHeaders        map[string]string `json:"proxySetHeaders,omitempty"`
	AlwaysSetCookie        bool              `json:"alwaysSetCookie,omitempty"`
}

// DefaultCacheDuration is the fallback value if no cache duration is provided
const DefaultCacheDuration = "200 202 401 5m"

// fallback values when no keepalive parameters are set
const (
	defaultKeepaliveConnections = 0
	defaultKeepaliveShareVars   = false
	defaultKeepaliveRequests    = 1000
	defaultKeepaliveTimeout     = 60
)

// Equal tests for equality between two Config types
func (e1 *Config) Equal(e2 *Config) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.URL != e2.URL {
		return false
	}
	if e1.Host != e2.Host {
		return false
	}
	if e1.SigninURL != e2.SigninURL {
		return false
	}
	if e1.SigninURLRedirectParam != e2.SigninURLRedirectParam {
		return false
	}
	if e1.Method != e2.Method {
		return false
	}

	match := sets.StringElementsMatch(e1.ResponseHeaders, e2.ResponseHeaders)
	if !match {
		return false
	}

	if e1.RequestRedirect != e2.RequestRedirect {
		return false
	}
	if e1.AuthSnippet != e2.AuthSnippet {
		return false
	}

	if e1.AuthCacheKey != e2.AuthCacheKey {
		return false
	}

	if e1.KeepaliveConnections != e2.KeepaliveConnections {
		return false
	}

	if e1.KeepaliveShareVars != e2.KeepaliveShareVars {
		return false
	}

	if e1.KeepaliveRequests != e2.KeepaliveRequests {
		return false
	}

	if e1.KeepaliveTimeout != e2.KeepaliveTimeout {
		return false
	}

	if e1.AlwaysSetCookie != e2.AlwaysSetCookie {
		return false
	}

	return sets.StringElementsMatch(e1.AuthCacheDuration, e2.AuthCacheDuration)
}

var (
	methodsRegex    = regexp.MustCompile("(GET|HEAD|POST|PUT|PATCH|DELETE|CONNECT|OPTIONS|TRACE)")
	headerRegexp    = regexp.MustCompile(`^[a-zA-Z\d\-_]+$`)
	statusCodeRegex = regexp.MustCompile(`^\d{3}$`)
	durationRegex   = regexp.MustCompile(`^\d+(ms|s|m|h|d|w|M|y)$`) // see https://nginx.org/en/docs/syntax.html
)

// ValidMethod checks is the provided string a valid HTTP method
func ValidMethod(method string) bool {
	return methodsRegex.MatchString(method)
}

// ValidHeader checks is the provided string satisfies the header's name regex
func ValidHeader(header string) bool {
	return headerRegexp.MatchString(header)
}

// ValidCacheDuration checks if the provided string is a valid cache duration
// spec: [code ...] [time ...];
// with: code is an http status code
//
//	time must match the time regex and may appear multiple times, e.g. `1h 30m`
func ValidCacheDuration(duration string) bool {
	elements := strings.Split(duration, " ")
	seenDuration := false

	for _, element := range elements {
		if element == "" {
			continue
		}
		if statusCodeRegex.MatchString(element) {
			if seenDuration {
				return false // code after duration
			}
			continue
		}
		if durationRegex.MatchString(element) {
			seenDuration = true
		}
	}
	return seenDuration
}

type authReq struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new authentication request annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return authReq{
		r:                r,
		annotationConfig: authReqAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to use an Config URL as source for authentication
//
//nolint:gocyclo // Ignore function complexity error
func (a authReq) Parse(ing *networking.Ingress) (interface{}, error) {
	// Required Parameters
	urlString, err := parser.GetStringAnnotation(authReqURLAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		return nil, err
	}

	authURL, err := parser.StringToURL(urlString)
	if err != nil {
		return nil, ing_errors.LocationDeniedError{Reason: fmt.Errorf("could not parse auth-url annotation: %v", err)}
	}

	authMethod, err := parser.GetStringAnnotation(authReqMethodAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return nil, ing_errors.NewLocationDenied("invalid HTTP method")
		}
	}

	// Optional Parameters
	signIn, err := parser.GetStringAnnotation(authReqSigninAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			klog.Warningf("%s value is invalid: %s", authReqSigninAnnotation, err)
		}
		klog.V(3).InfoS("auth-signin annotation is undefined and will not be set")
	}

	signInRedirectParam, err := parser.GetStringAnnotation(authReqSigninRedirParamAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			klog.Warningf("%s value is invalid: %s", authReqSigninRedirParamAnnotation, err)
		}
		klog.V(3).Infof("auth-signin-redirect-param annotation is undefined and will not be set")
	}

	authSnippet, err := parser.GetStringAnnotation(authReqSnippetAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("auth-snippet annotation is undefined and will not be set")
	}

	authCacheKey, err := parser.GetStringAnnotation(authReqCacheKeyAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			klog.Warningf("%s value is invalid: %s", authReqCacheKeyAnnotation, err)
		}
		klog.V(3).InfoS("auth-cache-key annotation is undefined and will not be set")
	}

	keepaliveConnections, err := parser.GetIntAnnotation(authReqKeepaliveAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("auth-keepalive annotation is undefined and will be set to its default value")
		keepaliveConnections = defaultKeepaliveConnections
	}
	switch {
	case keepaliveConnections < 0:
		klog.Warningf("auth-keepalive annotation (%s) contains a negative value, setting auth-keepalive to 0", authURL.Host)
		keepaliveConnections = 0
	case keepaliveConnections > 0:
		// NOTE: upstream block cannot reference a variable in the server directive
		if strings.IndexByte(authURL.Host, '$') != -1 {
			klog.Warningf("auth-url annotation (%s) contains $ in the host:port part, setting auth-keepalive to 0", authURL.Host)
			keepaliveConnections = 0
		}
	}

	keepaliveShareVars, err := parser.GetBoolAnnotation(authReqKeepaliveShareVarsAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("auth-keepalive-share-vars annotation is undefined and will be set to its default value")
		keepaliveShareVars = defaultKeepaliveShareVars
	}

	keepaliveRequests, err := parser.GetIntAnnotation(authReqKeepaliveRequestsAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("auth-keepalive-requests annotation is undefined or invalid and will be set to its default value")
		keepaliveRequests = defaultKeepaliveRequests
	}
	if keepaliveRequests <= 0 {
		klog.Warningf("auth-keepalive-requests annotation (%s) should be greater than zero, setting auth-keepalive to 0", authURL.Host)
		keepaliveConnections = 0
	}

	keepaliveTimeout, err := parser.GetIntAnnotation(authReqKeepaliveTimeout, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("auth-keepalive-timeout annotation is undefined and will be set to its default value")
		keepaliveTimeout = defaultKeepaliveTimeout
	}
	if keepaliveTimeout <= 0 {
		klog.Warningf("auth-keepalive-timeout annotation (%s) should be greater than zero, setting auth-keepalive 0", authURL.Host)
		keepaliveConnections = 0
	}

	durstr, err := parser.GetStringAnnotation(authReqCacheDuration, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsValidationError(err) {
		return nil, fmt.Errorf("%s contains invalid value", authReqCacheDuration)
	}
	authCacheDuration, err := ParseStringToCacheDurations(durstr)
	if err != nil {
		return nil, err
	}

	responseHeaders := []string{}
	hstr, err := parser.GetStringAnnotation(authReqResponseHeadersAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsValidationError(err) {
		return nil, ing_errors.NewLocationDenied("validation error")
	}
	if hstr != "" {
		harr := strings.Split(hstr, ",")
		for _, header := range harr {
			header = strings.TrimSpace(header)
			if header != "" {
				if !ValidHeader(header) {
					return nil, ing_errors.NewLocationDenied("invalid headers list")
				}
				responseHeaders = append(responseHeaders, header)
			}
		}
	}

	proxySetHeaderMap, err := parser.GetStringAnnotation(authReqProxySetHeadersAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("auth-set-proxy-headers annotation is undefined and will not be set", "err", err)
	}

	cns, _, err := cache.SplitMetaNamespaceKey(proxySetHeaderMap)
	if err != nil {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("error reading configmap name %s from annotation: %w", proxySetHeaderMap, err),
		}
	}

	if cns == "" {
		cns = ing.Namespace
	}

	secCfg := a.r.GetSecurityConfiguration()
	// We don't accept different namespaces for secrets.
	if !secCfg.AllowCrossNamespaceResources && cns != ing.Namespace {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("cross namespace usage of secrets is not allowed"),
		}
	}

	var proxySetHeaders map[string]string

	if proxySetHeaderMap != "" {
		proxySetHeadersMapContents, err := a.r.GetConfigMap(proxySetHeaderMap)
		if err != nil {
			return nil, ing_errors.NewLocationDenied(fmt.Sprintf("unable to find configMap %q", proxySetHeaderMap))
		}

		for header := range proxySetHeadersMapContents.Data {
			if !ValidHeader(header) {
				return nil, ing_errors.NewLocationDenied("invalid proxy-set-headers in configmap")
			}
		}

		proxySetHeaders = proxySetHeadersMapContents.Data
	}

	requestRedirect, err := parser.GetStringAnnotation(authReqRequestRedirectAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsValidationError(err) {
		return nil, fmt.Errorf("%s is invalid: %w", authReqRequestRedirectAnnotation, err)
	}

	alwaysSetCookie, err := parser.GetBoolAnnotation(authReqAlwaysSetCookieAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil && ing_errors.IsValidationError(err) {
		return nil, fmt.Errorf("%s is invalid: %w", authReqAlwaysSetCookieAnnotation, err)
	}

	return &Config{
		URL:                    urlString,
		Host:                   authURL.Hostname(),
		SigninURL:              signIn,
		SigninURLRedirectParam: signInRedirectParam,
		Method:                 authMethod,
		ResponseHeaders:        responseHeaders,
		RequestRedirect:        requestRedirect,
		AuthSnippet:            authSnippet,
		AuthCacheKey:           authCacheKey,
		AuthCacheDuration:      authCacheDuration,
		KeepaliveConnections:   keepaliveConnections,
		KeepaliveShareVars:     keepaliveShareVars,
		KeepaliveRequests:      keepaliveRequests,
		KeepaliveTimeout:       keepaliveTimeout,
		ProxySetHeaders:        proxySetHeaders,
		AlwaysSetCookie:        alwaysSetCookie,
	}, nil
}

// ParseStringToCacheDurations parses and validates the provided string
// into a list of cache durations.
// It will always return at least one duration (the default duration)
func ParseStringToCacheDurations(input string) ([]string, error) {
	authCacheDuration := []string{}
	if input != "" {
		arr := strings.Split(input, ",")
		for _, duration := range arr {
			duration = strings.TrimSpace(duration)
			if duration != "" {
				if !ValidCacheDuration(duration) {
					authCacheDuration = []string{DefaultCacheDuration}
					return authCacheDuration, ing_errors.NewLocationDenied(fmt.Sprintf("invalid cache duration: %s", duration))
				}
				authCacheDuration = append(authCacheDuration, duration)
			}
		}
	}

	if len(authCacheDuration) == 0 {
		authCacheDuration = append(authCacheDuration, DefaultCacheDuration)
	}
	return authCacheDuration, nil
}

func (a authReq) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a authReq) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, authReqAnnotations.Annotations)
}
