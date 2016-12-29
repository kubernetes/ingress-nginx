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
	"k8s.io/kubernetes/pkg/api"

	"k8s.io/ingress/core/pkg/ingress/annotations/authtls"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

// DefaultBackend has a method that returns the backend
// that must be used as default
type DefaultBackend interface {
	GetDefaultBackend() defaults.Backend
}

// Secret has a method that searchs for secrets contenating
// the namespace and name using a the character /
type Secret interface {
	GetSecret(string) (*api.Secret, error)
}

// AuthCertificate has a method that searchs for a secret
// that contains a SSL certificate.
// The secret must contain 3 keys named:
type AuthCertificate interface {
	GetAuthCertificate(string) (*authtls.SSLCert, error)
}
