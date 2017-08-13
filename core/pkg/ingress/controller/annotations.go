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

package controller

import (
	"github.com/golang/glog"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/ingress/core/pkg/ingress/annotations/auth"
	"k8s.io/ingress/core/pkg/ingress/annotations/authreq"
	"k8s.io/ingress/core/pkg/ingress/annotations/authtls"
	"k8s.io/ingress/core/pkg/ingress/annotations/cors"
	"k8s.io/ingress/core/pkg/ingress/annotations/healthcheck"
	"k8s.io/ingress/core/pkg/ingress/annotations/ipwhitelist"
	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/annotations/portinredirect"
	"k8s.io/ingress/core/pkg/ingress/annotations/proxy"
	"k8s.io/ingress/core/pkg/ingress/annotations/ratelimit"
	"k8s.io/ingress/core/pkg/ingress/annotations/rewrite"
	"k8s.io/ingress/core/pkg/ingress/annotations/secureupstream"
	"k8s.io/ingress/core/pkg/ingress/annotations/serviceupstream"
	"k8s.io/ingress/core/pkg/ingress/annotations/sessionaffinity"
	"k8s.io/ingress/core/pkg/ingress/annotations/snippet"
	"k8s.io/ingress/core/pkg/ingress/annotations/sslpassthrough"
	"k8s.io/ingress/core/pkg/ingress/errors"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

type extractorConfig interface {
	resolver.AuthCertificate
	resolver.DefaultBackend
	resolver.Secret
}

type annotationExtractor struct {
	secretResolver resolver.Secret
	annotations    map[string]parser.IngressAnnotation
}

func newAnnotationExtractor(cfg extractorConfig) annotationExtractor {
	return annotationExtractor{
		cfg,
		map[string]parser.IngressAnnotation{
			"BasicDigestAuth":      auth.NewParser(auth.AuthDirectory, cfg),
			"ExternalAuth":         authreq.NewParser(),
			"CertificateAuth":      authtls.NewParser(cfg),
			"EnableCORS":           cors.NewParser(),
			"HealthCheck":          healthcheck.NewParser(cfg),
			"Whitelist":            ipwhitelist.NewParser(cfg),
			"UsePortInRedirects":   portinredirect.NewParser(cfg),
			"Proxy":                proxy.NewParser(cfg),
			"RateLimit":            ratelimit.NewParser(cfg),
			"Redirect":             rewrite.NewParser(cfg),
			"SecureUpstream":       secureupstream.NewParser(cfg),
			"ServiceUpstream":      serviceupstream.NewParser(),
			"SessionAffinity":      sessionaffinity.NewParser(),
			"SSLPassthrough":       sslpassthrough.NewParser(),
			"ConfigurationSnippet": snippet.NewParser(),
		},
	}
}

func (e *annotationExtractor) Extract(ing *extensions.Ingress) map[string]interface{} {
	anns := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		val, err := annotationParser.Parse(ing)
		glog.V(5).Infof("annotation %v in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), val)
		if err != nil {
			if errors.IsMissingAnnotations(err) {
				continue
			}

			_, alreadyDenied := anns[DeniedKeyName]
			if !alreadyDenied {
				anns[DeniedKeyName] = err
				glog.Errorf("error reading %v annotation in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), err)
				continue
			}

			glog.V(5).Infof("error reading %v annotation in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), err)
		}

		if val != nil {
			anns[name] = val
		}
	}

	return anns
}

const (
	secureUpstream  = "SecureUpstream"
	healthCheck     = "HealthCheck"
	sslPassthrough  = "SSLPassthrough"
	sessionAffinity = "SessionAffinity"
	serviceUpstream = "ServiceUpstream"
)

func (e *annotationExtractor) ServiceUpstream(ing *extensions.Ingress) bool {
	val, _ := e.annotations[serviceUpstream].Parse(ing)
	return val.(bool)
}

func (e *annotationExtractor) SecureUpstream(ing *extensions.Ingress) *secureupstream.Secure {
	val, err := e.annotations[secureUpstream].Parse(ing)
	if err != nil {
		glog.Errorf("error parsing secure upstream: %v", err)
	}
	secure := val.(*secureupstream.Secure)
	return secure
}

func (e *annotationExtractor) HealthCheck(ing *extensions.Ingress) *healthcheck.Upstream {
	val, _ := e.annotations[healthCheck].Parse(ing)
	return val.(*healthcheck.Upstream)
}

func (e *annotationExtractor) SSLPassthrough(ing *extensions.Ingress) bool {
	val, _ := e.annotations[sslPassthrough].Parse(ing)
	return val.(bool)
}

func (e *annotationExtractor) SessionAffinity(ing *extensions.Ingress) *sessionaffinity.AffinityConfig {
	val, _ := e.annotations[sessionAffinity].Parse(ing)
	return val.(*sessionaffinity.AffinityConfig)
}
