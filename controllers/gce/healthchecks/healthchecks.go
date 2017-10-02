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
	"net/http"
	"time"

	computealpha "google.golang.org/api/compute/v0.alpha"
	compute "google.golang.org/api/compute/v1"

	"github.com/golang/glog"

	"k8s.io/ingress/controllers/gce/utils"
)

const (
	// These values set a low health threshold and a high failure threshold.
	// We're just trying to detect if the node networking is
	// borked, service level outages will get detected sooner
	// by kube-proxy.
	// DefaultHealthCheckInterval defines how frequently a probe runs with IG backends
	DefaultHealthCheckInterval = 60 * time.Second
	// DefaultNEGHealthCheckInterval defines how frequently a probe runs with NEG backends
	DefaultNEGHealthCheckInterval = 15 * time.Second
	// DefaultHealthyThreshold defines the threshold of success probes that declare a backend "healthy"
	DefaultHealthyThreshold = 1
	// DefaultUnhealthyThreshold defines the threshold of failure probes that declare a instance "unhealthy"
	DefaultUnhealthyThreshold = 10
	// DefaultNEGUnhealthyThreshold defines the threshold of failure probes that declare a network endpoint "unhealthy"
	DefaultNEGUnhealthyThreshold = 2
	// DefaultTimeout defines the timeout of each probe
	DefaultTimeout = 60 * time.Second

	//USE_SERVING_PORT: For NetworkEndpointGroup, the port specified for
	// each network endpoint is used for health checking. For other
	// backends, the port or named port specified in the Backend Service is
	// used for health checking.
	UseServingPortSpecification = "USE_SERVING_PORT"
)

// HealthChecks manages health checks.
type HealthChecks struct {
	cloud       HealthCheckProvider
	defaultPath string
	namer       *utils.Namer
}

// NewHealthChecker creates a new health checker.
// cloud: the cloud object implementing SingleHealthCheck.
// defaultHealthCheckPath: is the HTTP path to use for health checks.
func NewHealthChecker(cloud HealthCheckProvider, defaultHealthCheckPath string, namer *utils.Namer) HealthChecker {
	return &HealthChecks{cloud, defaultHealthCheckPath, namer}
}

// New returns a *HealthCheck with default settings and specified port/protocol
func (h *HealthChecks) New(port int64, protocol utils.AppProtocol, enableNEG bool) *HealthCheck {
	var hc *HealthCheck
	name := h.namer.BeName(port)
	if enableNEG {
		hc = DefaultNEGHealthCheck(protocol)
		hc.alphaHealthCheck.Name = name
	} else {
		hc = DefaultHealthCheck(port, protocol)
		hc.Name = name
	}
	// port is the key for retriving existing health-check
	// TODO: rename backend-service and health-check to not use port as key
	hc.Port = port
	return hc
}

// Sync retrieves a health check based on port, checks type and settings and updates/creates if necessary.
// Sync is only called by the backends.Add func - it's not a pool like other resources.
func (h *HealthChecks) Sync(hc *HealthCheck) (string, error) {
	// Verify default path
	if hc.RequestPath == "" {
		hc.RequestPath = h.defaultPath
	}

	existingHC, err := h.Get(hc.Port, hc.alphaHealthCheck != nil)
	if err != nil {
		if !utils.IsHTTPErrorCode(err, http.StatusNotFound) {
			return "", err
		}

		glog.V(2).Infof("Creating health check for port %v with protocol %v", hc.Port, hc.Type)
		if err = h.create(hc); err != nil {
			return "", err
		}

		return h.getHealthCheckLink(hc.Port)
	}

	if needToUpdate(existingHC, hc) {
		err = h.update(hc)
		return existingHC.SelfLink, err
	}

	if existingHC.RequestPath != hc.RequestPath {
		// TODO: reconcile health checks, and compare headers interval etc.
		// Currently Ingress doesn't expose all the health check params
		// natively, so some users prefer to hand modify the check.
		glog.V(2).Infof("Unexpected request path on health check %v, has %v want %v, NOT reconciling", hc.Name, existingHC.RequestPath, hc.RequestPath)
	} else {
		glog.V(2).Infof("Health check %v already exists and has the expected path %v", hc.Name, hc.RequestPath)
	}

	return existingHC.SelfLink, nil
}

func (h *HealthChecks) create(hc *HealthCheck) error {
	if hc.alphaHealthCheck == nil {
		return h.cloud.CreateHealthCheck(hc.ToComputeHealthCheck())
	} else {
		return h.cloud.CreateAlphaHealthCheck(hc.ToAlphaComputeHealthCheck())
	}
}

func (h *HealthChecks) update(hc *HealthCheck) error {
	if hc.alphaHealthCheck == nil {
		return h.cloud.UpdateHealthCheck(hc.ToComputeHealthCheck())
	} else {
		return h.cloud.UpdateAlphaHealthCheck(hc.ToAlphaComputeHealthCheck())
	}
}

func (h *HealthChecks) getHealthCheckLink(port int64) (string, error) {
	hc, err := h.Get(port, false)
	if err != nil {
		return "", err
	}
	return hc.SelfLink, nil
}

// Delete deletes the health check by port.
func (h *HealthChecks) Delete(port int64) error {
	name := h.namer.BeName(port)
	glog.V(2).Infof("Deleting health check %v", name)
	return h.cloud.DeleteHealthCheck(name)
}

// Get returns the health check by port
func (h *HealthChecks) Get(port int64, alpha bool) (*HealthCheck, error) {
	name := h.namer.BeName(port)
	if alpha {
		var ret *HealthCheck
		hc, err := h.cloud.GetAlphaHealthCheck(name)
		if err == nil {
			ret = &HealthCheck{alphaHealthCheck: hc}
			// SelfLink is used in return value
			ret.SelfLink = hc.SelfLink
		}
		return ret, err
	} else {
		hc, err := h.cloud.GetHealthCheck(name)
		return NewHealthCheck(hc), err
	}
}

// GetLegacy deletes legacy HTTP health checks
func (h *HealthChecks) GetLegacy(port int64) (*compute.HttpHealthCheck, error) {
	name := h.namer.BeName(port)
	return h.cloud.GetHttpHealthCheck(name)
}

// DeleteLegacy deletes legacy HTTP health checks
func (h *HealthChecks) DeleteLegacy(port int64) error {
	name := h.namer.BeName(port)
	glog.V(2).Infof("Deleting legacy HTTP health check %v", name)
	return h.cloud.DeleteHttpHealthCheck(name)
}

// DefaultHealthCheck simply returns the default health check.
func DefaultHealthCheck(port int64, protocol utils.AppProtocol) *HealthCheck {
	httpSettings := compute.HTTPHealthCheck{
		Port: port,
		// Empty string is used as a signal to the caller to use the appropriate
		// default.
		RequestPath: "",
	}

	hcSettings := compute.HealthCheck{
		// How often to health check.
		CheckIntervalSec: int64(DefaultHealthCheckInterval.Seconds()),
		// How long to wait before claiming failure of a health check.
		TimeoutSec: int64(DefaultTimeout.Seconds()),
		// Number of healthchecks to pass for a vm to be deemed healthy.
		HealthyThreshold: DefaultHealthyThreshold,
		// Number of healthchecks to fail before the vm is deemed unhealthy.
		UnhealthyThreshold: DefaultUnhealthyThreshold,
		Description:        "Default kubernetes L7 Loadbalancing health check.",
		Type:               string(protocol),
	}

	return &HealthCheck{
		HTTPHealthCheck: httpSettings,
		HealthCheck:     hcSettings,
	}
}

// DefaultHealthCheck simply returns the default health check.
func DefaultNEGHealthCheck(protocol utils.AppProtocol) *HealthCheck {
	hc := computealpha.HealthCheck{
		// How often to health check.
		CheckIntervalSec: int64(DefaultNEGHealthCheckInterval.Seconds()),
		// How long to wait before claiming failure of a health check.
		TimeoutSec: int64(DefaultNEGHealthCheckInterval.Seconds()),
		// Number of healthchecks to pass for a vm to be deemed healthy.
		HealthyThreshold: DefaultHealthyThreshold,
		// Number of healthchecks to fail before the vm is deemed unhealthy.
		UnhealthyThreshold: DefaultNEGUnhealthyThreshold,
		Description:        "Default kubernetes L7 Loadbalancing health check for NEG.",
		Type:               string(protocol),
	}
	if protocol == utils.ProtocolHTTP {
		hc.HttpHealthCheck = &computealpha.HTTPHealthCheck{
			PortSpecification: UseServingPortSpecification,
		}
	}
	if protocol == utils.ProtocolHTTPS {
		hc.HttpsHealthCheck = &computealpha.HTTPSHealthCheck{
			PortSpecification: UseServingPortSpecification,
		}
	}
	return &HealthCheck{
		alphaHealthCheck: &hc,
	}
}

// HealthCheck embeds two types - the generic healthcheck compute.HealthCheck
// and the HTTP settings compute.HTTPHealthCheck. By embedding both, consumers can modify
// all relevant settings (HTTP specific and HealthCheck generic) regardless of Type
// Consumers should call .Out() func to generate a compute.HealthCheck
// with the proper child struct (.HttpHealthCheck, .HttpshealthCheck, etc).
type HealthCheck struct {
	compute.HTTPHealthCheck
	compute.HealthCheck
	alphaHealthCheck *computealpha.HealthCheck
}

// NewHealthCheck creates a HealthCheck which abstracts nested structs away
func NewHealthCheck(hc *compute.HealthCheck) *HealthCheck {
	if hc == nil {
		return nil
	}

	v := &HealthCheck{HealthCheck: *hc}
	switch utils.AppProtocol(hc.Type) {
	case utils.ProtocolHTTP:
		v.HTTPHealthCheck = *hc.HttpHealthCheck
	case utils.ProtocolHTTPS:
		// HTTPHealthCheck and HTTPSHealthChecks have identical fields
		v.HTTPHealthCheck = compute.HTTPHealthCheck(*hc.HttpsHealthCheck)
	}

	// Users should be modifying HTTP(S) specific settings on the embedded
	// HTTPHealthCheck. Setting these to nil for preventing confusion.
	v.HealthCheck.HttpHealthCheck = nil
	v.HealthCheck.HttpsHealthCheck = nil

	return v
}

// Protocol returns the type cased to AppProtocol
func (hc *HealthCheck) Protocol() utils.AppProtocol {
	return utils.AppProtocol(hc.Type)
}

// ToComputeHealthCheck returns a valid compute.HealthCheck object
func (hc *HealthCheck) ToComputeHealthCheck() *compute.HealthCheck {
	// Zeroing out child settings as a precaution. GoogleAPI throws an error
	// if the wrong child struct is set.
	hc.HealthCheck.HttpsHealthCheck = nil
	hc.HealthCheck.HttpHealthCheck = nil

	switch hc.Protocol() {
	case utils.ProtocolHTTP:
		hc.HealthCheck.HttpHealthCheck = &hc.HTTPHealthCheck
	case utils.ProtocolHTTPS:
		https := compute.HTTPSHealthCheck(hc.HTTPHealthCheck)
		hc.HealthCheck.HttpsHealthCheck = &https
	}

	return &hc.HealthCheck
}

// ToComputeHealthCheck returns a valid compute.HealthCheck object
func (hc *HealthCheck) ToAlphaComputeHealthCheck() *computealpha.HealthCheck {
	return hc.alphaHealthCheck
}

func needToUpdate(old, new *HealthCheck) bool {
	if old.alphaHealthCheck != nil && new.alphaHealthCheck != nil {
		var oldPortSpec, newPortSpec string
		if old.alphaHealthCheck.HttpHealthCheck != nil {
			oldPortSpec = old.alphaHealthCheck.HttpHealthCheck.PortSpecification
		}
		if new.alphaHealthCheck.HttpHealthCheck != nil {
			newPortSpec = new.alphaHealthCheck.HttpHealthCheck.PortSpecification
		}
		if oldPortSpec != newPortSpec {
			glog.V(2).Infof("Updating health check %v because it has http port specification %q but need %q", old.Name, oldPortSpec, newPortSpec)
			return true
		}
		if old.alphaHealthCheck.HttpsHealthCheck != nil {
			oldPortSpec = old.alphaHealthCheck.HttpsHealthCheck.PortSpecification
		}
		if new.alphaHealthCheck.HttpsHealthCheck != nil {
			newPortSpec = new.alphaHealthCheck.HttpsHealthCheck.PortSpecification
		}
		if oldPortSpec != newPortSpec {
			glog.V(2).Infof("Updating health check %v because it has https port specification %q but need %q", old.Name, oldPortSpec, newPortSpec)
			return true
		}

	} else {
		if old.Protocol() != new.Protocol() {
			glog.V(2).Infof("Updating health check %v because it has protocol %v but need %v", old.Name, old.Type, new.Type)
			return true
		}
	}
	return false
}
