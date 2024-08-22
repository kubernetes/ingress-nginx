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

package customheaders

import (
	"fmt"
	"regexp"

	"k8s.io/klog/v2"

	networking "k8s.io/api/networking/v1"

	"golang.org/x/exp/slices"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Config returns the custom response headers for an Ingress rule
type Config struct {
	Headers map[string]string `json:"headers,omitempty"`
}

var (
	headerRegexp = regexp.MustCompile(`^[a-zA-Z\d\-_]+$`)
	valueRegexp  = regexp.MustCompile(`^[a-zA-Z\d_ :;.,\\/"'?!(){}\[\]@<>=\-+*#$&\x60|~^%]+$`)
)

// ValidHeader checks is the provided string satisfies the header's name regex
func ValidHeader(header string) bool {
	return headerRegexp.MatchString(header)
}

// ValidValue checks is the provided string satisfies the value regex
func ValidValue(header string) bool {
	return valueRegexp.MatchString(header)
}

const (
	customHeadersConfigMapAnnotation = "custom-headers"
)

var customHeadersAnnotation = parser.Annotation{
	Group: "backend",
	Annotations: parser.AnnotationFields{
		customHeadersConfigMapAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation sets the name of a ConfigMap that specifies headers to pass to the client.
			Only ConfigMaps on the same namespace are allowed`,
		},
	},
}

type customHeaders struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new custom response headers annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return customHeaders{r: r, annotationConfig: customHeadersAnnotation}
}

func (a customHeaders) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

// Parse parses the annotations contained in the ingress to use
// custom response headers
func (a customHeaders) Parse(ing *networking.Ingress) (interface{}, error) {
	clientHeadersConfigMapName, err := parser.GetStringAnnotation(customHeadersConfigMapAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		klog.V(3).InfoS("client-headers annotation is undefined and will not be set")
	}

	var headers map[string]string
	defBackend := a.r.GetDefaultBackend()

	if clientHeadersConfigMapName != "" {
		clientHeadersMapContents, err := a.r.GetConfigMap(clientHeadersConfigMapName)
		if err != nil {
			return nil, ing_errors.NewLocationDenied(fmt.Sprintf("unable to find configMap %q", clientHeadersConfigMapName))
		}

		for header, value := range clientHeadersMapContents.Data {
			if !ValidHeader(header) {
				return nil, ing_errors.NewLocationDenied("invalid header name in configmap")
			}
			if !ValidValue(value) {
				return nil, ing_errors.NewLocationDenied("invalid header value in configmap")
			}
			if !slices.Contains(defBackend.AllowedResponseHeaders, header) {
				return nil, ing_errors.NewLocationDenied(fmt.Sprintf("header %s is not allowed, defined allowed headers inside global-allowed-response-headers %v", header, defBackend.AllowedResponseHeaders))
			}
		}

		headers = clientHeadersMapContents.Data
	}

	return &Config{
		Headers: headers,
	}, nil
}

func (a customHeaders) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, customHeadersAnnotation.Annotations)
}
