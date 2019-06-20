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

// IngressAnnotationsLister makes a Store that lists annotations in Ingress rules.
type IngressAnnotationsLister struct {
	cache.Store
}

// SecretLister makes a Store that lists Secrets.
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

// SSLCertTracker holds a store of referenced Secrets in Ingress rules
type SSLCertTracker struct {
	cache.ThreadSafeStore
}

// NewSSLCertTracker creates a new SSLCertTracker store
func NewSSLCertTracker() *SSLCertTracker {
	return &SSLCertTracker{
		cache.NewThreadSafeStore(cache.Indexers{}, cache.Indices{}),
	}
}
