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

package store

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// EndpointLister makes a Store that lists Endpoints.
type EndpointLister struct {
	cache.Store
}

// GetServiceEndpoints returns the endpoints of a service, matched on service name.
func (s *EndpointLister) GetServiceEndpoints(svc *apiv1.Service) (*apiv1.Endpoints, error) {
	for _, m := range s.Store.List() {
		ep := m.(*apiv1.Endpoints)
		if svc.Name == ep.Name && svc.Namespace == ep.Namespace {
			return ep, nil
		}
	}
	return nil, fmt.Errorf("could not find endpoints for service: %v", svc.Name)
}
