/*
Copyright 2023 The Kubernetes Authors.

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

package disableproxyintercepterrors

import (
	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	disableProxyInterceptErrorsAnnotation = "disable-proxy-intercept-errors"
)

var disableProxyInterceptErrorsAnnotations = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		disableProxyInterceptErrorsAnnotation: {
			Validator: parser.ValidateBool,
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation allows to disable NGINX proxy-intercept-errors when custom-http-errors are set.
			If a default backend annotation is specified on the ingress, the errors will be routed to that annotation's default backend service (instead of the global default backend).
			Different ingresses can specify different sets of errors codes and there are UseCases where NGINX shall not intercept all errors returned from upstream.`,
		},
	},
}

type disableProxyInterceptErrors struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

func (pie disableProxyInterceptErrors) GetDocumentation() parser.AnnotationFields {
	return pie.annotationConfig.Annotations
}

func (pie disableProxyInterceptErrors) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(pie.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, disableProxyInterceptErrorsAnnotations.Annotations)
}

// NewParser creates a new disableProxyInterceptErrors annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return disableProxyInterceptErrors{
		r:                r,
		annotationConfig: disableProxyInterceptErrorsAnnotations,
	}
}

func (pie disableProxyInterceptErrors) Parse(ing *networking.Ingress) (interface{}, error) {
	val, err := parser.GetBoolAnnotation(disableProxyInterceptErrorsAnnotation, ing, pie.annotationConfig.Annotations)

	// A missing annotation is not a problem, just use the default
	if err == errors.ErrMissingAnnotations {
		return false, nil // default is false
	}

	return val, nil
}
