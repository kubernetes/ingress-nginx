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

package healthchecks

import (
	compute "google.golang.org/api/compute/v1"
)

// healthCheckGetter retrieves health checks.
type healthCheckGetter interface {
	// HealthCheck returns the HTTP readiness check for a node port.
	HealthCheck(nodePort int64) (*compute.HttpHealthCheck, error)
}

// SingleHealthCheck is an interface to manage a single GCE health check.
type SingleHealthCheck interface {
	CreateHttpHealthCheck(hc *compute.HttpHealthCheck) error
	UpdateHttpHealthCheck(hc *compute.HttpHealthCheck) error
	DeleteHttpHealthCheck(name string) error
	GetHttpHealthCheck(name string) (*compute.HttpHealthCheck, error)
}

// HealthChecker is an interface to manage cloud HTTPHealthChecks.
type HealthChecker interface {
	Init(h healthCheckGetter)

	Add(port int64) error
	Delete(port int64) error
	Get(port int64) (*compute.HttpHealthCheck, error)
}
