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

package annotations

import (
	"dario.cat/mergo"

	apiv1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/alias"
	"k8s.io/ingress-nginx/internal/ingress/annotations/auth"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authreq"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authreqglobal"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authtls"
	"k8s.io/ingress-nginx/internal/ingress/annotations/backendprotocol"
	"k8s.io/ingress-nginx/internal/ingress/annotations/canary"
	"k8s.io/ingress-nginx/internal/ingress/annotations/client"
	"k8s.io/ingress-nginx/internal/ingress/annotations/connection"
	"k8s.io/ingress-nginx/internal/ingress/annotations/cors"
	"k8s.io/ingress-nginx/internal/ingress/annotations/customheaders"
	"k8s.io/ingress-nginx/internal/ingress/annotations/customhttperrors"
	"k8s.io/ingress-nginx/internal/ingress/annotations/defaultbackend"
	"k8s.io/ingress-nginx/internal/ingress/annotations/disableproxyintercepterrors"
	"k8s.io/ingress-nginx/internal/ingress/annotations/fastcgi"
	"k8s.io/ingress-nginx/internal/ingress/annotations/http2pushpreload"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ipallowlist"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ipdenylist"
	"k8s.io/ingress-nginx/internal/ingress/annotations/loadbalancing"
	"k8s.io/ingress-nginx/internal/ingress/annotations/log"
	"k8s.io/ingress-nginx/internal/ingress/annotations/mirror"
	"k8s.io/ingress-nginx/internal/ingress/annotations/modsecurity"
	"k8s.io/ingress-nginx/internal/ingress/annotations/opentelemetry"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/annotations/portinredirect"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxy"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxyssl"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ratelimit"
	"k8s.io/ingress-nginx/internal/ingress/annotations/redirect"
	"k8s.io/ingress-nginx/internal/ingress/annotations/rewrite"
	"k8s.io/ingress-nginx/internal/ingress/annotations/satisfy"
	"k8s.io/ingress-nginx/internal/ingress/annotations/serversnippet"
	"k8s.io/ingress-nginx/internal/ingress/annotations/serviceupstream"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sessionaffinity"
	"k8s.io/ingress-nginx/internal/ingress/annotations/snippet"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sslcipher"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sslpassthrough"
	"k8s.io/ingress-nginx/internal/ingress/annotations/streamsnippet"
	"k8s.io/ingress-nginx/internal/ingress/annotations/upstreamhashby"
	"k8s.io/ingress-nginx/internal/ingress/annotations/upstreamvhost"
	"k8s.io/ingress-nginx/internal/ingress/annotations/xforwardedprefix"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// DeniedKeyName name of the key that contains the reason to deny a location
const DeniedKeyName = "Denied"

// Ingress defines the valid annotations present in one NGINX Ingress rule
type Ingress struct {
	metav1.ObjectMeta
	BackendProtocol             string
	Aliases                     []string
	BasicDigestAuth             auth.Config
	Canary                      canary.Config
	CertificateAuth             authtls.Config
	Client                      client.Config
	CustomHeaders               customheaders.Config
	ConfigurationSnippet        string
	Connection                  connection.Config
	CorsConfig                  cors.Config
	CustomHTTPErrors            []int
	DisableProxyInterceptErrors bool
	DefaultBackend              *apiv1.Service
	FastCGI                     fastcgi.Config
	Denied                      *string
	ExternalAuth                authreq.Config
	EnableGlobalAuth            bool
	HTTP2PushPreload            bool
	Opentelemetry               opentelemetry.Config
	Proxy                       proxy.Config
	ProxySSL                    proxyssl.Config
	RateLimit                   ratelimit.Config
	Redirect                    redirect.Config
	Rewrite                     rewrite.Config
	Satisfy                     string
	ServerSnippet               string
	ServiceUpstream             bool
	SessionAffinity             sessionaffinity.Config
	SSLPassthrough              bool
	UsePortInRedirects          bool
	UpstreamHashBy              upstreamhashby.Config
	LoadBalancing               string
	UpstreamVhost               string
	Denylist                    ipdenylist.SourceRange
	XForwardedPrefix            string
	SSLCipher                   sslcipher.Config
	Logs                        log.Config
	ModSecurity                 modsecurity.Config
	Mirror                      mirror.Config
	StreamSnippet               string
	Allowlist                   ipallowlist.SourceRange
}

// Extractor defines the annotation parsers to be used in the extraction of annotations
type Extractor struct {
	annotations map[string]parser.IngressAnnotation
}

func NewAnnotationFactory(cfg resolver.Resolver) map[string]parser.IngressAnnotation {
	return map[string]parser.IngressAnnotation{
		"Aliases":                     alias.NewParser(cfg),
		"BasicDigestAuth":             auth.NewParser(auth.AuthDirectory, cfg),
		"Canary":                      canary.NewParser(cfg),
		"CertificateAuth":             authtls.NewParser(cfg),
		"Client":                      client.NewParser(cfg),
		"CustomHeaders":               customheaders.NewParser(cfg),
		"ConfigurationSnippet":        snippet.NewParser(cfg),
		"Connection":                  connection.NewParser(cfg),
		"CorsConfig":                  cors.NewParser(cfg),
		"CustomHTTPErrors":            customhttperrors.NewParser(cfg),
		"DisableProxyInterceptErrors": disableproxyintercepterrors.NewParser(cfg),
		"DefaultBackend":              defaultbackend.NewParser(cfg),
		"FastCGI":                     fastcgi.NewParser(cfg),
		"ExternalAuth":                authreq.NewParser(cfg),
		"EnableGlobalAuth":            authreqglobal.NewParser(cfg),
		"HTTP2PushPreload":            http2pushpreload.NewParser(cfg),
		"Opentelemetry":               opentelemetry.NewParser(cfg),
		"Proxy":                       proxy.NewParser(cfg),
		"ProxySSL":                    proxyssl.NewParser(cfg),
		"RateLimit":                   ratelimit.NewParser(cfg),
		"Redirect":                    redirect.NewParser(cfg),
		"Rewrite":                     rewrite.NewParser(cfg),
		"Satisfy":                     satisfy.NewParser(cfg),
		"ServerSnippet":               serversnippet.NewParser(cfg),
		"ServiceUpstream":             serviceupstream.NewParser(cfg),
		"SessionAffinity":             sessionaffinity.NewParser(cfg),
		"SSLPassthrough":              sslpassthrough.NewParser(cfg),
		"UsePortInRedirects":          portinredirect.NewParser(cfg),
		"UpstreamHashBy":              upstreamhashby.NewParser(cfg),
		"LoadBalancing":               loadbalancing.NewParser(cfg),
		"UpstreamVhost":               upstreamvhost.NewParser(cfg),
		"Allowlist":                   ipallowlist.NewParser(cfg),
		"Denylist":                    ipdenylist.NewParser(cfg),
		"XForwardedPrefix":            xforwardedprefix.NewParser(cfg),
		"SSLCipher":                   sslcipher.NewParser(cfg),
		"Logs":                        log.NewParser(cfg),
		"BackendProtocol":             backendprotocol.NewParser(cfg),
		"ModSecurity":                 modsecurity.NewParser(cfg),
		"Mirror":                      mirror.NewParser(cfg),
		"StreamSnippet":               streamsnippet.NewParser(cfg),
	}
}

// NewAnnotationExtractor creates a new annotations extractor
func NewAnnotationExtractor(cfg resolver.Resolver) Extractor {
	return Extractor{
		NewAnnotationFactory(cfg),
	}
}

// Extract extracts the annotations from an Ingress
func (e Extractor) Extract(ing *networking.Ingress) (*Ingress, error) {
	pia := &Ingress{
		ObjectMeta: ing.ObjectMeta,
	}

	data := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		if err := annotationParser.Validate(ing.GetAnnotations()); err != nil {
			return nil, errors.NewRiskyAnnotations(name)
		}
		val, err := annotationParser.Parse(ing)
		klog.V(5).InfoS("Parsing Ingress annotation", "name", name, "ingress", klog.KObj(ing), "value", val)
		if err != nil {
			if errors.IsValidationError(err) {
				klog.ErrorS(err, "ingress contains invalid annotation value")
				return nil, err
			}
			if errors.IsMissingAnnotations(err) {
				continue
			}

			if !errors.IsLocationDenied(err) {
				continue
			}

			if name == "CertificateAuth" && data[name] == nil {
				data[name] = authtls.Config{
					AuthTLSError: err.Error(),
				}
				// avoid mapping the result from the annotation
				val = nil
			}

			_, alreadyDenied := data[DeniedKeyName]
			if !alreadyDenied {
				errString := err.Error()
				data[DeniedKeyName] = &errString
				klog.ErrorS(err, "error reading Ingress annotation", "name", name, "ingress", klog.KObj(ing))
				continue
			}

			klog.V(5).ErrorS(err, "error reading Ingress annotation", "name", name, "ingress", klog.KObj(ing))
		}

		if val != nil {
			data[name] = val
		}
	}

	err := mergo.MapWithOverwrite(pia, data)
	if err != nil {
		klog.ErrorS(err, "unexpected error merging extracted annotations")
	}

	return pia, nil
}
