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

package certificatewithouthostnameverification

import (
	networking "k8s.io/api/networking/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type certificateWithoutHostnameVerification struct {
	r resolver.Resolver
}

// NewParser creates an annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return certificateWithoutHostnameVerification{r}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the redirects must
func (a certificateWithoutHostnameVerification) Parse(ing *networking.Ingress) (interface{}, error) {
	// Default to false in case of no/empty value, since previous default behavior was to verify hostname
	value, err := parser.GetBoolAnnotation("use-certificate-without-hostname-verification", ing)
	if err != nil {
		return false, nil
	}
	return value, nil
}
