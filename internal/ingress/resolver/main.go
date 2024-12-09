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

	// GetSecurityConfiguration returns the configuration options from Ingress
	GetSecurityConfiguration() defaults.SecurityConfiguration

	// GetConfigMap searches for configmap containing the namespace and name using the character /
	GetConfigMap(string) (*apiv1.ConfigMap, error)

	// GetSecret searches for secrets containing the namespace and name using the character /
	GetSecret(string) (*apiv1.Secret, error)

	// GetAuthCertificate resolves a given secret name into an SSL certificate and CRL.
	// The secret must contain 2 keys named:

	//   ca.crt: contains the certificate chain used for authentication
	//   ca.crl: contains the revocation list used for authentication
	GetAuthCertificate(string) (*AuthSSLCert, error)

	// GetSSLClientCert resolves a given secret name into an SSL certificate.
	GetSSLClientCert(string) (*SSLClientCert, error)

	// GetSSLCA resolves a given configMap name into an SSL CA.
	GetSSLCA(string) (*SSLCA, error)

	// GetService searches for services containing the namespace and name using the character /
	GetService(string) (*apiv1.Service, error)
}

// AuthSSLCert contains the necessary information to do certificate based
// authentication of an ingress location
type AuthSSLCert struct {
	// Secret contains the name of the secret this was fetched from
	Secret string `json:"secret"`
	// CAFileName contains the path to the secrets 'ca.crt'
	CAFileName string `json:"caFilename"`
	// CASHA contains the SHA1 hash of the 'ca.crt' or combinations of (tls.crt, tls.key, tls.crt) depending on certs in secret
	CASHA string `json:"caSha"`
	// CRLFileName contains the path to the secrets 'ca.crl'
	CRLFileName string `json:"crlFileName"`
	// CRLSHA contains the SHA1 hash of the 'ca.crl' file
	CRLSHA string `json:"crlSha"`
	// PemFileName contains the path to the secrets 'tls.crt' and 'tls.key'
	PemFileName string `json:"pemFilename"`
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
	if asslc1.CASHA != assl2.CASHA {
		return false
	}

	if asslc1.CRLFileName != assl2.CRLFileName {
		return false
	}
	if asslc1.CRLSHA != assl2.CRLSHA {
		return false
	}

	return true
}

// SSLClientCert contains the clients certificate information
type SSLClientCert struct {
	// Secret contains the name of the secret this was fetched from
	Secret string `json:"secret"`
	// PemFileName contains the path to the secrets 'tls.crt' and 'tls.key'
	PemFileName string `json:"pemFilename"`
}

// Equal tests for equality between two SSLClientCert types
func (sslcc1 *SSLClientCert) Equal(sslcc2 *SSLClientCert) bool {
	if sslcc1 == sslcc2 {
		return true
	}
	if sslcc1 == nil || sslcc2 == nil {
		return false
	}

	if sslcc1.Secret != sslcc2.Secret {
		return false
	}

	return true
}

// SSLCA contains the CAs used to validate client certificates
type SSLCA struct {
	// ConfigMap contains the name of the configMap this was fetched from
	ConfigMap string `json:"configmap"`
	// CAFileName contains the path to the secrets 'ca.crt'
	CAFileName string `json:"caFilename"`
	// CASHA contains the SHA1 hash of the 'ca.crt'
	CASHA string `json:"caSha"`
	// CRLFileName contains the path to the secrets 'ca.crl'
	CRLFileName string `json:"crlFileName"`
	// CRLSHA contains the SHA1 hash of the 'ca.crl' file
	CRLSHA string `json:"crlSha"`
}

// Equal tests for equality between two SSLCA types
func (sslc1 *SSLCA) Equal(sslc2 *SSLCA) bool {
	if sslc1 == sslc2 {
		return true
	}
	if sslc1 == nil || sslc2 == nil {
		return false
	}

	if sslc1.ConfigMap != sslc2.ConfigMap {
		return false
	}
	if sslc1.CAFileName != sslc2.CAFileName {
		return false
	}
	if sslc1.CASHA != sslc2.CASHA {
		return false
	}

	if sslc1.CRLFileName != sslc2.CRLFileName {
		return false
	}
	if sslc1.CRLSHA != sslc2.CRLSHA {
		return false
	}

	return true
}
