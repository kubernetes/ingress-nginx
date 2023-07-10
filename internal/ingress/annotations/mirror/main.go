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

package mirror

import (
	"fmt"
	"regexp"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/klog/v2"
)

const (
	mirrorRequestBodyAnnotation = "mirror-request-body"
	mirrorTargetAnnotation      = "mirror-target"
	mirrorHostAnnotation        = "mirror-host"
)

var OnOffRegex = regexp.MustCompile(`^(on|off)$`)

var mirrorAnnotation = parser.Annotation{
	Group: "mirror",
	Annotations: parser.AnnotationFields{
		mirrorRequestBodyAnnotation: {
			Validator:     parser.ValidateRegex(OnOffRegex, true),
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines if the request-body should be sent to the mirror backend. Can be 'on' or 'off'`,
		},
		mirrorTargetAnnotation: {
			Validator:     parser.ValidateServerName,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation enables a request to be mirrored to a mirror backend.`,
		},
		mirrorHostAnnotation: {
			Validator:     parser.ValidateServerName,
			Scope:         parser.AnnotationScopeIngress,
			Risk:          parser.AnnotationRiskHigh,
			Documentation: `This annotation defines if a specific Host header should be set for mirrored request.`,
		},
	},
}

// Config returns the mirror to use in a given location
type Config struct {
	Source      string `json:"source"`
	RequestBody string `json:"requestBody"`
	Target      string `json:"target"`
	Host        string `json:"host"`
}

// Equal tests for equality between two Configuration types
func (m1 *Config) Equal(m2 *Config) bool {
	if m1 == m2 {
		return true
	}

	if m1 == nil || m2 == nil {
		return false
	}

	if m1.Source != m2.Source {
		return false
	}

	if m1.RequestBody != m2.RequestBody {
		return false
	}

	if m1.Target != m2.Target {
		return false
	}

	if m1.Host != m2.Host {
		return false
	}

	return true
}

type mirror struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// NewParser creates a new mirror configuration annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return mirror{
		r:                r,
		annotationConfig: mirrorAnnotation,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to configure mirror
func (a mirror) Parse(ing *networking.Ingress) (interface{}, error) {
	config := &Config{
		Source: fmt.Sprintf("/_mirror-%v", ing.UID),
	}

	var err error
	config.RequestBody, err = parser.GetStringAnnotation(mirrorRequestBodyAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil || config.RequestBody != "off" {
		if errors.IsValidationError(err) {
			klog.Warningf("annotation %s contains invalid value", mirrorRequestBodyAnnotation)
		}
		config.RequestBody = "on"
	}

	config.Target, err = parser.GetStringAnnotation(mirrorTargetAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("annotation %s contains invalid value, defaulting", mirrorTargetAnnotation)
		} else {
			config.Target = ""
			config.Source = ""
		}
	}

	config.Host, err = parser.GetStringAnnotation(mirrorHostAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("annotation %s contains invalid value, defaulting", mirrorHostAnnotation)
		}
		if config.Target != "" {
			target := strings.Split(config.Target, "$")

			url, err := parser.StringToURL(target[0])
			if err != nil {
				config.Host = ""
			} else {
				config.Host = url.Hostname()
			}
		}
	}

	return config, nil
}

func (a mirror) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a mirror) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, mirrorAnnotation.Annotations)
}
