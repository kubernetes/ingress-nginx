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
	computealpha "google.golang.org/api/compute/v0.alpha"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func fakeNotFoundErr() *googleapi.Error {
	return &googleapi.Error{Code: 404}
}

// NewFakeHealthCheckProvider returns a new FakeHealthChecks.
func NewFakeHealthCheckProvider() *FakeHealthCheckProvider {
	return &FakeHealthCheckProvider{
		http:    make(map[string]compute.HttpHealthCheck),
		generic: make(map[string]compute.HealthCheck),
	}
}

// FakeHealthCheckProvider fakes out health checks.
type FakeHealthCheckProvider struct {
	http    map[string]compute.HttpHealthCheck
	generic map[string]compute.HealthCheck
}

// CreateHttpHealthCheck fakes out http health check creation.
func (f *FakeHealthCheckProvider) CreateHttpHealthCheck(hc *compute.HttpHealthCheck) error {
	v := *hc
	v.SelfLink = "https://fake.google.com/compute/httpHealthChecks/" + hc.Name
	f.http[hc.Name] = v
	return nil
}

// GetHttpHealthCheck fakes out getting a http health check from the cloud.
func (f *FakeHealthCheckProvider) GetHttpHealthCheck(name string) (*compute.HttpHealthCheck, error) {
	if hc, found := f.http[name]; found {
		return &hc, nil
	}

	return nil, fakeNotFoundErr()
}

// DeleteHttpHealthCheck fakes out deleting a http health check.
func (f *FakeHealthCheckProvider) DeleteHttpHealthCheck(name string) error {
	if _, exists := f.http[name]; !exists {
		return fakeNotFoundErr()
	}

	delete(f.http, name)
	return nil
}

// UpdateHttpHealthCheck sends the given health check as an update.
func (f *FakeHealthCheckProvider) UpdateHttpHealthCheck(hc *compute.HttpHealthCheck) error {
	if _, exists := f.http[hc.Name]; !exists {
		return fakeNotFoundErr()
	}

	f.http[hc.Name] = *hc
	return nil
}

// CreateHealthCheck fakes out http health check creation.
func (f *FakeHealthCheckProvider) CreateHealthCheck(hc *compute.HealthCheck) error {
	v := *hc
	v.SelfLink = "https://fake.google.com/compute/healthChecks/" + hc.Name
	f.generic[hc.Name] = v
	return nil
}

// CreateHealthCheck fakes out http health check creation.
func (f *FakeHealthCheckProvider) CreateAlphaHealthCheck(hc *computealpha.HealthCheck) error {
	v := *hc
	v.SelfLink = "https://fake.google.com/compute/healthChecks/" + hc.Name
	f.generic[hc.Name] = toV1HealthCheck(hc)
	return nil
}

// GetHealthCheck fakes out getting a http health check from the cloud.
func (f *FakeHealthCheckProvider) GetHealthCheck(name string) (*compute.HealthCheck, error) {
	if hc, found := f.generic[name]; found {
		return &hc, nil
	}

	return nil, fakeNotFoundErr()
}

// GetHealthCheck fakes out getting a http health check from the cloud.
func (f *FakeHealthCheckProvider) GetAlphaHealthCheck(name string) (*computealpha.HealthCheck, error) {
	if hc, found := f.generic[name]; found {
		return toAlphaHealthCheck(&hc), nil
	}

	return nil, fakeNotFoundErr()
}

// DeleteHealthCheck fakes out deleting a http health check.
func (f *FakeHealthCheckProvider) DeleteHealthCheck(name string) error {
	if _, exists := f.generic[name]; !exists {
		return fakeNotFoundErr()
	}

	delete(f.generic, name)
	return nil
}

// UpdateHealthCheck sends the given health check as an update.
func (f *FakeHealthCheckProvider) UpdateHealthCheck(hc *compute.HealthCheck) error {
	if _, exists := f.generic[hc.Name]; !exists {
		return fakeNotFoundErr()
	}

	f.generic[hc.Name] = *hc
	return nil
}

func (f *FakeHealthCheckProvider) UpdateAlphaHealthCheck(hc *computealpha.HealthCheck) error {
	if _, exists := f.generic[hc.Name]; !exists {
		return fakeNotFoundErr()
	}

	f.generic[hc.Name] = toV1HealthCheck(hc)
	return nil
}

func toV1HealthCheck(hc *computealpha.HealthCheck) compute.HealthCheck {
	v1hc := compute.HealthCheck{
		Name:               hc.Name,
		Description:        hc.Description,
		CheckIntervalSec:   hc.CheckIntervalSec,
		HealthyThreshold:   hc.HealthyThreshold,
		UnhealthyThreshold: hc.UnhealthyThreshold,
		TimeoutSec:         hc.TimeoutSec,
		Type:               hc.Type,
		SelfLink:           hc.SelfLink,
	}

	if hc.HttpHealthCheck != nil {
		v1hc.HttpHealthCheck = &compute.HTTPHealthCheck{
			Port:        hc.HttpHealthCheck.Port,
			RequestPath: hc.HttpHealthCheck.RequestPath,
		}
	}

	if hc.HttpsHealthCheck != nil {
		v1hc.HttpsHealthCheck = &compute.HTTPSHealthCheck{
			Port:        hc.HttpsHealthCheck.Port,
			RequestPath: hc.HttpsHealthCheck.RequestPath,
		}
	}
	return v1hc
}

func toAlphaHealthCheck(hc *compute.HealthCheck) *computealpha.HealthCheck {
	alphahc := computealpha.HealthCheck{
		Name:               hc.Name,
		Description:        hc.Description,
		CheckIntervalSec:   hc.CheckIntervalSec,
		HealthyThreshold:   hc.HealthyThreshold,
		UnhealthyThreshold: hc.UnhealthyThreshold,
		TimeoutSec:         hc.TimeoutSec,
		Type:               hc.Type,
		SelfLink:           hc.SelfLink,
	}

	if hc.HttpHealthCheck != nil {
		alphahc.HttpHealthCheck = &computealpha.HTTPHealthCheck{
			Port:        hc.HttpHealthCheck.Port,
			RequestPath: hc.HttpHealthCheck.RequestPath,
		}
	}

	if hc.HttpsHealthCheck != nil {
		alphahc.HttpsHealthCheck = &computealpha.HTTPSHealthCheck{
			Port:        hc.HttpsHealthCheck.Port,
			RequestPath: hc.HttpsHealthCheck.RequestPath,
		}
	}
	return &alphahc
}
