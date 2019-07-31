/*
Copyright 2018 The Kubernetes Authors.

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

package fastcgi

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type fastcgi struct {
	r resolver.Resolver
}

// Config describes the per location fastcgi config
type Config struct {
	Index  string            `json:"index"`
	Params map[string]string `json:"params"`
}

// Equal tests for equality between two Configuration types
func (l1 *Config) Equal(l2 *Config) bool {
	if l1 == l2 {
		return true
	}

	if l1 == nil || l2 == nil {
		return false
	}

	if l1.Index != l2.Index {
		return false
	}

	return reflect.DeepEqual(l1.Params, l2.Params)
}

// NewParser creates a new fastcgiConfig protocol annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return fastcgi{r}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to indicate the fastcgiConfig.
func (a fastcgi) Parse(ing *networking.Ingress) (interface{}, error) {

	fcgiConfig := Config{}

	if ing.GetAnnotations() == nil {
		return fcgiConfig, nil
	}

	index, err := parser.GetStringAnnotation("fastcgi-index", ing)
	if err != nil {
		index = ""
	}
	fcgiConfig.Index = index

	cm, err := parser.GetStringAnnotation("fastcgi-params-configmap", ing)
	if err != nil {
		return fcgiConfig, nil
	}

	cmns, cmn, err := cache.SplitMetaNamespaceKey(cm)
	if err != nil {
		return fcgiConfig, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "error reading configmap name from annotation"),
		}
	}

	if cmns == "" {
		cmns = ing.Namespace
	}

	cm = fmt.Sprintf("%v/%v", cmns, cmn)
	cmap, err := a.r.GetConfigMap(cm)
	if err != nil {
		return fcgiConfig, ing_errors.LocationDenied{
			Reason: errors.Wrapf(err, "unexpected error reading configmap %v", cm),
		}
	}

	fcgiConfig.Params = cmap.Data

	return fcgiConfig, nil
}
