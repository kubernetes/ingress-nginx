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
	"errors"
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/apis/extensions"
)

var (
	// ErrMissingAnnotations is returned when the ingress rule
	// does not contains annotations related with rate limit
	ErrMissingAnnotations = errors.New("Ingress rule without annotations")

	// ErrInvalidName ...
	ErrInvalidName = errors.New("invalid annotation name")
)

type ingAnnotations map[string]string

func (a ingAnnotations) parseBool(name string) (bool, error) {
	val, ok := a[name]
	if ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b, nil
		}
	}
	return false, ErrMissingAnnotations
}

func (a ingAnnotations) parseString(name string) (string, error) {
	val, ok := a[name]
	if ok {
		return val, nil
	}
	return "", ErrMissingAnnotations
}

func (a ingAnnotations) parseInt(name string) (int, error) {
	val, ok := a[name]
	if ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("invalid annotations value: %v", err)
		}
		return i, nil
	}
	return 0, ErrMissingAnnotations
}

// GetBoolAnnotation ...
func GetBoolAnnotation(name string, ing *extensions.Ingress) (bool, error) {
	if ing == nil || ing.GetAnnotations() == nil {
		return false, ErrMissingAnnotations
	}
	if name == "" {
		return false, ErrInvalidName
	}

	return ingAnnotations(ing.GetAnnotations()).parseBool(name)
}

// GetStringAnnotation ...
func GetStringAnnotation(name string, ing *extensions.Ingress) (string, error) {
	if ing == nil || ing.GetAnnotations() == nil {
		return "", ErrMissingAnnotations
	}
	if name == "" {
		return "", ErrInvalidName
	}

	return ingAnnotations(ing.GetAnnotations()).parseString(name)
}

// GetIntAnnotation ...
func GetIntAnnotation(name string, ing *extensions.Ingress) (int, error) {
	if ing == nil || ing.GetAnnotations() == nil {
		return 0, ErrMissingAnnotations
	}
	if name == "" {
		return 0, ErrInvalidName
	}

	return ingAnnotations(ing.GetAnnotations()).parseInt(name)
}
