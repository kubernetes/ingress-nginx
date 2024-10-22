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

package sessionaffinity

import (
	"regexp"

	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	annotationAffinityType           = "affinity"
	annotationAffinityMode           = "affinity-mode"
	annotationAffinityCanaryBehavior = "affinity-canary-behavior"

	// If a cookie with this name exists,
	// its value is used as an index into the list of available backends.
	annotationAffinityCookieName = "session-cookie-name"

	defaultAffinityCookieName = "INGRESSCOOKIE"

	// This is used to force the Secure flag on the cookie even if the
	// incoming request is not secured. (https://github.com/kubernetes/ingress-nginx/issues/6812)
	annotationAffinityCookieSecure = "session-cookie-secure"

	// This is used to control the cookie expires, its value is a number of seconds until the
	// cookie expires
	annotationAffinityCookieExpires = "session-cookie-expires"

	// This is used to control the cookie expires, its value is a number of seconds until the
	// cookie expires
	annotationAffinityCookieMaxAge = "session-cookie-max-age"

	// This is used to control the cookie path when use-regex is set to true
	annotationAffinityCookiePath = "session-cookie-path"

	// This is used to control the cookie Domain
	annotationAffinityCookieDomain = "session-cookie-domain"

	// This is used to control the SameSite attribute of the cookie
	annotationAffinityCookieSameSite = "session-cookie-samesite"

	// This is used to control whether SameSite=None should be conditionally applied based on the User-Agent
	annotationAffinityCookieConditionalSameSiteNone = "session-cookie-conditional-samesite-none"

	// This is used to set the Partitioned flag on the cookie
	annotationAffinityCookiePartitioned = "session-cookie-partitioned"

	// This is used to control the cookie change after request failure
	annotationAffinityCookieChangeOnFailure = "session-cookie-change-on-failure"

	cookieAffinity = "cookie"
)

var sessionAffinityAnnotations = parser.Annotation{
	Group: "affinity",
	Annotations: parser.AnnotationFields{
		annotationAffinityType: {
			Validator:     parser.ValidateOptions([]string{cookieAffinity}, true, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables and sets the affinity type in all Upstreams of an Ingress. This way, a request will always be directed to the same upstream server. The only affinity type available for NGINX is cookie`,
		},
		annotationAffinityMode: {
			Validator: parser.ValidateOptions([]string{"balanced", "persistent"}, true, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the stickiness of a session. 
			Setting this to balanced (default) will redistribute some sessions if a deployment gets scaled up, therefore rebalancing the load on the servers. 
			Setting this to persistent will not rebalance sessions to new servers, therefore providing maximum stickiness.`,
		},
		annotationAffinityCanaryBehavior: {
			Validator: parser.ValidateOptions([]string{"sticky", "legacy"}, true, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation defines the behavior of canaries when session affinity is enabled.
			Setting this to sticky (default) will ensure that users that were served by canaries, will continue to be served by canaries.
			Setting this to legacy will restore original canary behavior, when session affinity was ignored.`,
		},
		annotationAffinityCookieName: {
			Validator:     parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation allows to specify the name of the cookie that will be used to route the requests`,
		},
		annotationAffinityCookieSecure: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation set the cookie as secure regardless the protocol of the incoming request`,
		},
		annotationAffinityCookieExpires: {
			Validator:     parser.ValidateRegex(affinityCookieExpiresRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation is a legacy version of "session-cookie-max-age" for compatibility with older browsers, generates an "Expires" cookie directive by adding the seconds to the current date`,
		},
		annotationAffinityCookieMaxAge: {
			Validator:     parser.ValidateRegex(affinityCookieExpiresRegex, false),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation sets the time until the cookie expires`,
		},
		annotationAffinityCookiePath: {
			Validator:     parser.ValidateRegex(parser.URLIsValidRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the Path that will be set on the cookie (required if your Ingress paths use regular expressions)`,
		},
		annotationAffinityCookieDomain: {
			Validator:     parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation defines the Domain attribute of the sticky cookie.`,
		},
		annotationAffinityCookieSameSite: {
			Validator: parser.ValidateOptions([]string{"none", "lax", "strict"}, false, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation is used to apply a SameSite attribute to the sticky cookie. 
			Browser accepted values are None, Lax, and Strict`,
		},
		annotationAffinityCookieConditionalSameSiteNone: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation is used to omit SameSite=None from browsers with SameSite attribute incompatibilities`,
		},
		annotationAffinityCookiePartitioned: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation sets the cookie as Partitioned`,
		},
		annotationAffinityCookieChangeOnFailure: {
			Validator: parser.ValidateBool,
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation, when set to false will send request to upstream pointed by sticky cookie even if previous attempt failed. 
			When set to true and previous attempt failed, sticky cookie will be changed to point to another upstream.`,
		},
	},
}

var affinityCookieExpiresRegex = regexp.MustCompile(`(^0|-?[1-9]\d*$)`)

// Config describes the per ingress session affinity config
type Config struct {
	// The type of affinity that will be used
	Type string `json:"type"`
	// The affinity mode, i.e. how sticky a session is
	Mode string `json:"mode"`
	// Affinity behavior for canaries (sticky or legacy)
	CanaryBehavior string `json:"canaryBehavior"`
	Cookie
}

// Cookie describes the Config of cookie type affinity
type Cookie struct {
	// The name of the cookie that will be used in case of cookie affinity type.
	Name string `json:"name"`
	// The time duration to control cookie expires
	Expires string `json:"expires"`
	// The number of seconds until the cookie expires
	MaxAge string `json:"maxage"`
	// The path that a cookie will be set on
	Path string `json:"path"`
	// The domain that a cookie will be set on
	Domain string `json:"domain"`
	// Flag that allows cookie regeneration on request failure
	ChangeOnFailure bool `json:"changeonfailure"`
	// Secure flag to be set
	Secure bool `json:"secure"`
	// SameSite attribute value
	SameSite string `json:"samesite"`
	// Flag that conditionally applies SameSite=None attribute on cookie if user agent accepts it.
	ConditionalSameSiteNone bool `json:"conditional-samesite-none"`
	// Partitioned flag to be set
	Partitioned bool `json:"partitioned"`
}

type affinity struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// cookieAffinityParse gets the annotation values related to Cookie Affinity
// It also sets default values when no value or incorrect value is found
func (a affinity) cookieAffinityParse(ing *networking.Ingress) *Cookie {
	var err error

	cookie := &Cookie{}

	cookie.Name, err = parser.GetStringAnnotation(annotationAffinityCookieName, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieName, "default", defaultAffinityCookieName)
		cookie.Name = defaultAffinityCookieName
	}

	cookie.Expires, err = parser.GetStringAnnotation(annotationAffinityCookieExpires, ing, a.annotationConfig.Annotations)
	if err != nil || !affinityCookieExpiresRegex.MatchString(cookie.Expires) {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieExpires)
		cookie.Expires = ""
	}

	cookie.MaxAge, err = parser.GetStringAnnotation(annotationAffinityCookieMaxAge, ing, a.annotationConfig.Annotations)
	if err != nil || !affinityCookieExpiresRegex.MatchString(cookie.MaxAge) {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieMaxAge)
		cookie.MaxAge = ""
	}

	cookie.Path, err = parser.GetStringAnnotation(annotationAffinityCookiePath, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookiePath)
	}

	cookie.Domain, err = parser.GetStringAnnotation(annotationAffinityCookieDomain, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieDomain)
	}

	cookie.SameSite, err = parser.GetStringAnnotation(annotationAffinityCookieSameSite, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieSameSite)
	}

	cookie.Secure, err = parser.GetBoolAnnotation(annotationAffinityCookieSecure, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieSecure)
	}

	cookie.ConditionalSameSiteNone, err = parser.GetBoolAnnotation(annotationAffinityCookieConditionalSameSiteNone, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieConditionalSameSiteNone)
	}

	cookie.Partitioned, err = parser.GetBoolAnnotation(annotationAffinityCookiePartitioned, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookiePartitioned)
	}

	cookie.ChangeOnFailure, err = parser.GetBoolAnnotation(annotationAffinityCookieChangeOnFailure, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("Invalid or no annotation value found. Ignoring", "ingress", klog.KObj(ing), "annotation", annotationAffinityCookieChangeOnFailure)
	}

	return cookie
}

// NewParser creates a new Affinity annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return affinity{
		r:                r,
		annotationConfig: sessionAffinityAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure the affinity directives
func (a affinity) Parse(ing *networking.Ingress) (interface{}, error) {
	cookie := &Cookie{}
	// Check the type of affinity that will be used
	at, err := parser.GetStringAnnotation(annotationAffinityType, ing, a.annotationConfig.Annotations)
	if err != nil {
		at = ""
	}

	// Check the affinity mode that will be used
	am, err := parser.GetStringAnnotation(annotationAffinityMode, ing, a.annotationConfig.Annotations)
	if err != nil {
		am = ""
	}

	cb, err := parser.GetStringAnnotation(annotationAffinityCanaryBehavior, ing, a.annotationConfig.Annotations)
	if err != nil {
		cb = ""
	}

	switch at {
	case cookieAffinity:
		cookie = a.cookieAffinityParse(ing)
	default:
		klog.V(3).InfoS("No default affinity found", "ingress", ing.Name)
	}

	return &Config{
		Type:           at,
		Mode:           am,
		CanaryBehavior: cb,
		Cookie:         *cookie,
	}, nil
}

func (a affinity) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a affinity) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, sessionAffinityAnnotations.Annotations)
}
