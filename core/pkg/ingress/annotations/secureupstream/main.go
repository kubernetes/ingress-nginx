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

package secureupstream

import (
	"fmt"

	"github.com/pkg/errors"
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	secureUpstream       = "ingress.kubernetes.io/secure-backends"
	secureVerifyCASecret = "ingress.kubernetes.io/secure-verify-ca-secret"
)

// Secure describes SSL backend configuration
type Secure struct {
	Secure bool                 `json:"secure"`
	CACert resolver.AuthSSLCert `json:"caCert"`
}

type su struct {
	certResolver resolver.AuthCertificate
}

// NewParser creates a new secure upstream annotation parser
func NewParser(resolver resolver.AuthCertificate) parser.IngressAnnotation {
	return su{
		certResolver: resolver,
	}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the upstream servers should use SSL
func (a su) Parse(ing *extensions.Ingress) (interface{}, error) {
	s, _ := parser.GetBoolAnnotation(secureUpstream, ing)
	ca, _ := parser.GetStringAnnotation(secureVerifyCASecret, ing)
	secure := &Secure{
		Secure: s,
		CACert: resolver.AuthSSLCert{},
	}
	if !s && ca != "" {
		return secure,
			errors.Errorf("trying to use CA from secret %v/%v on a non secure backend", ing.Namespace, ca)
	}
	if ca == "" {
		return secure, nil
	}
	caCert, err := a.certResolver.GetAuthCertificate(fmt.Sprintf("%v/%v", ing.Namespace, ca))
	if err != nil {
		return secure, errors.Wrap(err, "error obtaining certificate")
	}
	if caCert == nil {
		return secure, nil
	}
	return &Secure{
		Secure: s,
		CACert: *caCert,
	}, nil
}
