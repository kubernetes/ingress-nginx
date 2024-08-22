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

// ConfigSuccess dummy implementation
func (dc DummyCollector) ConfigSuccess(uint64, bool) {}

// SetAdmissionMetrics dummy implementation
func (dc DummyCollector) SetAdmissionMetrics(float64, float64, float64, float64, float64, float64) {}

// IncReloadCount dummy implementation
func (dc DummyCollector) IncReloadCount() {}

// IncReloadErrorCount dummy implementation
func (dc DummyCollector) IncReloadErrorCount() {}

// IncOrphanIngress dummy implementation
func (dc DummyCollector) IncOrphanIngress(string, string, string) {}

// DecOrphanIngress dummy implementation
func (dc DummyCollector) DecOrphanIngress(string, string, string) {}

// IncCheckCount dummy implementation
func (dc DummyCollector) IncCheckCount(string, string) {}

// IncCheckErrorCount dummy implementation
func (dc DummyCollector) IncCheckErrorCount(string, string) {}

// RemoveMetrics dummy implementation
func (dc DummyCollector) RemoveMetrics(_, _ []string) {}

// Start dummy implementation
func (dc DummyCollector) Start(_ string) {}

// Stop dummy implementation
func (dc DummyCollector) Stop(_ string) {}

// SetSSLInfo dummy implementation
func (dc DummyCollector) SetSSLInfo([]*ingress.Server) {}

// SetSSLExpireTime dummy implementation
func (dc DummyCollector) SetSSLExpireTime([]*ingress.Server) {}

// SetHosts dummy implementation
func (dc DummyCollector) SetHosts(_ sets.Set[string]) {}

// OnStartedLeading indicates the pod is not the current leader
func (dc DummyCollector) OnStartedLeading(_ string) {}

// OnStoppedLeading indicates the pod is not the current leader
func (dc DummyCollector) OnStoppedLeading(_ string) {}
