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

	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/auth"
	"k8s.io/ingress/core/pkg/ingress/annotations/authreq"
	"k8s.io/ingress/core/pkg/ingress/annotations/authtls"
	"k8s.io/ingress/core/pkg/ingress/annotations/cors"
	"k8s.io/ingress/core/pkg/ingress/annotations/healthcheck"
	"k8s.io/ingress/core/pkg/ingress/annotations/ipwhitelist"
	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/annotations/proxy"
	"k8s.io/ingress/core/pkg/ingress/annotations/ratelimit"
	"k8s.io/ingress/core/pkg/ingress/annotations/rewrite"
	"k8s.io/ingress/core/pkg/ingress/annotations/secureupstream"
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
	annotations map[string]parser.IngressAnnotation
}

func newAnnotationExtractor(cfg extractorConfig) annotationExtractor {
	return annotationExtractor{
		map[string]parser.IngressAnnotation{
			"BasicDigestAuth": auth.NewParser(auth.AuthDirectory, cfg),
			"ExternalAuth":    authreq.NewParser(),
			"CertificateAuth": authtls.NewParser(cfg),
			"EnableCORS":      cors.NewParser(),
			"HealthCheck":     healthcheck.NewParser(cfg),
			"Whitelist":       ipwhitelist.NewParser(cfg),
			"Proxy":           proxy.NewParser(cfg),
			"RateLimit":       ratelimit.NewParser(),
			"Redirect":        rewrite.NewParser(cfg),
			"SecureUpstream":  secureupstream.NewParser(),
			"SSLPassthrough":  sslpassthrough.NewParser(),
		},
	}
}

func (e *annotationExtractor) Extract(ing *extensions.Ingress) map[string]interface{} {
	anns := make(map[string]interface{}, 0)
	for name, annotationParser := range e.annotations {
		val, err := annotationParser.Parse(ing)
		glog.V(5).Infof("annotation %v in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), val)
		if err != nil {
			_, de := anns["Denied"]
			if errors.IsLocationDenied(err) && !de {
				anns["Denied"] = err
				glog.Errorf("error reading %v annotation in Ingress %v/%v: %v", name, ing.GetNamespace(), ing.GetName(), err)
				continue
			}
			if !errors.IsMissingAnnotations(err) {
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
	secureUpstream = "SecureUpstream"
	healthCheck    = "HealthCheck"
	sslPassthrough = "SSLPassthrough"
)

func (e *annotationExtractor) SecureUpstream(ing *extensions.Ingress) bool {
	val, _ := e.annotations[secureUpstream].Parse(ing)
	return val.(bool)
}

func (e *annotationExtractor) HealthCheck(ing *extensions.Ingress) *healthcheck.Upstream {
	val, _ := e.annotations[healthCheck].Parse(ing)
	return val.(*healthcheck.Upstream)
}

func (e *annotationExtractor) SSLPassthrough(ing *extensions.Ingress) bool {
	val, _ := e.annotations[sslPassthrough].Parse(ing)
	return val.(bool)
}
