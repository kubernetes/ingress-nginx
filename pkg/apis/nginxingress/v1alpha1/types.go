/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/pkg/ingress/resolver"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=nginxconfiguration

// Configuration is a specification for a NGINXConfiguration resource
type Configuration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigurationSpec `json:"spec"`
}

// ConfigurationSpec is the spec for a NGINXConfiguration resource
type ConfigurationSpec struct {
	Auth *Auth `json:"auth,omitempty"`
	// ConfigurationSnippet contains additional configuration for the backend
	// to be considered in the configuration of the location
	ConfigurationSnippet string `json:"configurationSnippet,omitempty"`
	DefaultBackend       string `json:"defaultBackend,omitempty"`
	// EnableCORS indicates if path must support CORS
	// +optional
	EnableCORS bool `json:"enableCors,omitempty"`
	// Proxy contains information about timeouts and buffer sizes
	// to be used in connections against endpoints
	// +optional
	Proxy *ProxyConfiguration `json:"proxy,omitempty"`
	// RateLimit describes a limit in the number of connections per IP
	// address or connections per second.
	// The Redirect annotation precedes RateLimit
	// +optional
	RateLimit *RateLimit `json:"rateLimit,omitempty"`
	// Redirect describes a temporal o permanent redirection this location.
	// +optional
	Redirect *Redirect `json:"redirect,omitempty"`
	// Rewrite describes the redirection this location.
	// +optional
	Rewrite *Rewrite `json:"rewrite,omitempty"`
	// ServerAlias return the alias of the server name
	// +optional
	ServerAlias    string `json:"serverAlias,omitempty"`
	ServerSnippet  string `json:"serverSnippet,omitempty"`
	SSLPassthrough bool   `json:"sslPassthrough,omitempty"`
	// UsePortInRedirects indicates if redirects must specify the port
	// +optional
	UsePortInRedirects bool `json:"usePortInRedirects,omitempty"`
	// VTSFilterKey contains the vts filter key on the location level
	// https://github.com/vozlt/nginx-module-vts#vhost_traffic_status_filter_by_set_key
	// +optional
	VTSFilterKey string `json:"vtsFilterKey,omitempty"`
	// Whitelist indicates only connections from certain client
	// addresses or networks are allowed.
	// +optional
	Whitelist *Whitelist `json:"whitelist,omitempty"`

	Upstream *Upstream `json:"upstream,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=foos

// ConfigurationList is a list of Configuration resources
type ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Configuration `json:"items"`
}

type Auth struct {
	Basic    BasicDigestAuth `json:"basic"`
	Digest   BasicDigestAuth `json:"digest"`
	Cert     CertAuth        `json:"cert"`
	External ExternalAuth    `json:"external"`
}

type CertAuth struct {
	resolver.AuthSSLCert
	VerifyClient    string `json:"verify_client"`
	ValidationDepth int    `json:"validationDepth"`
	ErrorPage       string `json:"errorPage"`
}

type BasicDigestAuth struct {
	Realm string `json:"realm"`
	File  string `json:"file"`
}

type ExternalAuth struct {
	URL string `json:"url"`
	// Host contains the hostname defined in the URL
	Host            string   `json:"host"`
	SigninURL       string   `json:"signinUrl"`
	Method          string   `json:"method"`
	SendBody        bool     `json:"sendBody"`
	ResponseHeaders []string `json:"responseHeaders,omitEmpty"`
}

// ProxyConfiguration returns the proxy timeout to use in the upstream server/s
type ProxyConfiguration struct {
	BodySize string `json:"bodySize"`
	// ClientBodyBufferSize allows for the configuration of the client body
	// buffer size for a specific location.
	ClientBodyBufferSize string `json:"clientBodyBufferSize"`
	ConnectTimeout       int    `json:"connectTimeout"`
	SendTimeout          int    `json:"sendTimeout"`
	ReadTimeout          int    `json:"readTimeout"`
	BufferSize           string `json:"bufferSize"`
	CookieDomain         string `json:"cookieDomain"`
	CookiePath           string `json:"cookiePath"`
	NextUpstream         string `json:"nextUpstream"`
	PassParams           string `json:"passParams"`
	RequestBuffering     string `json:"requestBuffering"`
	// UpstreamVirtualHost overwrites the Host header passed into the backend.
	// Defaults to virtual host of the incoming request.
	HostHeader string `json:"upstreamVirtualHost"`
}

// Upstream returns the URL and method to use check the status of
// the upstream server/s
type Upstream struct {
	UseServiceInUpstream bool
	// SecureUpstream describes SSL backend configuration
	SecureUpstream        bool        `json:"secure"`
	UpstreamCACertificate AuthSSLCert `json:"caCert"`
	HashBy                string      `json:"hashBy"`
	MaxFails              int         `json:"maxFails"`
	FailTimeout           int         `json:"failTimeout"`
	// SessionAffinity configures the nginx session affinity
	SessionAffinity *AffinityConfig `json:"sessionAfinity,omitempty"`
}

// AuthSSLCert contains the necessary information to do certificate based
// authentication of an ingress location
type AuthSSLCert struct {
	// Secret contains the name of the secret this was fetched from
	Secret string `json:"secret"`
	// CAFileName contains the path to the secrets 'ca.crt'
	CAFileName string `json:"caFilename"`
	// PemSHA contains the SHA1 hash of the 'ca.crt' or combinations of (tls.crt, tls.key, tls.crt) depending on certs in secret
	PemSHA string `json:"pemSha"`
}

type Whitelist struct {
	CIDR []string `json:"cidr,omitEmpty"`
}

type Redirect struct {
	URL       string `json:"url"`
	Code      int    `json:"code"`
	FromToWWW bool   `json:"fromToWWW"`
}

type Rewrite struct {
	// AddBaseURL indicates if is required to add a base tag in the head
	// of the responses from the upstream servers
	AddBaseURL bool `json:"addBaseUrl"`
	// AppRoot defines the Application Root that the Controller must redirect if it's not in '/' context
	AppRoot string `json:"appRoot"`
	// BaseURLScheme override for the scheme passed to the base tag
	BaseURLScheme string `json:"baseUrlScheme"`
	// ForceSSLRedirect indicates if the location section is accessible SSL only
	ForceSSLRedirect bool `json:"forceSSLRedirect"`
	// SSLRedirect indicates if the location section is accessible SSL only
	SSLRedirect bool `json:"sslRedirect"`
	// Target URI where the traffic must be redirected
	Target string `json:"target"`
}

// RateLimit returns rate limit configuration for an Ingress rule limiting the
// number of connections per IP address and/or connections per second.
// If you both annotations are specified in a single Ingress rule, RPS limits
// takes precedence
type RateLimit struct {
	// Connections indicates a limit with the number of connections per IP address
	Connections Zone `json:"connections"`
	// RPS indicates a limit with the number of connections per second
	RPS Zone `json:"rps"`
	// RPM indicates a limit with the number of connections per minute
	RPM Zone `json:"rpm"`

	LimitRate int `json:"limit-rate"`

	LimitRateAfter int `json:"limit-rate-after"`

	Name string `json:"name"`

	ID string `json:"id"`

	Whitelist Whitelist `json:"whitelist"`
}

// AffinityConfig describes the per ingress session affinity configuration
type AffinityConfig struct {
	// The type of affinity that will be used
	AffinityType string `json:"type"`
	CookieConfig
}

// CookieConfig describes the Config of cookie type affinity
type CookieConfig struct {
	// The name of the cookie that will be used in case of cookie affinity type.
	Name string `json:"name"`
	// The hash that will be used to encode the cookie in case of cookie affinity type
	Hash      string              `json:"hash"`
	Locations map[string][]string `json:"locations,omitempty"`
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
