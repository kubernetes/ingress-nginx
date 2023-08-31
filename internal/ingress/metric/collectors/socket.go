/*
Copyright 2016 The Kubernetes Authors.

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
	"io"
	"net"
	"os"
	"strings"
	"syscall"

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

type socketData struct {
	Host   string `json:"host"`
	Status string `json:"status"`

	ResponseLength float64 `json:"responseLength"`

	Method string `json:"method"`

	RequestLength float64 `json:"requestLength"`
	RequestTime   float64 `json:"requestTime"`

	Latency      float64 `json:"upstreamLatency"`
	HeaderTime   float64 `json:"upstreamHeaderTime"`
	ResponseTime float64 `json:"upstreamResponseTime"`
	Namespace    string  `json:"namespace"`
	Ingress      string  `json:"ingress"`
	Service      string  `json:"service"`
	Canary       string  `json:"canary"`
	Path         string  `json:"path"`
}

// HistogramBuckets allow customizing prometheus histogram buckets values
type HistogramBuckets struct {
	TimeBuckets   []float64
	LengthBuckets []float64
	SizeBuckets   []float64
}

type metricMapping map[string]prometheus.Collector

// SocketCollector stores prometheus metrics and ingress meta-data
type SocketCollector struct {
	prometheus.Collector

	upstreamLatency *prometheus.SummaryVec // TODO: DEPRECATED, remove
	connectTime     *prometheus.HistogramVec
	headerTime      *prometheus.HistogramVec
	requestTime     *prometheus.HistogramVec
	responseTime    *prometheus.HistogramVec

	requestLength  *prometheus.HistogramVec
	responseLength *prometheus.HistogramVec
	bytesSent      *prometheus.HistogramVec // TODO: DEPRECATED, remove

	requests *prometheus.CounterVec

	listener net.Listener

	metricMapping metricMapping

	hosts sets.Set[string]

	metricsPerHost      bool
	reportStatusClasses bool
}

var requestTags = []string{
	"status",

	"method",
	"path",

	"namespace",
	"ingress",
	"service",
	"canary",
}

// DefObjectives was removed in https://github.com/prometheus/client_golang/pull/262
// updating the library to latest version changed the output of the metrics
var defObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}

// NewSocketCollector creates a new SocketCollector instance using
// the ingress watch namespace and class used by the controller
func NewSocketCollector(pod, namespace, class string, metricsPerHost, reportStatusClasses bool, buckets HistogramBuckets, excludeMetrics []string) (*SocketCollector, error) {
	socket := "/tmp/nginx/prometheus-nginx.socket"
	// unix sockets must be unlink()ed before being used
	//nolint:errcheck // Ignore unlink error
	_ = syscall.Unlink(socket)

	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}

	err = os.Chmod(socket, 0o777) // #nosec
	if err != nil {
		return nil, err
	}

	constLabels := prometheus.Labels{
		"controller_namespace": namespace,
		"controller_class":     class,
		"controller_pod":       pod,
	}

	requestTags := requestTags
	if metricsPerHost {
		requestTags = append(requestTags, "host")
	}

	em := make(map[string]struct{}, len(excludeMetrics))
	for _, m := range excludeMetrics {
		// remove potential nginx_ingress_controller prefix from the metric name
		// TBD: how to handle fully qualified histogram metrics e.g. _buckets and _sum. Should we just remove the suffix and remove the histogram metric or ignore it?
		em[strings.TrimPrefix(m, "nginx_ingress_controller_")] = struct{}{}
	}

	// create metric mapping with only the metrics that are not excluded
	mm := make(metricMapping)

	sc := &SocketCollector{
		listener: listener,

		metricsPerHost:      metricsPerHost,
		reportStatusClasses: reportStatusClasses,

		connectTime: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "connect_duration_seconds",
				Help:        "The time spent on establishing a connection with the upstream server",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Buckets:     buckets.TimeBuckets,
			},
			requestTags,
			em,
			mm,
		),

		headerTime: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "header_duration_seconds",
				Help:        "The time spent on receiving first header from the upstream server",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Buckets:     buckets.TimeBuckets,
			},
			requestTags,
			em,
			mm,
		),
		responseTime: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "response_duration_seconds",
				Help:        "The time spent on receiving the response from the upstream server",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Buckets:     buckets.TimeBuckets,
			},
			requestTags,
			em,
			mm,
		),

		requestTime: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "request_duration_seconds",
				Help:        "The request processing time in milliseconds",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Buckets:     buckets.TimeBuckets,
			},
			requestTags,
			em,
			mm,
		),

		responseLength: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "response_size",
				Help:        "The response length (including request line, header, and request body)",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Buckets:     buckets.LengthBuckets,
			},
			requestTags,
			em,
			mm,
		),

		requestLength: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "request_size",
				Help:        "The request length (including request line, header, and request body)",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Buckets:     buckets.LengthBuckets,
			},
			requestTags,
			em,
			mm,
		),

		requests: counterMetric(
			&prometheus.CounterOpts{
				Name:        "requests",
				Help:        "The total number of client requests",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
			},
			requestTags,
			em,
			mm,
		),

		bytesSent: histogramMetric(
			&prometheus.HistogramOpts{
				Name:        "bytes_sent",
				Help:        "DEPRECATED The number of bytes sent to a client",
				Namespace:   PrometheusNamespace,
				Buckets:     buckets.SizeBuckets,
				ConstLabels: constLabels,
			},
			requestTags,
			em,
			mm,
		),

		upstreamLatency: summaryMetric(
			&prometheus.SummaryOpts{
				Name:        "ingress_upstream_latency_seconds",
				Help:        "DEPRECATED Upstream service latency per Ingress",
				Namespace:   PrometheusNamespace,
				ConstLabels: constLabels,
				Objectives:  defObjectives,
			},
			[]string{"ingress", "namespace", "service", "canary"},
			em,
			mm,
		),
	}

	sc.metricMapping = mm
	return sc, nil
}

func containsMetric(excludeMetrics map[string]struct{}, name string) bool {
	if _, ok := excludeMetrics[name]; ok {
		klog.V(3).InfoS("Skipping metric", "metric", name)
		return true
	}
	return false
}

func summaryMetric(opts *prometheus.SummaryOpts, requestTags []string, excludeMetrics map[string]struct{}, metricMapping metricMapping) *prometheus.SummaryVec {
	if containsMetric(excludeMetrics, opts.Name) {
		return nil
	}
	m := prometheus.NewSummaryVec(
		*opts,
		requestTags,
	)
	metricMapping[prometheus.BuildFQName(PrometheusNamespace, "", opts.Name)] = m
	return m
}

func counterMetric(opts *prometheus.CounterOpts, requestTags []string, excludeMetrics map[string]struct{}, metricMapping metricMapping) *prometheus.CounterVec {
	if containsMetric(excludeMetrics, opts.Name) {
		return nil
	}
	m := prometheus.NewCounterVec(
		*opts,
		requestTags,
	)
	metricMapping[prometheus.BuildFQName(PrometheusNamespace, "", opts.Name)] = m
	return m
}

func histogramMetric(opts *prometheus.HistogramOpts, requestTags []string, excludeMetrics map[string]struct{}, metricMapping metricMapping) *prometheus.HistogramVec {
	if containsMetric(excludeMetrics, opts.Name) {
		return nil
	}
	m := prometheus.NewHistogramVec(
		*opts,
		requestTags,
	)
	metricMapping[prometheus.BuildFQName(PrometheusNamespace, "", opts.Name)] = m
	return m
}

func (sc *SocketCollector) handleMessage(msg []byte) {
	klog.V(5).InfoS("Metric", "message", string(msg))

	// Unmarshal bytes
	var statsBatch []socketData
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(msg, &statsBatch)
	if err != nil {
		klog.ErrorS(err, "Unexpected error deserializing JSON", "payload", string(msg))
		return
	}

	for i := range statsBatch {
		stats := &statsBatch[i]
		if sc.metricsPerHost && !sc.hosts.Has(stats.Host) {
			klog.V(3).InfoS("Skipping metric for host not being served", "host", stats.Host)
			continue
		}

		if sc.reportStatusClasses && len(stats.Status) > 0 {
			stats.Status = fmt.Sprintf("%cxx", stats.Status[0])
		}

		// Note these must match the order in requestTags at the top
		requestLabels := prometheus.Labels{
			"status":    stats.Status,
			"method":    stats.Method,
			"path":      stats.Path,
			"namespace": stats.Namespace,
			"ingress":   stats.Ingress,
			"service":   stats.Service,
			"canary":    stats.Canary,
		}

		collectorLabels := prometheus.Labels{
			"namespace": stats.Namespace,
			"ingress":   stats.Ingress,
			"status":    stats.Status,
			"service":   stats.Service,
			"canary":    stats.Canary,
			"method":    stats.Method,
			"path":      stats.Path,
		}
		if sc.metricsPerHost {
			requestLabels["host"] = stats.Host
			collectorLabels["host"] = stats.Host
		}

		latencyLabels := prometheus.Labels{
			"namespace": stats.Namespace,
			"ingress":   stats.Ingress,
			"service":   stats.Service,
			"canary":    stats.Canary,
		}

		if sc.requests != nil {
			requestsMetric, err := sc.requests.GetMetricWith(collectorLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching requests metric")
			} else {
				requestsMetric.Inc()
			}
		}

		if stats.Latency != -1 {
			if sc.connectTime != nil {
				connectTimeMetric, err := sc.connectTime.GetMetricWith(requestLabels)
				if err != nil {
					klog.ErrorS(err, "Error fetching connect time metric")
				} else {
					connectTimeMetric.Observe(stats.Latency)
				}
			}

			if sc.upstreamLatency != nil {
				latencyMetric, err := sc.upstreamLatency.GetMetricWith(latencyLabels)
				if err != nil {
					klog.ErrorS(err, "Error fetching latency metric")
				} else {
					latencyMetric.Observe(stats.Latency)
				}
			}
		}

		if stats.HeaderTime != -1 && sc.headerTime != nil {
			headerTimeMetric, err := sc.headerTime.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching header time metric")
			} else {
				headerTimeMetric.Observe(stats.HeaderTime)
			}
		}

		if stats.RequestTime != -1 && sc.requestTime != nil {
			requestTimeMetric, err := sc.requestTime.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching request duration metric")
			} else {
				requestTimeMetric.Observe(stats.RequestTime)
			}
		}

		if stats.RequestLength != -1 && sc.requestLength != nil {
			requestLengthMetric, err := sc.requestLength.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching request length metric")
			} else {
				requestLengthMetric.Observe(stats.RequestLength)
			}
		}

		if stats.ResponseTime != -1 && sc.responseTime != nil {
			responseTimeMetric, err := sc.responseTime.GetMetricWith(requestLabels)
			if err != nil {
				klog.ErrorS(err, "Error fetching upstream response time metric")
			} else {
				responseTimeMetric.Observe(stats.ResponseTime)
			}
		}

		if stats.ResponseLength != -1 {
			if sc.bytesSent != nil {
				bytesSentMetric, err := sc.bytesSent.GetMetricWith(requestLabels)
				if err != nil {
					klog.ErrorS(err, "Error fetching bytes sent metric")
				} else {
					bytesSentMetric.Observe(stats.ResponseLength)
				}
			}

			if sc.responseLength != nil {
				responseSizeMetric, err := sc.responseLength.GetMetricWith(requestLabels)
				if err != nil {
					klog.ErrorS(err, "Error fetching bytes sent metric")
				} else {
					responseSizeMetric.Observe(stats.ResponseLength)
				}
			}
		}
	}
}

// Start listen for connections in the unix socket and spawns a goroutine to process the content
func (sc *SocketCollector) Start() {
	for {
		conn, err := sc.listener.Accept()
		if err != nil {
			continue
		}

		go handleMessages(conn, sc.handleMessage)
	}
}

// Stop stops unix listener
func (sc *SocketCollector) Stop() {
	sc.listener.Close()
}

// RemoveMetrics deletes prometheus metrics from prometheus for ingresses and
// host that are not available anymore.
// Ref: https://godoc.org/github.com/prometheus/client_golang/prometheus#CounterVec.Delete
func (sc *SocketCollector) RemoveMetrics(ingresses []string, registry prometheus.Gatherer) {
	mfs, err := registry.Gather()
	if err != nil {
		klog.ErrorS(err, "Error gathering metrics: %v")
		return
	}

	// 1. remove metrics of removed ingresses
	klog.V(2).InfoS("removing metrics", "ingresses", ingresses)
	for _, mf := range mfs {
		metricName := mf.GetName()
		metric, ok := sc.metricMapping[metricName]
		if !ok {
			continue
		}

		toRemove := sets.NewString(ingresses...)
		for _, m := range mf.GetMetric() {
			labels := make(map[string]string, len(m.GetLabel()))
			for _, labelPair := range m.GetLabel() {
				labels[*labelPair.Name] = *labelPair.Value
			}

			// remove labels that are constant
			deleteConstants(labels)

			ns, ok := labels["namespace"]
			if !ok {
				continue
			}
			ing, ok := labels["ingress"]
			if !ok {
				continue
			}

			ingKey := fmt.Sprintf("%v/%v", ns, ing)
			if !toRemove.Has(ingKey) {
				continue
			}

			klog.V(2).Infof("Removing prometheus metric from histogram %v for ingress %v", metricName, ingKey)

			h, ok := metric.(*prometheus.HistogramVec)
			if ok {
				removed := h.Delete(labels)
				if !removed {
					klog.V(2).InfoS("metric not removed", "name", metricName, "ingress", ingKey, "labels", labels)
				}
			}

			s, ok := metric.(*prometheus.SummaryVec)
			if ok {
				removed := s.Delete(labels)
				if !removed {
					klog.V(2).InfoS("metric not removed", "name", metricName, "ingress", ingKey, "labels", labels)
				}
			}

			if c, ok := metric.(*prometheus.CounterVec); ok {
				if removed := c.Delete(labels); !removed {
					klog.V(2).InfoS("metric not removed", "name", metricName, "ingress", ingKey, "labels", labels)
				}
			}
		}
	}
}

// Describe implements prometheus.Collector
func (sc *SocketCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range sc.metricMapping {
		metric.Describe(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (sc *SocketCollector) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range sc.metricMapping {
		metric.Collect(ch)
	}
}

// SetHosts sets the hostnames that are being served by the ingress controller
// This set of hostnames is used to filter the metrics to be exposed
func (sc *SocketCollector) SetHosts(hosts sets.Set[string]) {
	sc.hosts = hosts
}

// handleMessages process the content received in a network connection
func handleMessages(conn io.ReadCloser, fn func([]byte)) {
	defer conn.Close()
	data, err := io.ReadAll(conn)
	if err != nil {
		return
	}

	fn(data)
}

func deleteConstants(labels prometheus.Labels) {
	delete(labels, "controller_namespace")
	delete(labels, "controller_class")
	delete(labels, "controller_pod")
}
