/*
Copyright 2018 The Kubernetes Authors.

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

package metric

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

// NewDummyCollector returns a dummy metric collector
func NewDummyCollector() Collector {
	return &DummyCollector{}
}

// DummyCollector dummy implementation for mocks in tests
type DummyCollector struct{}

// ConfigSuccess ...
func (dc DummyCollector) ConfigSuccess(uint64, bool) {}

// SetAdmissionMetrics ...
func (dc DummyCollector) SetAdmissionMetrics(float64, float64, float64, float64, float64, float64) {}

// IncReloadCount ...
func (dc DummyCollector) IncReloadCount() {}

// IncReloadErrorCount ...
func (dc DummyCollector) IncReloadErrorCount() {}

// IncOrphanIngress ...
func (dc DummyCollector) IncOrphanIngress(string, string, string) {}

// DecOrphanIngress ...
func (dc DummyCollector) DecOrphanIngress(string, string, string) {}

// IncCheckCount ...
func (dc DummyCollector) IncCheckCount(string, string) {}

// IncCheckErrorCount ...
func (dc DummyCollector) IncCheckErrorCount(string, string) {}

// RemoveMetrics ...
func (dc DummyCollector) RemoveMetrics(ingresses, endpoints, certificates []string) {}

// Start ...
func (dc DummyCollector) Start(admissionStatus string) {}

// Stop ...
func (dc DummyCollector) Stop(admissionStatus string) {}

// SetSSLInfo ...
func (dc DummyCollector) SetSSLInfo([]*ingress.Server) {}

// SetSSLExpireTime ...
func (dc DummyCollector) SetSSLExpireTime([]*ingress.Server) {}

// SetHosts ...
func (dc DummyCollector) SetHosts(hosts sets.Set[string]) {}

// OnStartedLeading indicates the pod is not the current leader
func (dc DummyCollector) OnStartedLeading(electionID string) {}

// OnStoppedLeading indicates the pod is not the current leader
func (dc DummyCollector) OnStoppedLeading(electionID string) {}
