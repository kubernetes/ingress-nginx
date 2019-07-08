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
	networking "k8s.io/api/networking/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config describes SSL backend configuration
type Config struct {
	CACert resolver.AuthSSLCert `json:"caCert"`
}

type su struct {
	r resolver.Resolver
}

// NewParser creates a new secure upstream annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return su{r}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the upstream servers should use SSL
func (a su) Parse(ing *networking.Ingress) (interface{}, error) {
	bp, _ := parser.GetStringAnnotation("backend-protocol", ing)
	ca, _ := parser.GetStringAnnotation("secure-verify-ca-secret", ing)
	secure := &Config{
		CACert: resolver.AuthSSLCert{},
	}

	if (bp != "HTTPS" && bp != "GRPCS") && ca != "" {
		return secure,
			errors.Errorf("trying to use CA from secret %v/%v on a non secure backend", ing.Namespace, ca)
	}
	if ca == "" {
		return secure, nil
	}
	caCert, err := a.r.GetAuthCertificate(fmt.Sprintf("%v/%v", ing.Namespace, ca))
	if err != nil {
		return secure, errors.Wrap(err, "error obtaining certificate")
	}
	if caCert == nil {
		return secure, nil
	}
	return &Config{
		CACert: *caCert,
	}, nil
}
