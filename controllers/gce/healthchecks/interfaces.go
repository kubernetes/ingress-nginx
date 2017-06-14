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

package healthchecks

import (
	compute "google.golang.org/api/compute/v1"

	"k8s.io/ingress/controllers/gce/utils"
)

// HealthCheckProvider is an interface to manage a single GCE health check.
type HealthCheckProvider interface {
	CreateHttpHealthCheck(hc *compute.HttpHealthCheck) error
	UpdateHttpHealthCheck(hc *compute.HttpHealthCheck) error
	DeleteHttpHealthCheck(name string) error
	GetHttpHealthCheck(name string) (*compute.HttpHealthCheck, error)

	CreateHealthCheck(hc *compute.HealthCheck) error
	UpdateHealthCheck(hc *compute.HealthCheck) error
	DeleteHealthCheck(name string) error
	GetHealthCheck(name string) (*compute.HealthCheck, error)
}

// HealthChecker is an interface to manage cloud HTTPHealthChecks.
type HealthChecker interface {
	New(port int64, protocol utils.AppProtocol) *HealthCheck
	Sync(hc *HealthCheck) (string, error)
	Delete(port int64) error
	Get(port int64) (*HealthCheck, error)
	GetLegacy(port int64) (*compute.HttpHealthCheck, error)
	DeleteLegacy(port int64) error
}
