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

// IngressLister makes a Store that lists Ingress.
type IngressLister struct {
	cache.Store
}

// SecretsLister makes a Store that lists Secrets.
type SecretLister struct {
	cache.Store
}

// GetByName searches for a secret in the local secrets Store
func (sl *SecretLister) GetByName(name string) (*apiv1.Secret, error) {
	s, exists, err := sl.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("secret %v was not found", name)
	}
	return s.(*apiv1.Secret), nil
}

// ConfigMapLister makes a Store that lists Configmaps.
type ConfigMapLister struct {
	cache.Store
}

// GetByName searches for a configmap in the local configmaps Store
func (cml *ConfigMapLister) GetByName(name string) (*apiv1.ConfigMap, error) {
	s, exists, err := cml.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("configmap %v was not found", name)
	}
	return s.(*apiv1.ConfigMap), nil
}

// ServiceLister makes a Store that lists Services.
type ServiceLister struct {
	cache.Store
}

// GetByName searches for a service in the local secrets Store
func (sl *ServiceLister) GetByName(name string) (*apiv1.Service, error) {
	s, exists, err := sl.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("service %v was not found", name)
	}
	return s.(*apiv1.Service), nil
}

// NodeLister makes a Store that lists Nodes.
type NodeLister struct {
	cache.Store
}

// EndpointLister makes a Store that lists Endpoints.
type EndpointLister struct {
	cache.Store
}

// GetServiceEndpoints returns the endpoints of a service, matched on service name.
func (s *EndpointLister) GetServiceEndpoints(svc *apiv1.Service) (ep apiv1.Endpoints, err error) {
	for _, m := range s.Store.List() {
		ep = *m.(*apiv1.Endpoints)
		if svc.Name == ep.Name && svc.Namespace == ep.Namespace {
			return ep, nil
		}
	}
	err = fmt.Errorf("could not find endpoints for service: %v", svc.Name)
	return
}
