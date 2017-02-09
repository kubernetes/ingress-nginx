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

package backends

import (
	"fmt"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/client/cache"
)

// NewFakeBackendServices creates a new fake backend services manager.
func NewFakeBackendServices(ef func(op int, be *compute.BackendService) error) *FakeBackendServices {
	return &FakeBackendServices{
		errFunc: ef,
		backendServices: cache.NewStore(func(obj interface{}) (string, error) {
			svc := obj.(*compute.BackendService)
			return svc.Name, nil
		}),
	}
}

// FakeBackendServices fakes out GCE backend services.
type FakeBackendServices struct {
	backendServices cache.Store
	calls           []int
	errFunc         func(op int, be *compute.BackendService) error
}

// GetBackendService fakes getting a backend service from the cloud.
func (f *FakeBackendServices) GetBackendService(name string) (*compute.BackendService, error) {
	f.calls = append(f.calls, utils.Get)
	obj, exists, err := f.backendServices.GetByKey(name)
	if !exists {
		return nil, fmt.Errorf("Backend service %v not found", name)
	}
	if err != nil {
		return nil, err
	}

	svc := obj.(*compute.BackendService)
	if name == svc.Name {
		return svc, nil
	}
	return nil, fmt.Errorf("Backend service %v not found", name)
}

// CreateBackendService fakes backend service creation.
func (f *FakeBackendServices) CreateBackendService(be *compute.BackendService) error {
	if f.errFunc != nil {
		if err := f.errFunc(utils.Create, be); err != nil {
			return err
		}
	}
	f.calls = append(f.calls, utils.Create)
	be.SelfLink = be.Name
	return f.backendServices.Update(be)
}

// DeleteBackendService fakes backend service deletion.
func (f *FakeBackendServices) DeleteBackendService(name string) error {
	f.calls = append(f.calls, utils.Delete)
	svc, exists, err := f.backendServices.GetByKey(name)
	if !exists {
		return fmt.Errorf("Backend service %v not found", name)
	}
	if err != nil {
		return err
	}
	return f.backendServices.Delete(svc)
}

// ListBackendServices fakes backend service listing.
func (f *FakeBackendServices) ListBackendServices() (*compute.BackendServiceList, error) {
	var svcs []*compute.BackendService
	for _, s := range f.backendServices.List() {
		svc := s.(*compute.BackendService)
		svcs = append(svcs, svc)
	}
	return &compute.BackendServiceList{Items: svcs}, nil
}

// UpdateBackendService fakes updating a backend service.
func (f *FakeBackendServices) UpdateBackendService(be *compute.BackendService) error {
	f.calls = append(f.calls, utils.Update)
	return f.backendServices.Update(be)
}

// GetHealth fakes getting backend service health.
func (f *FakeBackendServices) GetHealth(name, instanceGroupLink string) (*compute.BackendServiceGroupHealth, error) {
	be, err := f.GetBackendService(name)
	if err != nil {
		return nil, err
	}
	states := []*compute.HealthStatus{
		{
			HealthState: "HEALTHY",
			IpAddress:   "",
			Port:        be.Port,
		},
	}
	return &compute.BackendServiceGroupHealth{
		HealthStatus: states}, nil
}
