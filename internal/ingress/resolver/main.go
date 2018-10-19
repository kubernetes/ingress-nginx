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

package resolver

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
)

// Resolver is an interface that knows how to extract information from a controller
type Resolver interface {
	// GetDefaultBackend returns the backend that must be used as default
	GetDefaultBackend() defaults.Backend

	// GetSecret searches for secrets containing the namespace and name using a the character /
	GetSecret(string) (*apiv1.Secret, error)

	// GetAuthCertificate resolves a given secret name into an SSL certificate.
	// The secret must contain 3 keys named:

	//   ca.crt: contains the certificate chain used for authentication
	//   tls.crt: contains the server certificate
	//   tls.key: contains the server key
	GetAuthCertificate(string) (*AuthSSLCert, error)

	// GetService searches for services containing the namespace and name using a the character /
	GetService(string) (*apiv1.Service, error)
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

// Equal tests for equality between two AuthSSLCert types
func (asslc1 *AuthSSLCert) Equal(assl2 *AuthSSLCert) bool {
	if asslc1 == assl2 {
		return true
	}
	if asslc1 == nil || assl2 == nil {
		return false
	}

	if asslc1.Secret != assl2.Secret {
		return false
	}
	if asslc1.CAFileName != assl2.CAFileName {
		return false
	}
	if asslc1.PemSHA != assl2.PemSHA {
		return false
	}

	return true
}
