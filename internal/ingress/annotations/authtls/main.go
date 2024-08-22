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
	"regexp"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
)

const (
	defaultAuthTLSDepth     = 1
	defaultAuthVerifyClient = "on"

	annotationAuthTLSSecret             = "auth-tls-secret" //#nosec G101
	annotationAuthTLSVerifyClient       = "auth-tls-verify-client"
	annotationAuthTLSVerifyDepth        = "auth-tls-verify-depth"
	annotationAuthTLSErrorPage          = "auth-tls-error-page"
	annotationAuthTLSPassCertToUpstream = "auth-tls-pass-certificate-to-upstream" //#nosec G101
	annotationAuthTLSMatchCN            = "auth-tls-match-cn"
)

var (
	authVerifyClientRegex = regexp.MustCompile(`^(on|off|optional|optional_no_ca)$`)
	redirectRegex         = regexp.MustCompile(`^((https?://)?[A-Za-z0-9\-.]*(:\d+)?/[A-Za-z0-9\-.]*)?$`)
)

var authTLSAnnotations = parser.Annotation{
	Group: "authentication",
	Annotations: parser.AnnotationFields{
		annotationAuthTLSSecret: {
			Validator:     parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium, // Medium as it allows a subset of chars
			Documentation: `This annotation defines the secret that contains the certificate chain of allowed certs`,
		},
		annotationAuthTLSVerifyClient: {
			Validator:     parser.ValidateRegex(authVerifyClientRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium, // Medium as it allows a subset of chars
			Documentation: `This annotation enables verification of client certificates. Can be "on", "off", "optional" or "optional_no_ca"`,
		},
		annotationAuthTLSVerifyDepth: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines validation depth between the provided client certificate and the Certification Authority chain.`,
		},
		annotationAuthTLSErrorPage: {
			Validator:     parser.ValidateRegex(redirectRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation defines the URL/Page that user should be redirected in case of a Certificate Authentication Error`,
		},
		annotationAuthTLSPassCertToUpstream: {
			Validator:     parser.ValidateBool,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines if the received certificates should be passed or not to the upstream server in the header "ssl-client-cert"`,
		},
		annotationAuthTLSMatchCN: {
			Validator:     parser.CommonNameAnnotationValidator,
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation adds a sanity check for the CN of the client certificate that is sent over using a string / regex starting with "CN="`,
		},
	},
}

// Config contains the AuthSSLCert used for mutual authentication
// and the configured ValidationDepth
type Config struct {
	resolver.AuthSSLCert
	VerifyClient       string `json:"verify_client"`
	ValidationDepth    int    `json:"validationDepth"`
	ErrorPage          string `json:"errorPage"`
	PassCertToUpstream bool   `json:"passCertToUpstream"`
	MatchCN            string `json:"matchCN"`
	AuthTLSError       string
}

// Equal tests for equality between two Config types
func (assl1 *Config) Equal(assl2 *Config) bool {
	if assl1 == assl2 {
		return true
	}
	if assl1 == nil || assl2 == nil {
		return false
	}
	if !(&assl1.AuthSSLCert).Equal(&assl2.AuthSSLCert) {
		return false
	}
	if assl1.VerifyClient != assl2.VerifyClient {
		return false
	}
	if assl1.ValidationDepth != assl2.ValidationDepth {
		return false
	}
	if assl1.ErrorPage != assl2.ErrorPage {
		return false
	}
	if assl1.PassCertToUpstream != assl2.PassCertToUpstream {
		return false
	}
	if assl1.MatchCN != assl2.MatchCN {
		return false
	}

	return true
}

// NewParser creates a new TLS authentication annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return authTLS{
		r:                r,
		annotationConfig: authTLSAnnotations,
	}
}

type authTLS struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Parse parses the annotations contained in the ingress
// rule used to use a Certificate as authentication method
func (a authTLS) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	tlsauthsecret, err := parser.GetStringAnnotation(annotationAuthTLSSecret, ing, a.annotationConfig.Annotations)
	if err != nil {
		return &Config{}, err
	}

	ns, _, err := k8s.ParseNameNS(tlsauthsecret)
	if err != nil {
		return &Config{}, ing_errors.NewLocationDenied(err.Error())
	}
	if ns == "" {
		ns = ing.Namespace
	}
	secCfg := a.r.GetSecurityConfiguration()
	// We don't accept different namespaces for secrets.
	if !secCfg.AllowCrossNamespaceResources && ns != ing.Namespace {
		return &Config{}, ing_errors.NewLocationDenied("cross namespace secrets are not supported")
	}

	authCert, err := a.r.GetAuthCertificate(tlsauthsecret)
	if err != nil {
		e := fmt.Errorf("error obtaining certificate: %w", err)
		return &Config{}, ing_errors.LocationDeniedError{Reason: e}
	}
	config.AuthSSLCert = *authCert

	config.VerifyClient, err = parser.GetStringAnnotation(annotationAuthTLSVerifyClient, ing, a.annotationConfig.Annotations)
	// We can set a default value here in case of validation error
	if err != nil || !authVerifyClientRegex.MatchString(config.VerifyClient) {
		config.VerifyClient = defaultAuthVerifyClient
	}

	config.ValidationDepth, err = parser.GetIntAnnotation(annotationAuthTLSVerifyDepth, ing, a.annotationConfig.Annotations)
	// We can set a default value here in case of validation error
	if err != nil || config.ValidationDepth == 0 {
		config.ValidationDepth = defaultAuthTLSDepth
	}

	config.ErrorPage, err = parser.GetStringAnnotation(annotationAuthTLSErrorPage, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return &Config{}, err
		}
		config.ErrorPage = ""
	}

	config.PassCertToUpstream, err = parser.GetBoolAnnotation(annotationAuthTLSPassCertToUpstream, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return &Config{}, err
		}
		config.PassCertToUpstream = false
	}

	config.MatchCN, err = parser.GetStringAnnotation(annotationAuthTLSMatchCN, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return &Config{}, err
		}
		config.MatchCN = ""
	}

	return config, nil
}

func (a authTLS) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a authTLS) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, authTLSAnnotations.Annotations)
}
