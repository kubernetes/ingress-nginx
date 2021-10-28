/*
Copyright 2021 The Kubernetes Authors.

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

package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

// AdmissionCollector stores prometheus metrics of the admission webhook
type AdmissionCollector struct {
	prometheus.Collector

	testedIngressLength prometheus.Gauge
	testedIngressTime   prometheus.Gauge

	renderingIngressLength prometheus.Gauge
	renderingIngressTime   prometheus.Gauge

	admissionTime prometheus.Gauge

	testedConfigurationSize prometheus.Gauge

	constLabels prometheus.Labels
	labels      prometheus.Labels
}

// NewAdmissionCollector creates a new AdmissionCollector instance for the admission collector
func NewAdmissionCollector(pod, namespace, class string) *AdmissionCollector {
	constLabels := prometheus.Labels{
		"controller_namespace": namespace,
		"controller_class":     class,
		"controller_pod":       pod,
	}

	am := &AdmissionCollector{
		constLabels: constLabels,

		labels: prometheus.Labels{
			"namespace": namespace,
			"class":     class,
		},

		testedIngressLength: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "admission_tested_ingresses",
				Help:        "The length of ingresses processed by the admission controller",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			}),
		testedIngressTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "admission_tested_duration",
				Help:        "The processing duration of the admission controller tests (float seconds)",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			}),
		renderingIngressLength: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "admission_render_ingresses",
				Help:        "The length of ingresses rendered by the admission controller",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			}),
		renderingIngressTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "admission_render_duration",
				Help:        "The processing duration of ingresses rendering by the admission controller (float seconds)",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			}),
		testedConfigurationSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   PrometheusNamespace,
				Name:        "admission_config_size",
				Help:        "The size of the tested configuration",
				ConstLabels: constLabels,
			}),
		admissionTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        "admission_roundtrip_duration",
				Help:        "The complete duration of the admission controller at the time to process a new event (float seconds)",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			}),
	}
	return am
}

// Describe implements prometheus.Collector
func (am AdmissionCollector) Describe(ch chan<- *prometheus.Desc) {
	am.testedIngressLength.Describe(ch)
	am.testedIngressTime.Describe(ch)
	am.renderingIngressLength.Describe(ch)
	am.renderingIngressTime.Describe(ch)
	am.testedConfigurationSize.Describe(ch)
	am.admissionTime.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (am AdmissionCollector) Collect(ch chan<- prometheus.Metric) {
	am.testedIngressLength.Collect(ch)
	am.testedIngressTime.Collect(ch)
	am.renderingIngressLength.Collect(ch)
	am.renderingIngressTime.Collect(ch)
	am.testedConfigurationSize.Collect(ch)
	am.admissionTime.Collect(ch)
}

// ByteFormat formats humanReadable bytes
func ByteFormat(bytes int64) string {
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB",
		float64(bytes)/float64(div), "kMGTPE"[exp])
}

// SetAdmissionMetrics sets the values for AdmissionMetrics that can be called externally
func (am *AdmissionCollector) SetAdmissionMetrics(testedIngressLength float64, testedIngressTime float64, renderingIngressLength float64, renderingIngressTime float64, testedConfigurationSize float64, admissionTime float64) {
	am.testedIngressLength.Set(testedIngressLength)
	am.testedIngressTime.Set(testedIngressTime)
	am.renderingIngressLength.Set(renderingIngressLength)
	am.renderingIngressTime.Set(renderingIngressTime)
	am.testedConfigurationSize.Set(testedConfigurationSize)
	am.admissionTime.Set(admissionTime)
	klog.Infof("processed ingress via admission controller {testedIngressLength:%v testedIngressTime:%vs renderingIngressLength:%v renderingIngressTime:%vs admissionTime:%vs testedConfigurationSize:%v}",
		testedIngressLength,
		testedIngressTime,
		renderingIngressLength,
		renderingIngressTime,
		ByteFormat(int64(testedConfigurationSize)),
		admissionTime,
	)
}
