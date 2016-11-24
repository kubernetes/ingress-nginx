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

package authtls

import (
	"fmt"

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/k8s"
)

const (
	// name of the secret
	authTLSSecret = "ingress.kubernetes.io/auth-tls-secret"
)

// SSLCert returns external authentication configuration for an Ingress rule
type SSLCert struct {
	Secret       string `json:"secret"`
	CertFileName string `json:"certFilename"`
	KeyFileName  string `json:"keyFilename"`
	CAFileName   string `json:"caFilename"`
	PemSHA       string `json:"pemSha"`
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to use an external URL as source for authentication
func ParseAnnotations(ing *extensions.Ingress,
	fn func(secret string) (*SSLCert, error)) (*SSLCert, error) {
	if ing.GetAnnotations() == nil {
		return &SSLCert{}, parser.ErrMissingAnnotations
	}

	str, err := parser.GetStringAnnotation(authTLSSecret, ing)
	if err != nil {
		return &SSLCert{}, err
	}

	if str == "" {
		return &SSLCert{}, fmt.Errorf("an empty string is not a valid secret name")
	}

	_, _, err = k8s.ParseNameNS(str)
	if err != nil {
		return &SSLCert{}, err
	}

	return fn(str)
}
