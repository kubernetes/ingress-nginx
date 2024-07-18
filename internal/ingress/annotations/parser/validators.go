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

package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	networking "k8s.io/api/networking/v1"
	machineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/net"
	"k8s.io/klog/v2"
)

type AnnotationValidator func(string) error

const (
	AnnotationRiskLow AnnotationRisk = iota
	AnnotationRiskMedium
	AnnotationRiskHigh
	AnnotationRiskCritical
)

var (
	alphaNumericChars    = `\-\.\_\~a-zA-Z0-9\/:`
	extendedAlphaNumeric = alphaNumericChars + ", "
	regexEnabledChars    = regexp.QuoteMeta(`^$[](){}*+?|&=\`)
	urlEnabledChars      = regexp.QuoteMeta(`,:?&=`)
)

// IsValidRegex checks if the tested string can be used as a regex, but without any weird character.
// It includes regex characters for paths that may contain regexes
var IsValidRegex = regexp.MustCompile("^[/" + alphaNumericChars + regexEnabledChars + "]*$")

// SizeRegex validates sizes understood by NGINX, like 1000, 100k, 1000M
var SizeRegex = regexp.MustCompile(`^(?i)\d+[bkmg]?$`)

// URLRegex is used to validate a URL but with only a specific set of characters:
// It is alphanumericChar + ":", "?", "&"
// A valid URL would be proto://something.com:port/something?arg=param
var (
	// URLIsValidRegex is used on full URLs, containing query strings (:, ? and &)
	URLIsValidRegex = regexp.MustCompile("^[" + alphaNumericChars + urlEnabledChars + "]*$")
	// BasicChars is alphanumeric and ".", "-", "_", "~" and ":", usually used on simple host:port/path composition.
	// This combination can also be used on fields that may contain characters like / (as ns/name)
	BasicCharsRegex = regexp.MustCompile("^[/" + alphaNumericChars + "]*$")
	// ExtendedChars is alphanumeric and ".", "-", "_", "~" and ":" plus "," and spaces, usually used on simple host:port/path composition
	ExtendedCharsRegex = regexp.MustCompile("^[/" + extendedAlphaNumeric + "]*$")
	// CharsWithSpace is like basic chars, but includes the space character
	CharsWithSpace = regexp.MustCompile("^[/" + alphaNumericChars + " ]*$")
	// NGINXVariable allows entries with alphanumeric characters, -, _ and the special "$"
	NGINXVariable = regexp.MustCompile(`^[A-Za-z0-9\-\_\$\{\}]*$`)
	// RegexPathWithCapture allows entries that SHOULD start with "/" and may contain alphanumeric + capture
	// character for regex based paths, like /something/$1/anything/$2
	RegexPathWithCapture = regexp.MustCompile(`^/?[` + alphaNumericChars + `\/\$]*$`)
	// HeadersVariable defines a regex that allows headers separated by comma
	HeadersVariable = regexp.MustCompile(`^[A-Za-z0-9-_, ]*$`)
	// URLWithNginxVariableRegex defines a url that can contain nginx variables.
	// It is a risky operation
	URLWithNginxVariableRegex = regexp.MustCompile("^[" + extendedAlphaNumeric + urlEnabledChars + "$]*$")
	// MaliciousRegex defines chars that are known to inject RCE
	MaliciousRegex = regexp.MustCompile(`\r|\n`)
)

// ValidateArrayOfServerName validates if all fields on a Server name annotation are
// regexes. They can be *.something*, ~^www\d+\.example\.com$ but not fancy character
func ValidateArrayOfServerName(value string) error {
	for _, fqdn := range strings.Split(value, ",") {
		if err := ValidateServerName(fqdn); err != nil {
			return err
		}
	}
	return nil
}

// ValidateServerName validates if the passed value is an acceptable server name. The server name
// can contain regex characters, as those are accepted values on nginx configuration
func ValidateServerName(value string) error {
	value = strings.TrimSpace(value)
	if !IsValidRegex.MatchString(value) {
		return fmt.Errorf("value %s is invalid server name", value)
	}
	return nil
}

// ValidateRegex receives a regex as an argument and uses it to validate
// the value of the field.
// Annotation can define if the spaces should be trimmed before validating the value
func ValidateRegex(regex *regexp.Regexp, removeSpace bool) AnnotationValidator {
	return func(s string) error {
		if removeSpace {
			s = strings.ReplaceAll(s, " ", "")
		}
		if !regex.MatchString(s) {
			return fmt.Errorf("value %s is invalid", s)
		}
		if MaliciousRegex.MatchString(s) {
			return fmt.Errorf("value %s contains malicious string", s)
		}

		return nil
	}
}

// CommonNameAnnotationValidator checks whether the annotation value starts with
// 'CN=' and is followed by a valid regex.
func CommonNameAnnotationValidator(s string) error {
	if !strings.HasPrefix(s, "CN=") {
		return fmt.Errorf("value %s is not a valid Common Name annotation: missing prefix 'CN='", s)
	}

	if _, err := regexp.Compile(s[3:]); err != nil {
		return fmt.Errorf("value %s is not a valid regex: %w", s, err)
	}

	return nil
}

// ValidateOptions receives an array of valid options that can be the value of annotation.
// If no valid option is found, it will return an error
func ValidateOptions(options []string, caseSensitive, trimSpace bool) AnnotationValidator {
	return func(s string) error {
		if trimSpace {
			s = strings.TrimSpace(s)
		}
		if !caseSensitive {
			s = strings.ToLower(s)
		}
		for _, option := range options {
			if s == option {
				return nil
			}
		}
		return fmt.Errorf("value does not match any valid option")
	}
}

// ValidateBool validates if the specified value is a bool
func ValidateBool(value string) error {
	_, err := strconv.ParseBool(value)
	return err
}

// ValidateInt validates if the specified value is an integer
func ValidateInt(value string) error {
	_, err := strconv.Atoi(value)
	return err
}

// ValidateCIDRs validates if the specified value is an array of IPs and CIDRs
func ValidateCIDRs(value string) error {
	_, err := net.ParseCIDRs(value)
	return err
}

// ValidateDuration validates if the specified value is a valid time
func ValidateDuration(value string) error {
	_, err := time.ParseDuration(value)
	return err
}

// ValidateNull always return null values and should not be widely used.
// It is used on the "snippet" annotations, as it is up to the admin to allow its
// usage, knowing it can be critical!
func ValidateNull(_ string) error {
	return nil
}

// ValidateServiceName validates if a provided service name is a valid string
func ValidateServiceName(value string) error {
	errs := machineryvalidation.NameIsDNS1035Label(value, false)
	if len(errs) != 0 {
		return fmt.Errorf("annotation does not contain a valid service name: %+v", errs)
	}
	return nil
}

// checkAnnotation will check each annotation for:
// 1 - Does it contain the internal validation and docs config?
// 2 - Does the ingress contains annotations? (validate null pointers)
// 3 - Does it contains a validator? Should it contain a validator (not containing is a bug!)
// 4 - Does the annotation contain aliases? So we should use if the alias is defined an the annotation not.
// 4 - Runs the validator on the value
// It will return the full annotation name if all is fine
func checkAnnotation(name string, ing *networking.Ingress, fields AnnotationFields) (string, error) {
	var validateFunc AnnotationValidator
	if fields != nil {
		config, ok := fields[name]
		if !ok {
			return "", fmt.Errorf("annotation does not contain a valid internal configuration, this is an Ingress Controller issue! Please raise an issue on github.com/kubernetes/ingress-nginx")
		}
		validateFunc = config.Validator
	}

	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return "", ing_errors.ErrMissingAnnotations
	}

	annotationFullName := GetAnnotationWithPrefix(name)
	if annotationFullName == "" {
		return "", ing_errors.ErrInvalidAnnotationName
	}

	annotationValue := ing.GetAnnotations()[annotationFullName]
	if fields != nil {
		if validateFunc == nil {
			return "", fmt.Errorf("annotation does not contain a validator. This is an ingress-controller bug. Please open an issue")
		}
		if annotationValue == "" {
			for _, annotationAlias := range fields[name].AnnotationAliases {
				tempAnnotationFullName := GetAnnotationWithPrefix(annotationAlias)
				if aliasVal := ing.GetAnnotations()[tempAnnotationFullName]; aliasVal != "" {
					annotationValue = aliasVal
					annotationFullName = tempAnnotationFullName
					break
				}
			}
		}
		// We don't run validation against empty values
		if EnableAnnotationValidation && annotationValue != "" {
			if err := validateFunc(annotationValue); err != nil {
				klog.Warningf("validation error on ingress %s/%s: annotation %s contains invalid value %s", ing.GetNamespace(), ing.GetName(), name, annotationValue)
				return "", ing_errors.NewValidationError(annotationFullName)
			}
		}
	}

	return annotationFullName, nil
}

func CheckAnnotationRisk(annotations map[string]string, maxrisk AnnotationRisk, config AnnotationFields) error {
	var err error
	for annotation := range annotations {
		annPure := TrimAnnotationPrefix(annotation)
		if cfg, ok := config[annPure]; ok && cfg.Risk > maxrisk {
			err = errors.Join(err, fmt.Errorf("annotation %s is too risky for environment", annotation))
		}
	}
	return err
}
