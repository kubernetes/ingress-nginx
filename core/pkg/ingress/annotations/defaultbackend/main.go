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

package defaultbackend

import (
	"fmt"

	"github.com/pkg/errors"
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	defaultBackend = "ingress.kubernetes.io/default-backend"
)

type backend struct {
	serviceResolver resolver.Service
}

// NewParser creates a new default backend annotation parser
func NewParser(sr resolver.Service) parser.IngressAnnotation {
	return backend{sr}
}

// Parse parses the annotations contained in the ingress to use
// a custom default backend
func (db backend) Parse(ing *extensions.Ingress) (interface{}, error) {
	s, err := parser.GetStringAnnotation(defaultBackend, ing)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%v/%v", ing.Namespace, s)
	svc, err := db.serviceResolver.GetService(name)
	if err != nil {
		return nil, errors.Wrapf(err, "unexpected error reading service %v", name)
	}

	return svc, nil
}
