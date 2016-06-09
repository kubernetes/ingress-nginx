/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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
	"k8s.io/contrib/ingress/controllers/gce/utils"
)

// NewFakeBackendServices creates a new fake backend services manager.
func NewFakeBackendServices() *FakeBackendServices {
	return &FakeBackendServices{
		backendServices: []*compute.BackendService{},
	}
}

// FakeBackendServices fakes out GCE backend services.
type FakeBackendServices struct {
	backendServices []*compute.BackendService
	calls           []int
}

// GetBackendService fakes getting a backend service from the cloud.
func (f *FakeBackendServices) GetBackendService(name string) (*compute.BackendService, error) {
	f.calls = append(f.calls, utils.Get)
	for i := range f.backendServices {
		if name == f.backendServices[i].Name {
			return f.backendServices[i], nil
		}
	}
	return nil, fmt.Errorf("Backend service %v not found", name)
}

// CreateBackendService fakes backend service creation.
func (f *FakeBackendServices) CreateBackendService(be *compute.BackendService) error {
	f.calls = append(f.calls, utils.Create)
	be.SelfLink = be.Name
	f.backendServices = append(f.backendServices, be)
	return nil
}

// DeleteBackendService fakes backend service deletion.
func (f *FakeBackendServices) DeleteBackendService(name string) error {
	f.calls = append(f.calls, utils.Delete)
	newBackends := []*compute.BackendService{}
	for i := range f.backendServices {
		if name != f.backendServices[i].Name {
			newBackends = append(newBackends, f.backendServices[i])
		}
	}
	f.backendServices = newBackends
	return nil
}

// ListBackendServices fakes backend service listing.
func (f *FakeBackendServices) ListBackendServices() (*compute.BackendServiceList, error) {
	return &compute.BackendServiceList{Items: f.backendServices}, nil
}

// UpdateBackendService fakes updating a backend service.
func (f *FakeBackendServices) UpdateBackendService(be *compute.BackendService) error {
	f.calls = append(f.calls, utils.Update)
	for i := range f.backendServices {
		if f.backendServices[i].Name == be.Name {
			f.backendServices[i] = be
		}
	}
	return nil
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
