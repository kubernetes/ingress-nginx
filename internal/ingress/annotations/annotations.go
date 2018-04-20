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
	"github.com/golang/glog"
	"github.com/imdario/mergo"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sslcipher"

	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/alias"
	"k8s.io/ingress-nginx/internal/ingress/annotations/auth"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authreq"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authtls"
	"k8s.io/ingress-nginx/internal/ingress/annotations/clientbodybuffersize"
	"k8s.io/ingress-nginx/internal/ingress/annotations/connection"
	"k8s.io/ingress-nginx/internal/ingress/annotations/cors"
	"k8s.io/ingress-nginx/internal/ingress/annotations/defaultbackend"
	"k8s.io/ingress-nginx/internal/ingress/annotations/grpc"
	"k8s.io/ingress-nginx/internal/ingress/annotations/healthcheck"
	"k8s.io/ingress-nginx/internal/ingress/annotations/http2"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ipwhitelist"
	"k8s.io/ingress-nginx/internal/ingress/annotations/loadbalancing"
	"k8s.io/ingress-nginx/internal/ingress/annotations/log"
	"k8s.io/ingress-nginx/internal/ingress/annotations/luarestywaf"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/annotations/portinredirect"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxy"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ratelimit"
	"k8s.io/ingress-nginx/internal/ingress/annotations/redirect"
	"k8s.io/ingress-nginx/internal/ingress/annotations/rewrite"
	"k8s.io/ingress-nginx/internal/ingress/annotations/secureupstream"
	"k8s.io/ingress-nginx/internal/ingress/annotations/serversnippet"
	"k8s.io/ingress-nginx/internal/ingress/annotations/serviceupstream"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sessionaffinity"
	"k8s.io/ingress-nginx/internal/ingress/annotations/snippet"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sslpassthrough"
	"k8s.io/ingress-nginx/internal/ingress/annotations/upstreamhashby"
	"k8s.io/ingress-nginx/internal/ingress/annotations/upstreamvhost"
	"k8s.io/ingress-nginx/internal/ingress/annotations/vtsfilterkey"
	"k8s.io/ingress-nginx/internal/ingress/annotations/xforwardedprefix"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// DeniedKeyName name of the key that contains the reason to deny a location
const DeniedKeyName = "Denied"

// Ingress defines the valid annotations present in one NGINX Ingress rule
type Ingress struct {
	metav1.ObjectMeta
	Alias                string
	BasicDigestAuth      auth.Config
	CertificateAuth      authtls.Config
	ClientBodyBufferSize string
	ConfigurationSnippet string
	Connection           connection.Config
	CorsConfig           cors.Config
	DefaultBackend       string
	Denied               error
	ExternalAuth         authreq.Config
	HealthCheck          healthcheck.Config
	Proxy                proxy.Config
	RateLimit            ratelimit.Config
	Redirect             redirect.Config
	Rewrite              rewrite.Config
	SecureUpstream       secureupstream.Config
	ServerSnippet        string
	ServiceUpstream      bool
	SessionAffinity      sessionaffinity.Config
	SSLPassthrough       bool
	UsePortInRedirects   bool
	UpstreamHashBy       string
	LoadBalancing        string
	UpstreamVhost        string
	VtsFilterKey         string
	Whitelist            ipwhitelist.SourceRange
	XForwardedPrefix     bool
	SSLCiphers           string
	Logs                 log.Config
	GRPC                 bool
	LuaRestyWAF          luarestywaf.Config
	UseHTTP2             bool
}

// Extractor defines the annotation parsers to be used in the extraction of annotations
type Extractor struct {
	annotations map[string]parser.IngressAnnotation
}

// NewAnnotationExtractor creates a new annotations extractor
func NewAnnotationExtractor(cfg resolver.Resolver) Extractor {
	return Extractor{
		map[string]parser.IngressAnnotation{
			"Alias":                alias.NewParser(cfg),
			"BasicDigestAuth":      auth.NewParser(auth.AuthDirectory, cfg),
			"CertificateAuth":      authtls.NewParser(cfg),
			"ClientBodyBufferSize": clientbodybuffersize.NewParser(cfg),
			"ConfigurationSnippet": snippet.NewParser(cfg),
			"Connection":           connection.NewParser(cfg),
			"CorsConfig":           cors.NewParser(cfg),
			"DefaultBackend":       defaultbackend.NewParser(cfg),
			"ExternalAuth":         authreq.NewParser(cfg),
			"HealthCheck":          healthcheck.NewParser(cfg),
			"Proxy":                proxy.NewParser(cfg),
			"RateLimit":            ratelimit.NewParser(cfg),
			"Redirect":             redirect.NewParser(cfg),
			"Rewrite":              rewrite.NewParser(cfg),
			"SecureUpstream":       secureupstream.NewParser(cfg),
			"ServerSnippet":        serversnippet.NewParser(cfg),
			"ServiceUpstream":      serviceupstream.NewParser(cfg),
			"SessionAffinity":      sessionaffinity.NewParser(cfg),
			"SSLPassthrough":       sslpassthrough.NewParser(cfg),
			"UsePortInRedirects":   portinredirect.NewParser(cfg),
			"UpstreamHashBy":       upstreamhashby.NewParser(cfg),
			"LoadBalancing":        loadbalancing.NewParser(cfg),
			"UpstreamVhost":        upstreamvhost.NewParser(cfg),
			"VtsFilterKey":         vtsfilterkey.NewParser(cfg),
			"Whitelist":            ipwhitelist.NewParser(cfg),
			"XForwardedPrefix":     xforwardedprefix.NewParser(cfg),
			"SSLCiphers":           sslcipher.NewParser(cfg),
			"Logs":                 log.NewParser(cfg),
			"GRPC":                 grpc.NewParser(cfg),
			"LuaRestyWAF":          luarestywaf.NewParser(cfg),
			"UseHTTP2":             http2.NewParser(cfg),
		},
	}
}

// Extract extracts the annotations from an Ingress
func (e Extractor) Extract(ing *extensions.Ingress) *Ingress {
	pia := &Ingress{
		ObjectMeta: ing.ObjectMeta,
	}

	data := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		val, err := annotationParser.Parse(ing)
		glog.V(5).Infof("annotation %v in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), val)
		if err != nil {
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
				data[DeniedKeyName] = err
				glog.Errorf("error reading %v annotation in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), err)
				continue
			}

			glog.V(5).Infof("error reading %v annotation in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), err)
		}

		if val != nil {
			data[name] = val
		}
	}

	err := mergo.MapWithOverwrite(pia, data)
	if err != nil {
		glog.Errorf("unexpected error merging extracted annotations: %v", err)
	}

	return pia
}
