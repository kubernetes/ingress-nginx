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
	compute "google.golang.org/api/compute/v1"
)

// BackendPool is an interface to manage a pool of kubernetes nodePort services
// as gce backendServices, and sync them through the BackendServices interface.
type BackendPool interface {
	Add(port int64) error
	Get(port int64) (*compute.BackendService, error)
	Delete(port int64) error
	Sync(ports []int64) error
	GC(ports []int64) error
	Shutdown() error
	Status(name string) string
	List() (*compute.BackendServiceList, error)
}

// BackendServices is an interface for managing gce backend services.
type BackendServices interface {
	GetBackendService(name string) (*compute.BackendService, error)
	UpdateBackendService(bg *compute.BackendService) error
	CreateBackendService(bg *compute.BackendService) error
	DeleteBackendService(name string) error
	ListBackendServices() (*compute.BackendServiceList, error)
	GetHealth(name, instanceGroupLink string) (*compute.BackendServiceGroupHealth, error)
}

// SingleHealthCheck is an interface to manage a single GCE health check.
type SingleHealthCheck interface {
	CreateHttpHealthCheck(hc *compute.HttpHealthCheck) error
	DeleteHttpHealthCheck(name string) error
	GetHttpHealthCheck(name string) (*compute.HttpHealthCheck, error)
}

// HealthChecker is an interface to manage cloud HTTPHealthChecks.
type HealthChecker interface {
	Add(port int64, path string) error
	Delete(port int64) error
	Get(port int64) (*compute.HttpHealthCheck, error)
}
