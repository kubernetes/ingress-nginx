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

package ingress

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/healthz"

	"k8s.io/ingress/core/pkg/ingress/annotations/auth"
	"k8s.io/ingress/core/pkg/ingress/annotations/authreq"
	"k8s.io/ingress/core/pkg/ingress/annotations/ipwhitelist"
	"k8s.io/ingress/core/pkg/ingress/annotations/proxy"
	"k8s.io/ingress/core/pkg/ingress/annotations/ratelimit"
	"k8s.io/ingress/core/pkg/ingress/annotations/rewrite"
	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

var (
	// DefaultSSLDirectory defines the location where the SSL certificates will be generated
	// This directory contains all the SSL certificates that are specified in Ingress rules.
	// The name of each file is <namespace>-<secret name>.pem. The content is the concatenated
	// certificate and key.
	DefaultSSLDirectory = "/ingress-controller/ssl"
)

// Controller holds the methods to handle an Ingress backend
// TODO (#18): Make sure this is sufficiently supportive of other backends.
type Controller interface {
	// HealthzChecker returns is a named healthz check that returns the ingress
	// controller status
	healthz.HealthzChecker

	// Reload takes a byte array representing the new loadbalancer configuration,
	// and returns a byte array containing any output/errors from the backend and
	// if a reload was required.
	// Before returning the backend must load the configuration in the given array
	// into the loadbalancer and restart it, or fail with an error and message string.
	// If reloading fails, there should be not change in the running configuration or
	// the given byte array.
	Reload(data []byte) ([]byte, bool, error)
	// OnUpdate callback invoked from the sync queue https://k8s.io/ingress/core/blob/master/pkg/ingress/controller/controller.go#L387
	// when an update occurs. This is executed frequently because Ingress
	// controllers watches changes in:
	// - Ingresses: main work
	// - Secrets: referenced from Ingress rules with TLS configured
	// - ConfigMaps: where the controller reads custom configuration
	// - Services: referenced from Ingress rules and required to obtain
	//	 information about ports and annotations
	// - Endpoints: referenced from Services and what the backend uses
	//	 to route traffic
	// Any update to services, endpoints, secrets (only those referenced from Ingress)
	// and ingress trigger the execution.
	// Notifications of type Add, Update and Delete:
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/client/cache/controller.go#L164
	//
	// ConfigMap content of --configmap
	// Configuration returns the translation from Ingress rules containing
	// information about all the upstreams (service endpoints ) "virtual"
	// servers (FQDN) and all the locations inside each server. Each
	// location contains information about all the annotations were configured
	// https://k8s.io/ingress/core/blob/master/pkg/ingress/types.go#L83
	// The backend returns the contents of the configuration file or an error
	// with the reason why was not possible to generate the file.
	//
	// The returned configuration is then passed to test, and then to reload
	// if there is no errors.
	OnUpdate(*api.ConfigMap, Configuration) ([]byte, error)
	// BackendDefaults returns the minimum settings required to configure the
	// communication to endpoints
	BackendDefaults() defaults.Backend
	// Info returns information about the ingress controller
	Info() *BackendInfo
}

// BackendInfo returns information about the backend.
// This fields contains information that helps to track issues or to
// map the running ingress controller to source code
type BackendInfo struct {
	// Name returns the name of the backend implementation
	Name string `json:"name"`
	// Release returns the running version (semver)
	Release string `json:"release"`
	// Build returns information about the git commit
	Build string `json:"build"`
	// Repository return information about the git repository
	Repository string `json:"repository"`
}

// Configuration holds the definition of all the parts required to describe all
// ingresses reachable by the ingress controller (using a filter by namespace)
type Configuration struct {
	// Backends are a list of backends used by all the Ingress rules in the
	// ingress controller. This list includes the default backend
	Backends []*Backend `json:"namespace"`
	// Servers
	Servers []*Server `json:"servers"`
	// TCPEndpoints contain endpoints for tcp streams handled by this backend
	// +optional
	TCPEndpoints []*Location `json:"tcpEndpoints,omitempty"`
	// UPDEndpoints contain endpoints for udp streams handled by this backend
	// +optional
	UPDEndpoints []*Location `json:"udpEndpoints,omitempty"`
	// PassthroughBackend contains the backends used for SSL passthrough.
	// It contains information about the associated Server Name Indication (SNI).
	// +optional
	PassthroughBackends []*SSLPassthroughBackend `json:"passthroughBackends,omitempty"`
}

// Backend describes one or more remote server/s (endpoints) associated with a service
type Backend struct {
	// Name represents an unique api.Service name formatted as <namespace>-<name>-<port>
	Name string `json:"name"`
	// This indicates if the communication protocol between the backend and the endpoint is HTTP or HTTPS
	// Allowing the use of HTTPS
	// The endpoint/s must provide a TLS connection.
	// The certificate used in the endpoint cannot be a self signed certificate
	// TODO: add annotation to allow the load of ca certificate
	Secure bool `json:"secure"`
	// Endpoints contains the list of endpoints currently running
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint describes a kubernetes endpoint in an backend
type Endpoint struct {
	// Address IP address of the endpoint
	Address string `json:"address"`
	// Port number of the TCP port
	Port string `json:"port"`
	// MaxFails returns the number of unsuccessful attempts to communicate
	// allowed before this should be considered dow.
	// Setting 0 indicates that the check is performed by a Kubernetes probe
	MaxFails int `json:"maxFails"`
	// FailTimeout returns the time in seconds during which the specified number
	// of unsuccessful attempts to communicate with the server should happen
	// to consider the endpoint unavailable
	FailTimeout int `json:"failTimeout"`
}

// Server describes a website
type Server struct {
	// Hostname returns the FQDN of the server
	Hostname string `json:"hostname"`
	// SSLPassthrough indicates if the TLS termination is realized in
	// the server or in the remote endpoint
	SSLPassthrough bool `json:"sslPassthrough"`
	// SSLCertificate path to the SSL certificate on disk
	SSLCertificate string `json:"sslCertificate"`
	// SSLPemChecksum returns the checksum of the certificate file on disk.
	// There is no restriction in the hash generator. This checksim can be
	// used to  determine if the secret changed without the use of file
	// system notifications
	SSLPemChecksum string `json:"sslPemChecksum"`
	// Locations list of URIs configured in the server.
	Locations []*Location `json:"locations,omitempty"`
}

// Location describes an URI inside a server.
// Also contains additional information about annotations in the Ingress.
//
// Important:
// The implementation of annotations is optional
//
// In some cases when more than one annotations is defined a particular order in the execution
// is required.
// The chain in the execution order of annotations should be:
// - CertificateAuth
// - Whitelist
// - RateLimit
// - BasicDigestAuth
// - ExternalAuth
// - Redirect
type Location struct {
	// Path is an extended POSIX regex as defined by IEEE Std 1003.1,
	// (i.e this follows the egrep/unix syntax, not the perl syntax)
	// matched against the path of an incoming request. Currently it can
	// contain characters disallowed from the conventional "path"
	// part of a URL as defined by RFC 3986. Paths must begin with
	// a '/'. If unspecified, the path defaults to a catch all sending
	// traffic to the backend.
	Path string `json:"path"`
	// IsDefBackend indicates if service specified in the Ingress
	// contains active endpoints or not. Returning true means the location
	// uses the default backend.
	IsDefBackend bool `json:"isDefBackend"`
	// Backend describes the name of the backend to use.
	Backend string `json:"backend"`
	// BasicDigestAuth returns authentication configuration for
	// an Ingress rule.
	// +optional
	BasicDigestAuth auth.BasicDigest `json:"basicDigestAuth,omitempty"`
	// Denied returns an error when this location cannot not be allowed
	// Requesting a denied location should return HTTP code 403.
	Denied error
	// EnableCORS indicates if path must support CORS
	// +optional
	EnableCORS bool `json:"enableCors,omitempty"`
	// ExternalAuth indicates the access to this location requires
	// authentication using an external provider
	// +optional
	ExternalAuth authreq.External `json:"externalAuth,omitempty"`
	// RateLimit describes a limit in the number of connections per IP
	// address or connections per second.
	// The Redirect annotation precedes RateLimit
	// +optional
	RateLimit ratelimit.RateLimit `json:"rateLimit,omitempty"`
	// Redirect describes the redirection this location.
	// +optional
	Redirect rewrite.Redirect `json:"redirect,omitempty"`
	// Whitelist indicates only connections from certain client
	// addresses or networks are allowed.
	// +optional
	Whitelist ipwhitelist.SourceRange `json:"whitelist,omitempty"`
	// Proxy contains information about timeouts and buffer sizes
	// to be used in connections against endpoints
	// +optional
	Proxy proxy.Configuration `json:"proxy,omitempty"`
	// CertificateAuth indicates the access to this location requires
	// external authentication
	// +optional
	CertificateAuth resolver.AuthSSLCert `json:"certificateAuth,omitempty"`
}

// SSLPassthroughBackend describes a SSL upstream server configured
// as passthrough (no TLS termination in the ingress controller)
// The endpoints must provide the TLS termination exposing the required SSL certificate.
// The ingress controller only pipes the underlying TCP connection
type SSLPassthroughBackend struct {
	// Backend describes the endpoints to use.
	Backend string `json:"namespace,omitempty"`
	// Hostname returns the FQDN of the server
	Hostname string `json:"hostname"`
}
