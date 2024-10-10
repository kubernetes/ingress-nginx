/*
Copyright 2015 The Kubernetes Authors.

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

package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress-nginx/internal/ingress/errors"
)

// DefaultAnnotationsPrefix defines the common prefix used in the nginx ingress controller
const (
	DefaultAnnotationsPrefix          = "nginx.ingress.kubernetes.io"
	DefaultEnableAnnotationValidation = true
)

var (
	// AnnotationsPrefix is the mutable attribute that the controller explicitly refers to
	AnnotationsPrefix = DefaultAnnotationsPrefix
	// Enable is the mutable attribute for enabling or disabling the validation functions
	EnableAnnotationValidation = DefaultEnableAnnotationValidation
)

// AnnotationGroup defines the group that this annotation may belong
// eg.: Security, Snippets, Rewrite, etc
type AnnotationGroup string

// AnnotationScope defines which scope this annotation applies. May be to the whole
// ingress, per location, etc
type AnnotationScope string

var (
	AnnotationScopeLocation AnnotationScope = "location"
	AnnotationScopeIngress  AnnotationScope = "ingress"
)

// AnnotationRisk is a subset of risk that an annotation may represent.
// Based on the Risk, the admin will be able to allow or disallow users to set it
// on their ingress objects
type AnnotationRisk int

type AnnotationFields map[string]AnnotationConfig

// AnnotationConfig defines the configuration that a single annotation field
// has, with the Validator and the documentation of this field.
type AnnotationConfig struct {
	// Validator defines a function to validate the annotation value
	Validator AnnotationValidator
	// Documentation defines a user facing documentation for this annotation. This
	// field will be used to auto generate documentations
	Documentation string
	// Risk defines a risk of this annotation being exposed to the user. Annotations
	// with bool fields, or to set timeout are usually low risk. Annotations that allows
	// string input without a limited set of options may represent a high risk
	Risk AnnotationRisk

	// Scope defines which scope this annotation applies, may be to location, to an Ingress object, etc
	Scope AnnotationScope

	// AnnotationAliases defines other names this annotation may have.
	AnnotationAliases []string
}

// Annotation defines an annotation feature an Ingress may have.
// It should contain the internal resolver, and all the annotations
// with configs and Validators that should be used for each Annotation
type Annotation struct {
	// Annotations contains all the annotations that belong to this feature
	Annotations AnnotationFields
	// Group defines which annotation group this feature belongs to
	Group AnnotationGroup
}

// IngressAnnotation has a method to parse annotations located in Ingress
type IngressAnnotation interface {
	Parse(ing *networking.Ingress) (interface{}, error)
	GetDocumentation() AnnotationFields
	Validate(anns map[string]string) error
}

type ingAnnotations map[string]string

// TODO: We already parse all of this on checkAnnotation and can just do a parse over the
// value
func (a ingAnnotations) parseBool(name string) (bool, error) {
	val, ok := a[name]
	if ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, errors.NewInvalidAnnotationContent(name, val)
		}
		return b, nil
	}
	return false, errors.ErrMissingAnnotations
}

func (a ingAnnotations) parseString(name string) (string, error) {
	val, ok := a[name]
	if ok {
		s := normalizeString(val)
		if s == "" {
			return "", errors.NewInvalidAnnotationContent(name, val)
		}

		return s, nil
	}
	return "", errors.ErrMissingAnnotations
}

func (a ingAnnotations) parseInt(name string) (int, error) {
	val, ok := a[name]
	if ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, errors.NewInvalidAnnotationContent(name, val)
		}
		return i, nil
	}
	return 0, errors.ErrMissingAnnotations
}

func (a ingAnnotations) parseFloat32(name string) (float32, error) {
	val, ok := a[name]
	if ok {
		i, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return 0, errors.NewInvalidAnnotationContent(name, val)
		}
		return float32(i), nil
	}
	return 0, errors.ErrMissingAnnotations
}

// GetBoolAnnotation extracts a boolean from an Ingress annotation
func GetBoolAnnotation(name string, ing *networking.Ingress, fields AnnotationFields) (bool, error) {
	v, err := checkAnnotation(name, ing, fields)
	if err != nil {
		return false, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseBool(v)
}

// GetStringAnnotation extracts a string from an Ingress annotation
func GetStringAnnotation(name string, ing *networking.Ingress, fields AnnotationFields) (string, error) {
	v, err := checkAnnotation(name, ing, fields)
	if err != nil {
		return "", err
	}

	return ingAnnotations(ing.GetAnnotations()).parseString(v)
}

// GetIntAnnotation extracts an int from an Ingress annotation
func GetIntAnnotation(name string, ing *networking.Ingress, fields AnnotationFields) (int, error) {
	v, err := checkAnnotation(name, ing, fields)
	if err != nil {
		return 0, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseInt(v)
}

// GetFloatAnnotation extracts a float32 from an Ingress annotation
func GetFloatAnnotation(name string, ing *networking.Ingress, fields AnnotationFields) (float32, error) {
	v, err := checkAnnotation(name, ing, fields)
	if err != nil {
		return 0, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseFloat32(v)
}

// GetAnnotationWithPrefix returns the prefix of ingress annotations
func GetAnnotationWithPrefix(suffix string) string {
	return fmt.Sprintf("%v/%v", AnnotationsPrefix, suffix)
}

func TrimAnnotationPrefix(annotation string) string {
	return strings.TrimPrefix(annotation, AnnotationsPrefix+"/")
}

func StringRiskToRisk(risk string) AnnotationRisk {
	switch strings.ToLower(risk) {
	case "critical":
		return AnnotationRiskCritical
	case "high":
		return AnnotationRiskHigh
	case "medium":
		return AnnotationRiskMedium
	default:
		return AnnotationRiskLow
	}
}

func (a AnnotationRisk) ToString() string {
	switch a {
	case AnnotationRiskCritical:
		return "Critical"
	case AnnotationRiskHigh:
		return "High"
	case AnnotationRiskMedium:
		return "Medium"
	case AnnotationRiskLow:
		return "Low"
	default:
		return "Unknown"
	}
}

func normalizeString(input string) string {
	trimmedContent := []string{}
	for _, line := range strings.Split(input, "\n") {
		trimmedContent = append(trimmedContent, strings.TrimSpace(line))
	}

	return strings.Join(trimmedContent, "\n")
}

var configmapAnnotations = sets.NewString(
	"auth-proxy-set-header",
	"fastcgi-params-configmap",
)

// AnnotationsReferencesConfigmap checks if at least one annotation in the Ingress rule
// references a configmap.
func AnnotationsReferencesConfigmap(ing *networking.Ingress) bool {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return false
	}

	for name := range ing.GetAnnotations() {
		if configmapAnnotations.Has(name) {
			return true
		}
	}

	return false
}

// StringToURL parses the provided string into URL and returns error
// message in case of failure
func StringToURL(input string) (*url.URL, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("%v is not a valid URL: %v", input, err)
	}

	switch {
	case parsedURL.Scheme == "":
		return nil, fmt.Errorf("url scheme is empty")
	case parsedURL.Host == "":
		return nil, fmt.Errorf("url host is empty")
	case strings.Contains(parsedURL.Host, ".."):
		return nil, fmt.Errorf("invalid url host")
	default:
		return parsedURL, nil
	}
}
