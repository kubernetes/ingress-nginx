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

package ingress

// Equal tests for equality between two BackendInfo types
func (bi1 *BackendInfo) Equal(bi2 *BackendInfo) bool {
	if bi1 == bi2 {
		return true
	}
	if bi1 == nil || bi2 == nil {
		return false
	}
	if bi1.Name != bi2.Name {
		return false
	}
	if bi1.Release != bi2.Release {
		return false
	}
	if bi1.Build != bi2.Build {
		return false
	}
	if bi1.Repository != bi2.Repository {
		return false
	}

	return true
}

// Equal tests for equality between two Configuration types
func (c1 *Configuration) Equal(c2 *Configuration) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}

	if len(c1.Backends) != len(c2.Backends) {
		return false
	}

	for _, c1b := range c1.Backends {
		found := false
		for _, c2b := range c2.Backends {
			if c1b.Equal(c2b) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(c1.Servers) != len(c2.Servers) {
		return false
	}

	for _, c1s := range c1.Servers {
		found := false
		for _, c2s := range c2.Servers {
			if c1s.Equal(c2s) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(c1.TCPEndpoints) != len(c2.TCPEndpoints) {
		return false
	}

	for _, tcp1 := range c1.TCPEndpoints {
		found := false
		for _, tcp2 := range c2.TCPEndpoints {
			if (&tcp1).Equal(&tcp2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(c1.UDPEndpoints) != len(c2.UDPEndpoints) {
		return false
	}

	for _, udp1 := range c1.UDPEndpoints {
		found := false
		for _, udp2 := range c2.UDPEndpoints {
			if (&udp1).Equal(&udp2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(c1.PassthroughBackends) != len(c2.PassthroughBackends) {
		return false
	}

	for _, ptb1 := range c1.PassthroughBackends {
		found := false
		for _, ptb2 := range c2.PassthroughBackends {
			if ptb1.Equal(ptb2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Equal tests for equality between two Backend types
func (b1 *Backend) Equal(b2 *Backend) bool {
	if b1 == b2 {
		return true
	}
	if b1 == nil || b2 == nil {
		return false
	}
	if b1.Name != b2.Name {
		return false
	}

	if b1.Service == nil || b2.Service == nil {
		return false
	}
	if b1.Service.GetNamespace() != b2.Service.GetNamespace() {
		return false
	}
	if b1.Service.GetName() != b2.Service.GetName() {
		return false
	}
	if b1.Service.GetResourceVersion() != b2.Service.GetResourceVersion() {
		return false
	}

	if b1.Port != b2.Port {
		return false
	}
	if b1.Secure != b2.Secure {
		return false
	}
	if !(&b1.SecureCACert).Equal(&b2.SecureCACert) {
		return false
	}
	if b1.SSLPassthrough != b2.SSLPassthrough {
		return false
	}
	if !(&b1.SessionAffinity).Equal(&b2.SessionAffinity) {
		return false
	}

	if len(b1.Endpoints) != len(b2.Endpoints) {
		return false
	}

	for _, udp1 := range b1.Endpoints {
		found := false
		for _, udp2 := range b2.Endpoints {
			if (&udp1).Equal(&udp2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Equal tests for equality between two SessionAffinityConfig types
func (sac1 *SessionAffinityConfig) Equal(sac2 *SessionAffinityConfig) bool {
	if sac1 == sac2 {
		return true
	}
	if sac1 == nil || sac2 == nil {
		return false
	}
	if sac1.AffinityType != sac2.AffinityType {
		return false
	}
	if !(&sac1.CookieSessionAffinity).Equal(&sac2.CookieSessionAffinity) {
		return false
	}

	return true
}

// Equal tests for equality between two CookieSessionAffinity types
func (csa1 *CookieSessionAffinity) Equal(csa2 *CookieSessionAffinity) bool {
	if csa1 == csa2 {
		return true
	}
	if csa1 == nil || csa2 == nil {
		return false
	}
	if csa1.Name != csa2.Name {
		return false
	}
	if csa1.Hash != csa2.Hash {
		return false
	}

	return true
}

// Equal checks the equality against an Endpoint
func (e1 *Endpoint) Equal(e2 *Endpoint) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.Address != e2.Address {
		return false
	}
	if e1.Port != e2.Port {
		return false
	}
	if e1.MaxFails != e2.MaxFails {
		return false
	}
	if e1.FailTimeout != e2.FailTimeout {
		return false
	}

	if e1.Target == nil || e2.Target == nil {
		return false
	}
	if e1.Target.UID != e2.Target.UID {
		return false
	}
	if e1.Target.ResourceVersion != e2.Target.ResourceVersion {
		return false
	}

	return true
}

// Equal tests for equality between two Server types
func (s1 *Server) Equal(s2 *Server) bool {
	if s1 == s2 {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	if s1.Hostname != s2.Hostname {
		return false
	}
	if s1.SSLPassthrough != s2.SSLPassthrough {
		return false
	}
	if s1.SSLCertificate != s2.SSLCertificate {
		return false
	}
	if s1.SSLPemChecksum != s2.SSLPemChecksum {
		return false
	}

	if len(s1.Locations) != len(s2.Locations) {
		return false
	}

	for _, s1l := range s1.Locations {
		found := false
		for _, sl2 := range s2.Locations {
			if s1l.Equal(sl2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Equal tests for equality between two Location types
func (l1 *Location) Equal(l2 *Location) bool {
	if l1 == l2 {
		return true
	}
	if l1 == nil || l2 == nil {
		return false
	}
	if l1.Path != l2.Path {
		return false
	}
	if l1.IsDefBackend != l2.IsDefBackend {
		return false
	}
	if l1.Backend != l2.Backend {
		return false
	}

	if l1.Service == nil || l2.Service == nil {
		return false
	}
	if l1.Service.GetNamespace() != l2.Service.GetNamespace() {
		return false
	}
	if l1.Service.GetName() != l2.Service.GetName() {
		return false
	}
	if l1.Service.GetResourceVersion() != l2.Service.GetResourceVersion() {
		return false
	}

	if l1.Port.StrVal != l2.Port.StrVal {
		return false
	}
	if !(&l1.BasicDigestAuth).Equal(&l2.BasicDigestAuth) {
		return false
	}
	if l1.Denied != l2.Denied {
		return false
	}
	if l1.EnableCORS != l2.EnableCORS {
		return false
	}
	if !(&l1.ExternalAuth).Equal(&l2.ExternalAuth) {
		return false
	}
	if !(&l1.RateLimit).Equal(&l2.RateLimit) {
		return false
	}
	if !(&l1.Redirect).Equal(&l2.Redirect) {
		return false
	}
	if !(&l1.Rewrite).Equal(&l2.Rewrite) {
		return false
	}
	if !(&l1.Whitelist).Equal(&l2.Whitelist) {
		return false
	}
	if !(&l1.Proxy).Equal(&l2.Proxy) {
		return false
	}
	if !(&l1.CertificateAuth).Equal(&l2.CertificateAuth) {
		return false
	}
	if l1.UsePortInRedirects != l2.UsePortInRedirects {
		return false
	}
	if l1.ConfigurationSnippet != l2.ConfigurationSnippet {
		return false
	}
	if l1.ClientBodyBufferSize != l2.ClientBodyBufferSize {
		return false
	}

	return true
}

// Equal tests for equality between two SSLPassthroughBackend types
func (ptb1 *SSLPassthroughBackend) Equal(ptb2 *SSLPassthroughBackend) bool {
	if ptb1 == ptb2 {
		return true
	}
	if ptb1 == nil || ptb2 == nil {
		return false
	}
	if ptb1.Backend != ptb2.Backend {
		return false
	}
	if ptb1.Hostname != ptb2.Hostname {
		return false
	}
	if ptb1.Port != ptb2.Port {
		return false
	}

	if ptb1.Service == nil || ptb2.Service == nil {
		return false
	}
	if ptb1.Service.GetNamespace() != ptb2.Service.GetNamespace() {
		return false
	}
	if ptb1.Service.GetName() != ptb2.Service.GetName() {
		return false
	}
	if ptb1.Service.GetResourceVersion() != ptb2.Service.GetResourceVersion() {
		return false
	}

	return true
}

// Equal tests for equality between two L4Service types
func (e1 *L4Service) Equal(e2 *L4Service) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.Port != e2.Port {
		return false
	}
	if !(&e1.Backend).Equal(&e2.Backend) {
		return false
	}
	if len(e1.Endpoints) != len(e2.Endpoints) {
		return false
	}

	for _, ep1 := range e1.Endpoints {
		found := false
		for _, ep2 := range e2.Endpoints {
			if (&ep1).Equal(&ep2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Equal tests for equality between two L4Backend types
func (l4b1 *L4Backend) Equal(l4b2 *L4Backend) bool {
	if l4b1 == l4b2 {
		return true
	}
	if l4b1 == nil || l4b2 == nil {
		return false
	}
	if l4b1.Port != l4b2.Port {
		return false
	}
	if l4b1.Name != l4b2.Name {
		return false
	}
	if l4b1.Namespace != l4b2.Namespace {
		return false
	}
	if l4b1.Protocol != l4b2.Protocol {
		return false
	}

	return true
}
