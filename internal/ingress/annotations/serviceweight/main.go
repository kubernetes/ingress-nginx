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

package serviceweight

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"github.com/golang/glog"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"strings"
)

type serviceWeight struct {
	r resolver.Resolver
}

// Config returns service weight configuration for an Ingress rule
// that the key is service name, and the value is weight value.
type Config struct {
	ServiceWeight map[string]string
}

// Equal tests for equality between two ServiceWeight types
func (c1 *Config) Equal(c2 *Config) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	if len(c1.ServiceWeight) != len(c2.ServiceWeight) {
		return false
	}
	for k1, v1 := range c1.ServiceWeight {
		if v2, ok := c2.ServiceWeight[k1]; !ok || v1 != v2 {
			return false
		}
	}
	return true
}

// NewParser creates a new serviceWeight annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return serviceWeight{r}
}

func (s serviceWeight) Parse(ing *extensions.Ingress) (interface{}, error) {
	serviceWeight, err := parser.GetStringAnnotation("service-weight", ing)
	if err != nil {
		return nil, err
	}

	swMap := make(map[string]string)
	swList := strings.Split(serviceWeight, ",")
	for _, sw := range swList {
		sw = strings.TrimSpace(sw)
		if sw == "" {
			continue
		}
		pair := strings.Split(sw, ":")
		if len(pair) != 2 {
			glog.Errorf("invalid service weight format: %s", pair)
			continue
		}
		name := strings.TrimSpace(pair[0])
		weight := strings.TrimSpace(pair[1])
		if name == "" || weight == "" {
			glog.Errorf("invalid service weight format: %s", pair)
			continue
		}
		swMap[name] = weight
	}
	return &Config{ServiceWeight: swMap}, nil
}
