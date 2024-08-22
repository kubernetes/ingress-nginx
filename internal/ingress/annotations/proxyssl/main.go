/*
Copyright 2019 The Kubernetes Authors.

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

package proxyssl

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/klog/v2"
)

const (
	defaultProxySSLCiphers     = "DEFAULT"
	defaultProxySSLProtocols   = "TLSv1.2"
	defaultProxySSLVerify      = "off"
	defaultProxySSLVerifyDepth = 1
	defaultProxySSLServerName  = "off"
)

var (
	proxySSLOnOffRegex    = regexp.MustCompile(`^(on|off)$`)
	proxySSLProtocolRegex = regexp.MustCompile(`^(TLSv1\.2|TLSv1\.3| )*$`)
	proxySSLCiphersRegex  = regexp.MustCompile(`^[A-Za-z0-9\+:\_\-!]*$`)
)

const (
	proxySSLSecretAnnotation      = "proxy-ssl-secret"
	proxySSLCiphersAnnotation     = "proxy-ssl-ciphers"
	proxySSLProtocolsAnnotation   = "proxy-ssl-protocols"
	proxySSLNameAnnotation        = "proxy-ssl-name"
	proxySSLVerifyAnnotation      = "proxy-ssl-verify"
	proxySSLVerifyDepthAnnotation = "proxy-ssl-verify-depth"
	proxySSLServerNameAnnotation  = "proxy-ssl-server-name"
)

var proxySSLAnnotation = parser.Annotation{
	Group: "proxy",
	Annotations: parser.AnnotationFields{
		proxySSLSecretAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation specifies a Secret with the certificate tls.crt, key tls.key in PEM format used for authentication to a proxied HTTPS server. 
			It should also contain trusted CA certificates ca.crt in PEM format used to verify the certificate of the proxied HTTPS server. 
			This annotation expects the Secret name in the form "namespace/secretName"
			Just secrets on the same namespace of the ingress can be used.`,
		},
		proxySSLCiphersAnnotation: {
			Validator: parser.ValidateRegex(proxySSLCiphersRegex, true),
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation Specifies the enabled ciphers for requests to a proxied HTTPS server. 
			The ciphers are specified in the format understood by the OpenSSL library.`,
		},
		proxySSLProtocolsAnnotation: {
			Validator:     parser.ValidateRegex(proxySSLProtocolRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables the specified protocols for requests to a proxied HTTPS server.`,
		},
		proxySSLNameAnnotation: {
			Validator: parser.ValidateServerName,
			Scope:     parser.AnnotationScopeIngress,
			Risk:      parser.AnnotationRiskHigh,
			Documentation: `This annotation allows to set proxy_ssl_name. This allows overriding the server name used to verify the certificate of the proxied HTTPS server. 
			This value is also passed through SNI when a connection is established to the proxied HTTPS server.`,
		},
		proxySSLVerifyAnnotation: {
			Validator:     parser.ValidateRegex(proxySSLOnOffRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables or disables verification of the proxied HTTPS server certificate. (default: off)`,
		},
		proxySSLVerifyDepthAnnotation: {
			Validator:     parser.ValidateInt,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation Sets the verification depth in the proxied HTTPS server certificates chain. (default: 1).`,
		},
		proxySSLServerNameAnnotation: {
			Validator:     parser.ValidateRegex(proxySSLOnOffRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation enables passing of the server name through TLS Server Name Indication extension (SNI, RFC 6066) when establishing a connection with the proxied HTTPS server.`,
		},
	},
}

// Config contains the AuthSSLCert used for mutual authentication
// and the configured VerifyDepth
type Config struct {
	resolver.AuthSSLCert
	Ciphers            string `json:"ciphers"`
	Protocols          string `json:"protocols"`
	ProxySSLName       string `json:"proxySSLName"`
	Verify             string `json:"verify"`
	VerifyDepth        int    `json:"verifyDepth"`
	ProxySSLServerName string `json:"proxySSLServerName"`
}

// Equal tests for equality between two Config types
func (pssl1 *Config) Equal(pssl2 *Config) bool {
	if pssl1 == pssl2 {
		return true
	}
	if pssl1 == nil || pssl2 == nil {
		return false
	}
	if !(&pssl1.AuthSSLCert).Equal(&pssl2.AuthSSLCert) {
		return false
	}
	if pssl1.Ciphers != pssl2.Ciphers {
		return false
	}
	if pssl1.Protocols != pssl2.Protocols {
		return false
	}
	if pssl1.Verify != pssl2.Verify {
		return false
	}
	if pssl1.VerifyDepth != pssl2.VerifyDepth {
		return false
	}
	if pssl1.ProxySSLServerName != pssl2.ProxySSLServerName {
		return false
	}
	return true
}

// NewParser creates a new TLS authentication annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return proxySSL{
		r:                r,
		annotationConfig: proxySSLAnnotation,
	}
}

type proxySSL struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

func sortProtocols(protocols string) string {
	protolist := strings.Split(protocols, " ")

	n := 0
	for _, proto := range protolist {
		proto = strings.TrimSpace(proto)
		if proto == "" || !proxySSLProtocolRegex.MatchString(proto) {
			continue
		}
		protolist[n] = proto
		n++
	}

	if n == 0 {
		return defaultProxySSLProtocols
	}

	protolist = protolist[:n]
	sort.Strings(protolist)
	return strings.Join(protolist, " ")
}

// Parse parses the annotations contained in the ingress
// rule used to use a Certificate as authentication method
func (p proxySSL) Parse(ing *networking.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	proxysslsecret, err := parser.GetStringAnnotation(proxySSLSecretAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil {
		return &Config{}, err
	}

	ns, _, err := k8s.ParseNameNS(proxysslsecret)
	if err != nil {
		return &Config{}, ing_errors.NewLocationDenied(err.Error())
	}

	secCfg := p.r.GetSecurityConfiguration()
	// We don't accept different namespaces for secrets.
	if !secCfg.AllowCrossNamespaceResources && ns != ing.Namespace {
		return &Config{}, ing_errors.NewLocationDenied("cross namespace secrets are not supported")
	}

	proxyCert, err := p.r.GetAuthCertificate(proxysslsecret)
	if err != nil {
		e := fmt.Errorf("error obtaining certificate: %w", err)
		return &Config{}, ing_errors.LocationDeniedError{Reason: e}
	}
	config.AuthSSLCert = *proxyCert

	config.Ciphers, err = parser.GetStringAnnotation(proxySSLCiphersAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			klog.Warningf("invalid value passed to proxy-ssl-ciphers, defaulting to %s", defaultProxySSLCiphers)
		}
		config.Ciphers = defaultProxySSLCiphers
	}

	config.Protocols, err = parser.GetStringAnnotation(proxySSLProtocolsAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			klog.Warningf("invalid value passed to proxy-ssl-protocols, defaulting to %s", defaultProxySSLProtocols)
		}
		config.Protocols = defaultProxySSLProtocols
	} else {
		config.Protocols = sortProtocols(config.Protocols)
	}

	config.ProxySSLName, err = parser.GetStringAnnotation(proxySSLNameAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			klog.Warningf("invalid value passed to proxy-ssl-name, defaulting to empty")
		}
		config.ProxySSLName = ""
	}

	config.Verify, err = parser.GetStringAnnotation(proxySSLVerifyAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil || !proxySSLOnOffRegex.MatchString(config.Verify) {
		config.Verify = defaultProxySSLVerify
	}

	config.VerifyDepth, err = parser.GetIntAnnotation(proxySSLVerifyDepthAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil || config.VerifyDepth == 0 {
		config.VerifyDepth = defaultProxySSLVerifyDepth
	}

	config.ProxySSLServerName, err = parser.GetStringAnnotation(proxySSLServerNameAnnotation, ing, p.annotationConfig.Annotations)
	if err != nil || !proxySSLOnOffRegex.MatchString(config.ProxySSLServerName) {
		config.ProxySSLServerName = defaultProxySSLServerName
	}

	return config, nil
}

func (p proxySSL) GetDocumentation() parser.AnnotationFields {
	return p.annotationConfig.Annotations
}

func (p proxySSL) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(p.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, proxySSLAnnotation.Annotations)
}
