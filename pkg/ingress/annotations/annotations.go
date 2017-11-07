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

	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/pkg/ingress/annotations/alias"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/auth"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/authreq"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/authtls"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/clientbodybuffersize"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/cors"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/defaultbackend"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/healthcheck"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/ipwhitelist"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/parser"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/portinredirect"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/proxy"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/ratelimit"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/redirect"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/rewrite"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/secureupstream"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/serversnippet"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/serviceupstream"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/sessionaffinity"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/snippet"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/sslpassthrough"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/upstreamhashby"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/upstreamvhost"
	"k8s.io/ingress-nginx/pkg/ingress/annotations/vtsfilterkey"
	"k8s.io/ingress-nginx/pkg/ingress/errors"
	"k8s.io/ingress-nginx/pkg/ingress/resolver"
)

// DeniedKeyName name of the key that contains the reason to deny a location
const DeniedKeyName = "Denied"

type config interface {
	resolver.AuthCertificate
	resolver.DefaultBackend
	resolver.Secret
	resolver.Service
}

// Ingress defines the valid annotations present in one NGINX Ingress rule
type Ingress struct {
	metav1.ObjectMeta
	Alias                string
	BasicDigestAuth      auth.Config
	CertificateAuth      authtls.Config
	ClientBodyBufferSize string
	ConfigurationSnippet string
	CorsConfig           cors.Config
	DefaultBackend       string
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
	UpstreamVhost        string
	VtsFilterKey         string
	Whitelist            ipwhitelist.SourceRange
}

// Extractor defines the annotation parsers to be used in the extraction of annotations
type Extractor struct {
	secretResolver resolver.Secret
	annotations    map[string]parser.IngressAnnotation
}

// NewAnnotationExtractor creates a new annotations extractor
func NewAnnotationExtractor(cfg config) Extractor {
	return Extractor{
		cfg,
		map[string]parser.IngressAnnotation{
			"Alias":                alias.NewParser(),
			"BasicDigestAuth":      auth.NewParser(auth.AuthDirectory, cfg),
			"CertificateAuth":      authtls.NewParser(cfg),
			"ClientBodyBufferSize": clientbodybuffersize.NewParser(),
			"ConfigurationSnippet": snippet.NewParser(),
			"CorsConfig":           cors.NewParser(),
			"DefaultBackend":       defaultbackend.NewParser(cfg),
			"ExternalAuth":         authreq.NewParser(),
			"HealthCheck":          healthcheck.NewParser(cfg),
			"Proxy":                proxy.NewParser(cfg),
			"RateLimit":            ratelimit.NewParser(cfg),
			"Redirect":             redirect.NewParser(),
			"Rewrite":              rewrite.NewParser(cfg),
			"SecureUpstream":       secureupstream.NewParser(cfg),
			"ServerSnippet":        serversnippet.NewParser(),
			"ServiceUpstream":      serviceupstream.NewParser(),
			"SessionAffinity":      sessionaffinity.NewParser(),
			"SSLPassthrough":       sslpassthrough.NewParser(),
			"UsePortInRedirects":   portinredirect.NewParser(cfg),
			"UpstreamHashBy":       upstreamhashby.NewParser(),
			"UpstreamVhost":        upstreamvhost.NewParser(),
			"VtsFilterKey":         vtsfilterkey.NewParser(),
			"Whitelist":            ipwhitelist.NewParser(cfg),
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

	err := mergo.Map(pia, data)
	if err != nil {
		glog.Errorf("unexpected error merging extracted annotations: %v", err)
	}

	return pia
}
