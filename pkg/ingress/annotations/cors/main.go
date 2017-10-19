/*
Copyright 2016 The Kubernetes Authors.

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

package cors

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/pkg/ingress/annotations/parser"
)

const (
	annotationCorsEnabled          = "ingress.kubernetes.io/enable-cors"
	annotationCorsAllowOrigin      = "ingress.kubernetes.io/cors-allow-origin"
	annotationCorsAllowMethods     = "ingress.kubernetes.io/cors-allow-methods"
	annotationCorsAllowHeaders     = "ingress.kubernetes.io/cors-allow-headers"
	annotationCorsAllowCredentials = "ingress.kubernetes.io/cors-allow-credentials"
)

type cors struct {
}

// CorsConfig contains the Cors configuration to be used in the Ingress
type CorsConfig struct {
	CorsEnabled          bool   `json:"corsEnabled"`
	CorsAllowOrigin      string `json:"corsAllowOrigin"`
	CorsAllowMethods     string `json:"corsAllowMethods"`
	CorsAllowHeaders     string `json:"corsAllowHeaders"`
	CorsAllowCredentials bool   `json:"corsAllowCredentials"`
}

// NewParser creates a new CORS annotation parser
func NewParser() parser.IngressAnnotation {
	return cors{}
}

// Parse parses the annotations contained in the ingress
// rule used to indicate if the location/s should allows CORS
func (a cors) Parse(ing *extensions.Ingress) (interface{}, error) {
	corsenabled, err := parser.GetBoolAnnotation(annotationCorsEnabled, ing)
	if err != nil {
		corsenabled = false
	}

	corsalloworigin, err := parser.GetStringAnnotation(annotationCorsAllowOrigin, ing)
	if err != nil || corsalloworigin == "" {
		corsalloworigin = "*"
	}

	corsallowheaders, err := parser.GetStringAnnotation(annotationCorsAllowHeaders, ing)
	if err != nil || corsallowheaders == "" {
		corsallowheaders = "'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization"
	}

	corsallowmethods, err := parser.GetStringAnnotation(annotationCorsAllowMethods, ing)
	if err != nil || corsallowmethods == "" {
		corsallowheaders = "GET, PUT, POST, DELETE, PATCH, OPTIONS"
	}

	corsallowcredentials, err := parser.GetBoolAnnotation(annotationCorsAllowCredentials, ing)
	if err != nil {
		corsallowcredentials = true
	}

	return &CorsConfig{
		CorsEnabled:          corsenabled,
		CorsAllowOrigin:      corsalloworigin,
		CorsAllowHeaders:     corsallowheaders,
		CorsAllowMethods:     corsallowmethods,
		CorsAllowCredentials: corsallowcredentials,
	}, nil

}
