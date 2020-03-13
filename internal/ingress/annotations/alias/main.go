/*
Copyright 2017 The Kubernetes Authors.

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

package alias

import (
	"sort"
	"strings"

	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type alias struct {
	r resolver.Resolver
}

// NewParser creates a new Alias annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return alias{r}
}

// Parse parses the annotations contained in the ingress rule
// used to add an alias to the provided hosts
func (a alias) Parse(ing *networking.Ingress) (interface{}, error) {
	val, err := parser.GetStringAnnotation("server-alias", ing)
	if err != nil {
		return []string{}, err
	}

	aliases := sets.NewString()
	for _, alias := range strings.Split(val, ",") {
		alias = strings.TrimSpace(alias)
		if len(alias) == 0 {
			continue
		}

		if !aliases.Has(alias) {
			aliases.Insert(alias)
		}
	}

	l := aliases.List()
	sort.Strings(l)

	return l, nil
}
