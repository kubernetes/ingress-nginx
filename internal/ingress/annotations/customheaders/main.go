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
)

// ValidHeader checks is the provided string satisfies the header's name regex
func ValidHeader(header string) bool {
	return headerRegexp.MatchString(header)
}

type customHeaders struct {
	r resolver.Resolver
}

// NewParser creates a new custom response headers annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return customHeaders{r}
}

// Parse parses the annotations contained in the ingress to use
// custom response headers
func (a customHeaders) Parse(ing *networking.Ingress) (interface{}, error) {
	clientHeadersConfigMapName, err := parser.GetStringAnnotation("custom-headers", ing)
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

		for header := range clientHeadersMapContents.Data {
			if !ValidHeader(header) {
				return nil, ing_errors.NewLocationDenied("invalid client-headers in configmap")
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
