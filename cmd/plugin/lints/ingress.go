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

package lints

import (
	"fmt"
	"strings"

	networking "k8s.io/api/networking/v1"
	kmeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/cmd/plugin/util"
)

// IngressLint is a validation for an ingress
type IngressLint struct {
	message string
	issue   int
	version string
	f       func(ing networking.Ingress) bool
}

// Check returns true if the lint detects an issue
func (lint IngressLint) Check(obj kmeta.Object) bool {
	ing := obj.(*networking.Ingress)
	return lint.f(*ing)
}

// Message is a description of the lint
func (lint IngressLint) Message() string {
	return lint.message
}

// Link is a URL to the issue or PR explaining the lint
func (lint IngressLint) Link() string {
	if lint.issue > 0 {
		return fmt.Sprintf("%v%v", util.IssuePrefix, lint.issue)
	}

	return ""
}

// Version is the ingress-nginx version the lint was added for, or the empty string
func (lint IngressLint) Version() string {
	return lint.version
}

// GetIngressLints returns all of the lints for ingresses
func GetIngressLints() []IngressLint {
	return []IngressLint{
		removedAnnotation("secure-backends", 3203, "0.21.0"),
		removedAnnotation("grpc-backend", 3203, "0.21.0"),
		removedAnnotation("add-base-url", 3174, "0.22.0"),
		removedAnnotation("base-url-scheme", 3174, "0.22.0"),
		removedAnnotation("session-cookie-hash", 3743, "0.24.0"),
		removedAnnotation("mirror-uri", 5015, "0.28.1"),
		{
			message: "The rewrite-target annotation value does not reference a capture group",
			issue:   3174,
			version: "0.22.0",
			f:       rewriteTargetWithoutCaptureGroup,
		},
		{
			message: "Contains an annotation with the prefix 'nginx.org'. This is a prefix for https://github.com/nginxinc/kubernetes-ingress",
			f:       annotationPrefixIsNginxOrg,
		},
		{
			message: "Contains an annotation with the prefix 'nginx.com'. This is a prefix for https://github.com/nginxinc/kubernetes-ingress",
			f:       annotationPrefixIsNginxCom,
		},
		{
			message: "The x-forwarded-prefix annotation value is a boolean instead of a string",
			issue:   3786,
			version: "0.24.0",
			f:       xForwardedPrefixIsBool,
		},
		{
			message: "Contains an configuration-snippet that contains a Satisfy directive.\nPlease use https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#satisfy",
			f:       satisfyDirective,
		},
	}
}

func xForwardedPrefixIsBool(ing networking.Ingress) bool {
	for name, val := range ing.Annotations {
		if strings.HasSuffix(name, "/x-forwarded-prefix") && (val == "true" || val == "false") {
			return true
		}
	}
	return false
}

func annotationPrefixIsNginxCom(ing networking.Ingress) bool {
	for name := range ing.Annotations {
		if strings.HasPrefix(name, "nginx.com/") {
			return true
		}
	}
	return false
}

func annotationPrefixIsNginxOrg(ing networking.Ingress) bool {
	for name := range ing.Annotations {
		if strings.HasPrefix(name, "nginx.org/") {
			return true
		}
	}
	return false
}

func rewriteTargetWithoutCaptureGroup(ing networking.Ingress) bool {
	for name, val := range ing.Annotations {
		if strings.HasSuffix(name, "/rewrite-target") && !strings.Contains(val, "$1") {
			return true
		}
	}
	return false
}

func removedAnnotation(annotationName string, issueNumber int, version string) IngressLint {
	return IngressLint{
		message: fmt.Sprintf("Contains the removed %v annotation.", annotationName),
		issue:   issueNumber,
		version: version,
		f: func(ing networking.Ingress) bool {
			for annotation := range ing.Annotations {
				if strings.HasSuffix(annotation, "/"+annotationName) {
					return true
				}
			}
			return false
		},
	}
}

func satisfyDirective(ing networking.Ingress) bool {
	for name, val := range ing.Annotations {
		if strings.HasSuffix(name, "/configuration-snippet") {
			return strings.Contains(val, "satisfy")
		}
	}

	return false
}
