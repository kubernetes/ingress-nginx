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

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	annotationAffinityType = "affinity"
	// If a cookie with this name exists,
	// its value is used as an index into the list of available backends.
	annotationAffinityCookieName = "session-cookie-name"

	defaultAffinityCookieName = "INGRESSCOOKIE"

	// This is the algorithm used by nginx to generate a value for the session cookie, if
	// one isn't supplied and affinity is set to "cookie".
	annotationAffinityCookieHash = "session-cookie-hash"
	defaultAffinityCookieHash    = "md5"

	// This is used to control the cookie expires, its value is a number of seconds until the
	// cookie expires
	annotationAffinityCookieExpires = "session-cookie-expires"

	// This is used to control the cookie expires, its value is a number of seconds until the
	// cookie expires
	annotationAffinityCookieMaxAge = "session-cookie-max-age"

	// This is used to control the cookie path when use-regex is set to true
	annotationAffinityCookiePath = "session-cookie-path"
)

var (
	affinityCookieHashRegex    = regexp.MustCompile(`^(index|md5|sha1)$`)
	affinityCookieExpiresRegex = regexp.MustCompile(`(^0|-?[1-9]\d*$)`)
)

// Config describes the per ingress session affinity config
type Config struct {
	// The type of affinity that will be used
	Type string `json:"type"`
	Cookie
}

// Cookie describes the Config of cookie type affinity
type Cookie struct {
	// The name of the cookie that will be used in case of cookie affinity type.
	Name string `json:"name"`
	// The hash that will be used to encode the cookie in case of cookie affinity type
	Hash string `json:"hash"`
	// The time duration to control cookie expires
	Expires string `json:"expires"`
	// The number of seconds until the cookie expires
	MaxAge string `json:"maxage"`
	// The path that a cookie will be set on
	Path string `json:"path"`
}

// cookieAffinityParse gets the annotation values related to Cookie Affinity
// It also sets default values when no value or incorrect value is found
func (a affinity) cookieAffinityParse(ing *extensions.Ingress) *Cookie {
	var err error

	cookie := &Cookie{}

	cookie.Name, err = parser.GetStringAnnotation(annotationAffinityCookieName, ing)
	if err != nil {
		klog.V(3).Infof("Ingress %v: No value found in annotation %v. Using the default %v", ing.Name, annotationAffinityCookieName, defaultAffinityCookieName)
		cookie.Name = defaultAffinityCookieName
	}

	cookie.Hash, err = parser.GetStringAnnotation(annotationAffinityCookieHash, ing)
	if err != nil || !affinityCookieHashRegex.MatchString(cookie.Hash) {
		klog.V(3).Infof("Invalid or no annotation value found in Ingress %v: %v. Setting it to default %v", ing.Name, annotationAffinityCookieHash, defaultAffinityCookieHash)
		cookie.Hash = defaultAffinityCookieHash
	}

	cookie.Expires, err = parser.GetStringAnnotation(annotationAffinityCookieExpires, ing)
	if err != nil || !affinityCookieExpiresRegex.MatchString(cookie.Expires) {
		klog.V(3).Infof("Invalid or no annotation value found in Ingress %v: %v. Ignoring it", ing.Name, annotationAffinityCookieExpires)
		cookie.Expires = ""
	}

	cookie.MaxAge, err = parser.GetStringAnnotation(annotationAffinityCookieMaxAge, ing)
	if err != nil || !affinityCookieExpiresRegex.MatchString(cookie.MaxAge) {
		klog.V(3).Infof("Invalid or no annotation value found in Ingress %v: %v. Ignoring it", ing.Name, annotationAffinityCookieMaxAge)
		cookie.MaxAge = ""
	}

	cookie.Path, err = parser.GetStringAnnotation(annotationAffinityCookiePath, ing)
	if err != nil {
		klog.V(3).Infof("Invalid or no annotation value found in Ingress %v: %v. Ignoring it", ing.Name, annotationAffinityCookieMaxAge)
	}

	return cookie
}

// NewParser creates a new Affinity annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return affinity{r}
}

type affinity struct {
	r resolver.Resolver
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure the affinity directives
func (a affinity) Parse(ing *extensions.Ingress) (interface{}, error) {
	cookie := &Cookie{}
	// Check the type of affinity that will be used
	at, err := parser.GetStringAnnotation(annotationAffinityType, ing)
	if err != nil {
		at = ""
	}

	switch at {
	case "cookie":
		cookie = a.cookieAffinityParse(ing)
	default:
		klog.V(3).Infof("No default affinity was found for Ingress %v", ing.Name)

	}

	return &Config{
		Type:   at,
		Cookie: *cookie,
	}, nil
}
