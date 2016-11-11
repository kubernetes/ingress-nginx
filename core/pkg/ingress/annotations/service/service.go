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

package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/kubernetes/pkg/api"

	"github.com/golang/glog"
)

const (
	// NamedPortAnnotation annotation used to map named port in services
	NamedPortAnnotation = "ingress.kubernetes.io/named-ports"
)

type namedPortMapping map[string]string

// getPort returns the port defined in a named port
func (npm namedPortMapping) getPort(name string) (string, bool) {
	val, ok := npm.getPortMappings()[name]
	return val, ok
}

// getPortMappings returns a map containing the mapping of named ports names and number
func (npm namedPortMapping) getPortMappings() map[string]string {
	data := npm[NamedPortAnnotation]
	var mapping map[string]string
	if data == "" {
		return mapping
	}
	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		glog.Errorf("unexpected error reading annotations: %v", err)
	}

	return mapping
}

// GetPortMapping returns the number of the named port or an error if is not valid
func GetPortMapping(name string, s *api.Service) (int32, error) {
	if s == nil {
		return -1, fmt.Errorf("impossible to extract por mapping from %v (missing service)", name)
	}
	namedPorts := s.ObjectMeta.Annotations
	val, ok := namedPortMapping(namedPorts).getPort(name)
	if ok {
		port, err := strconv.Atoi(val)
		if err != nil {
			return -1, fmt.Errorf("service %v contains an invalid port mapping for %v (%v), %v", s.Name, name, val, err)
		}

		return int32(port), nil
	}

	return -1, fmt.Errorf("there is no port with name %v", name)
}
