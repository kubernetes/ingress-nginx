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

package sslcipher

import (
	"regexp"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	sslPreferServerCipherAnnotation = "ssl-prefer-server-ciphers"
	sslCipherAnnotation             = "ssl-ciphers"
)

// Should cover something like "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP"
var regexValidSSLCipher = regexp.MustCompile(`^[A-Za-z0-9!:+\-]*$`)

var sslCipherAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		sslPreferServerCipherAnnotation: {
			Validator: parser.ValidateBool,
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `The following annotation will set the ssl_prefer_server_ciphers directive at the server level. 
			This configuration specifies that server ciphers should be preferred over client ciphers when using the SSLv3 and TLS protocols.`,
		},
		sslCipherAnnotation: {
			Validator:     parser.ValidateRegex(regexValidSSLCipher, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `Using this annotation will set the ssl_ciphers directive at the server level. This configuration is active for all the paths in the host.`,
		},
	},
}

type sslCipher struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config contains the ssl-ciphers & ssl-prefer-server-ciphers configuration
type Config struct {
	SSLCiphers             string
	SSLPreferServerCiphers string
}

// NewParser creates a new sslCipher annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return sslCipher{
		r:                r,
		annotationConfig: sslCipherAnnotations,
	}
}

// Parse parses the annotations contained in the ingress rule
// used to add ssl-ciphers & ssl-prefer-server-ciphers to the server name
func (sc sslCipher) Parse(ing *networking.Ingress) (interface{}, error) {
	config := &Config{}
	var err error
	var sslPreferServerCiphers bool

	sslPreferServerCiphers, err = parser.GetBoolAnnotation(sslPreferServerCipherAnnotation, ing, sc.annotationConfig.Annotations)
	if err != nil {
		config.SSLPreferServerCiphers = ""
	} else {
		if sslPreferServerCiphers {
			config.SSLPreferServerCiphers = "on"
		} else {
			config.SSLPreferServerCiphers = "off"
		}
	}

	config.SSLCiphers, err = parser.GetStringAnnotation(sslCipherAnnotation, ing, sc.annotationConfig.Annotations)
	if err != nil && !errors.IsInvalidContent(err) && !errors.IsMissingAnnotations(err) {
		return config, err
	}

	return config, nil
}

func (sc sslCipher) GetDocumentation() parser.AnnotationFields {
	return sc.annotationConfig.Annotations
}

func (sc sslCipher) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(sc.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, sslCipherAnnotations.Annotations)
}
