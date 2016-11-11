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

package ingress

import (
	"os/exec"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"

	"k8s.io/ingress/core/pkg/ingress/annotations/auth"
	"k8s.io/ingress/core/pkg/ingress/annotations/authreq"
	"k8s.io/ingress/core/pkg/ingress/annotations/authtls"
	"k8s.io/ingress/core/pkg/ingress/annotations/ipwhitelist"
	"k8s.io/ingress/core/pkg/ingress/annotations/proxy"
	"k8s.io/ingress/core/pkg/ingress/annotations/ratelimit"
	"k8s.io/ingress/core/pkg/ingress/annotations/rewrite"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

var (
	// DefaultSSLDirectory defines the location where the SSL certificates will be generated
	DefaultSSLDirectory = "/ingress-controller/ssl"
)

// Controller ...
type Controller interface {
	// Start returns the command is executed to start the backend.
	// The command must run in foreground.
	Start()
	// Stop stops the backend
	Stop() error
	// Restart reload the backend with the a configuration file returning
	// the combined output of Stdout and Stderr
	Restart(data []byte) ([]byte, error)
	// Tests returns a commands that checks if the configuration file is valid
	// Example: nginx -t -c <file>
	Test(file string) *exec.Cmd
	// OnUpdate callback invoked from the sync queue https://k8s.io/ingress/core/blob/master/pkg/ingress/controller/controller.go#L355
	// when an update occurs. This is executed frequently because Ingress
	// controllers watches changes in:
	// - Ingresses: main work
	// - Secrets: referenced from Ingress rules with TLS configured
	// - ConfigMaps: where the controller reads custom configuration
	// - Services: referenced from Ingress rules and required to obtain
	//	 information about ports and annotations
	// - Endpoints: referenced from Services and what the backend uses
	//	 to route traffic
	//
	// ConfigMap content of --configmap
	// Configuration returns the translation from Ingress rules containing
	// information about all the upstreams (service endpoints ) "virtual"
	// servers (FQDN)
	// and all the locations inside each server. Each location contains
	// information about all the annotations were configured
	// https://k8s.io/ingress/core/blob/master/pkg/ingress/types.go#L48
	OnUpdate(*api.ConfigMap, Configuration) ([]byte, error)
	// UpstreamDefaults returns the minimum settings required to configure the
	// communication to upstream servers (endpoints)
	UpstreamDefaults() defaults.Backend
	// IsReloadRequired checks if the backend must be reloaded or not.
	// The parameter contains the new rendered template
	IsReloadRequired([]byte) bool
	// Info returns information about the ingress controller
	// This can include build version, repository, etc.
	Info() string
}

// Configuration describes
type Configuration struct {
	HealthzURL           string
	Upstreams            []*Upstream
	Servers              []*Server
	TCPUpstreams         []*Location
	UDPUpstreams         []*Location
	PassthroughUpstreams []*SSLPassthroughUpstreams
}

// Upstream describes an upstream server (endpoint)
type Upstream struct {
	// Secure indicates if the communication with the en
	Secure bool
	// Name represents an unique api.Service name formatted
	// as <namespace>-<name>-<port>
	Name string
	// Backends
	Backends []UpstreamServer
}

// SSLPassthroughUpstreams describes an SSL upstream server configured
// as passthrough (no TLS termination)
type SSLPassthroughUpstreams struct {
	Upstream

	Host string
}

// UpstreamByNameServers sorts upstreams by name
type UpstreamByNameServers []*Upstream

func (c UpstreamByNameServers) Len() int      { return len(c) }
func (c UpstreamByNameServers) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c UpstreamByNameServers) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

// UpstreamServer describes a server in an upstream
type UpstreamServer struct {
	// Address IP address of the endpoint
	Address string
	Port    string
	// MaxFails returns the maximum number of check failures
	// allowed before this should be considered dow.
	// Setting 0 indicates that the check is performed by a Kubernetes probe
	MaxFails    int
	FailTimeout int
}

// Server describes a virtual server
type Server struct {
	Name           string
	SSL            bool
	SSLPassthrough bool
	SSLCertificate string
	//SSLCertificateKey string
	SSLPemChecksum string
	Locations      []*Location
}

// Location describes a server location
type Location struct {
	IsDefBackend    bool
	SecureUpstream  bool
	EnableCORS      bool
	Path            string
	Upstream        Upstream
	BasicDigestAuth auth.BasicDigest
	RateLimit       ratelimit.RateLimit
	Redirect        rewrite.Redirect
	Whitelist       ipwhitelist.SourceRange
	ExternalAuth    authreq.External
	Proxy           proxy.Configuration
	CertificateAuth authtls.SSLCert
}

// UpstreamServerByAddrPort sorts upstream servers by address and port
type UpstreamServerByAddrPort []UpstreamServer

func (c UpstreamServerByAddrPort) Len() int      { return len(c) }
func (c UpstreamServerByAddrPort) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c UpstreamServerByAddrPort) Less(i, j int) bool {
	iName := c[i].Address
	jName := c[j].Address
	if iName != jName {
		return iName < jName
	}

	iU := c[i].Port
	jU := c[j].Port
	return iU < jU
}

// ServerByName sorts server by name
type ServerByName []*Server

func (c ServerByName) Len() int      { return len(c) }
func (c ServerByName) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c ServerByName) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

// LocationByPath sorts location by path
// Location / is the last one
type LocationByPath []*Location

func (c LocationByPath) Len() int      { return len(c) }
func (c LocationByPath) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c LocationByPath) Less(i, j int) bool {
	return c[i].Path > c[j].Path
}

// SSLCert describes a SSL certificate to be used in a server
type SSLCert struct {
	api.ObjectMeta

	//CertFileName string
	//KeyFileName  string
	CAFileName string

	// PemFileName contains the path to the file with the certificate and key concatenated
	PemFileName string
	// PemSHA contains the sha1 of the pem file.
	// This is used to detect changes in the secret that contains the certificates
	PemSHA string
	// CN contains all the common names defined in the SSL certificate
	CN []string
}

// GetObjectKind implements the ObjectKind interface as a noop
func (s SSLCert) GetObjectKind() unversioned.ObjectKind { return unversioned.EmptyObjectKind }
