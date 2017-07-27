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
	compute "google.golang.org/api/compute/v1"
	api_v1 "k8s.io/api/core/v1"
)

// ProbeProvider retrieves a probe struct given a nodePort
type probeProvider interface {
	GetProbe(sp ServicePort) (*api_v1.Probe, error)
}

// BackendPool is an interface to manage a pool of kubernetes nodePort services
// as gce backendServices, and sync them through the BackendServices interface.
type BackendPool interface {
	Init(p probeProvider)
	Add(port ServicePort) error
	Get(port int64) (*compute.BackendService, error)
	Delete(port int64) error
	Sync(ports []ServicePort) error
	GC(ports []ServicePort) error
	Shutdown() error
	Status(name string) string
	List() ([]interface{}, error)
}

// BackendServices is an interface for managing gce backend services.
type BackendServices interface {
	GetGlobalBackendService(name string) (*compute.BackendService, error)
	UpdateGlobalBackendService(bg *compute.BackendService) error
	CreateGlobalBackendService(bg *compute.BackendService) error
	DeleteGlobalBackendService(name string) error
	ListGlobalBackendServices() (*compute.BackendServiceList, error)
	GetGlobalBackendServiceHealth(name, instanceGroupLink string) (*compute.BackendServiceGroupHealth, error)
}
