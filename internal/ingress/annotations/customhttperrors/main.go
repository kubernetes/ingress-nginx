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

package customhttperrors

import (
	"strconv"
	"strings"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type customhttperrors struct {
	r resolver.Resolver
}

// NewParser creates a new custom http errors annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return customhttperrors{r}
}

// Parse parses the annotations contained in the ingress to use
// custom http errors
func (e customhttperrors) Parse(ing *networking.Ingress) (interface{}, error) {
	c, err := parser.GetStringAnnotation("custom-http-errors", ing)
	if err != nil {
		return nil, err
	}

	cSplit := strings.Split(c, ",")
	var codes []int
	for _, i := range cSplit {
		num, err := strconv.Atoi(i)
		if err != nil {
			return nil, err
		}
		codes = append(codes, num)
	}

	return codes, nil
}
