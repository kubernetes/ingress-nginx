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

package collector

import (
	"bytes"
	"encoding/json"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type socketData struct {
	Host   string `json:"host"`   // Label
	Status string `json:"status"` // Label

	BytesSent float64 `json:"bytesSent"` // Metric

	Protocol string `json:"protocol"` // Label
	Method   string `json:"method"`   // Label

	RequestLength float64 `json:"requestLength"` // Metric
	RequestTime   float64 `json:"requestTime"`   // Metric

	UpstreamName         string  `json:"upstreamName"`         // Label
	UpstreamIP           string  `json:"upstreamIP"`           // Label
	UpstreamResponseTime float64 `json:"upstreamResponseTime"` // Metric
	UpstreamStatus       string  `json:"upstreamStatus"`       // Label

	Namespace string `json:"namespace"` // Label
	Ingress   string `json:"ingress"`   // Label
	Service   string `json:"service"`   // Label
	Path      string `json:"path"`      // Label
}

// SocketCollector stores prometheus metrics and ingress meta-data
type SocketCollector struct {
	upstreamResponseTime *prometheus.HistogramVec
	requestTime          *prometheus.HistogramVec
	requestLength        *prometheus.HistogramVec
	bytesSent            *prometheus.HistogramVec
	collectorSuccess     *prometheus.GaugeVec
	collectorSuccessTime *prometheus.GaugeVec
	requests             *prometheus.CounterVec
	listener             net.Listener
	ns                   string
	ingressClass         string

	metricMapping map[string]*prometheus.HistogramVec
}

// NewInstance creates a new SocketCollector instance
func NewInstance(ns string, class string) (*SocketCollector, error) {
	sc := &SocketCollector{}

	ns = strings.Replace(ns, "-", "_", -1)

	listener, err := net.Listen("unix", "/tmp/prometheus-nginx.socket")
	if err != nil {
		return nil, err
	}

	sc.listener = listener
	sc.ns = ns
	sc.ingressClass = class

	requestTags := []string{"host", "status", "protocol", "method", "path", "upstream_name", "upstream_ip", "upstream_status", "namespace", "ingress", "service"}
	collectorTags := []string{"namespace", "ingress_class"}

	sc.upstreamResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "upstream_response_time_seconds",
			Help:      "The time spent on receiving the response from the upstream server",
			Namespace: ns,
		},
		requestTags,
	)

	sc.requestTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_duration_seconds",
			Help:      "The request processing time in seconds",
			Namespace: ns,
		},
		requestTags,
	)

	sc.requestLength = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_length_bytes",
			Help:      "The request length (including request line, header, and request body)",
			Namespace: ns,
			Buckets:   prometheus.LinearBuckets(10, 10, 10), // 10 buckets, each 10 bytes wide.
		},
		requestTags,
	)

	sc.requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "requests",
			Help:      "The total number of client requests.",
			Namespace: ns,
		},
		collectorTags,
	)

	sc.bytesSent = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "bytes_sent",
			Help:      "The the number of bytes sent to a client",
			Namespace: ns,
			Buckets:   prometheus.ExponentialBuckets(10, 10, 7), // 7 buckets, exponential factor of 10.
		},
		requestTags,
	)

	sc.collectorSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "collector_last_run_successful",
			Help:      "Whether the last collector run was successful (success = 1, failure = 0).",
			Namespace: ns,
		},
		collectorTags,
	)

	sc.collectorSuccessTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "collector_last_run_successful_timestamp_seconds",
			Help:      "Timestamp of the last successful collector run",
			Namespace: ns,
		},
		collectorTags,
	)

	prometheus.MustRegister(sc.upstreamResponseTime)
	prometheus.MustRegister(sc.requestTime)
	prometheus.MustRegister(sc.requestLength)
	prometheus.MustRegister(sc.requests)
	prometheus.MustRegister(sc.bytesSent)
	prometheus.MustRegister(sc.collectorSuccess)
	prometheus.MustRegister(sc.collectorSuccessTime)

	go sc.Run()

	sc.metricMapping = map[string]*prometheus.HistogramVec{
		"upstream_response_time_seconds_bucket": sc.upstreamResponseTime,
		"request_duration_seconds_bucket":       sc.requestLength,
		"request_length_bytes_bucket":           sc.requestLength,
		"bytes_sent_bucket":                     sc.bytesSent,
	}

	return sc, nil
}

func (sc *SocketCollector) handleMessage(msg []byte) {
	glog.V(5).Infof("msg: %v", string(msg))

	collectorSuccess := true

	// Unmarshall bytes
	var stats socketData
	err := json.Unmarshal(msg, &stats)
	if err != nil {
		glog.Errorf("Unexpected error deserializing JSON paylod: %v", err)
		collectorSuccess = false
		return
	}

	// Create Request Labels Map
	requestLabels := prometheus.Labels{
		"host":            stats.Host,
		"status":          stats.Status,
		"protocol":        stats.Protocol,
		"method":          stats.Method,
		"path":            stats.Path,
		"upstream_name":   stats.UpstreamName,
		"upstream_ip":     stats.UpstreamIP,
		"upstream_status": stats.UpstreamStatus,
		"namespace":       stats.Namespace,
		"ingress":         stats.Ingress,
		"service":         stats.Service,
	}

	// Create Collector Labels Map
	collectorLabels := prometheus.Labels{
		"namespace":     sc.ns,
		"ingress_class": sc.ingressClass,
	}

	// Emit metrics
	requestsMetric, err := sc.requests.GetMetricWith(collectorLabels)
	if err != nil {
		glog.Errorf("Error fetching requests metric: %v", err)
		collectorSuccess = false
	} else {
		requestsMetric.Inc()
	}

	if stats.UpstreamResponseTime != -1 {
		upstreamResponseTimeMetric, err := sc.upstreamResponseTime.GetMetricWith(requestLabels)
		if err != nil {
			glog.Errorf("Error fetching upstream response time metric: %v", err)
			collectorSuccess = false
		} else {
			upstreamResponseTimeMetric.Observe(stats.UpstreamResponseTime)
		}
	}

	if stats.RequestTime != -1 {
		requestTimeMetric, err := sc.requestTime.GetMetricWith(requestLabels)
		if err != nil {
			glog.Errorf("Error fetching request duration metric: %v", err)
			collectorSuccess = false
		} else {
			requestTimeMetric.Observe(stats.RequestTime)
		}
	}

	if stats.RequestLength != -1 {
		requestLengthMetric, err := sc.requestLength.GetMetricWith(requestLabels)
		if err != nil {
			glog.Errorf("Error fetching request length metric: %v", err)
			collectorSuccess = false
		} else {
			requestLengthMetric.Observe(stats.RequestLength)
		}
	}

	if stats.BytesSent != -1 {
		bytesSentMetric, err := sc.bytesSent.GetMetricWith(requestLabels)
		if err != nil {
			glog.Errorf("Error fetching bytes sent metric: %v", err)
			collectorSuccess = false
		} else {
			bytesSentMetric.Observe(stats.BytesSent)
		}
	}

	collectorSuccessMetric, err := sc.collectorSuccess.GetMetricWith(collectorLabels)
	if err != nil {
		glog.Errorf("Error fetching collector success metric: %v", err)
	} else {
		if collectorSuccess {
			collectorSuccessMetric.Set(1)
			collectorSuccessTimeMetric, err := sc.collectorSuccessTime.GetMetricWith(collectorLabels)
			if err != nil {
				glog.Errorf("Error fetching collector success time metric: %v", err)
			} else {
				collectorSuccessTimeMetric.Set(float64(time.Now().Unix()))
			}
		} else {
			collectorSuccessMetric.Set(0)
		}
	}
}

// Run listen for connections in the unix socket and spawns a goroutine to process the content
func (sc *SocketCollector) Run() {
	for {
		conn, err := sc.listener.Accept()
		if err != nil {
			continue
		}

		go handleMessages(conn, sc.handleMessage)
	}
}

// RemoveMetrics deletes prometheus metrics from prometheus for endpoints that
// are not available anymore.
// To remove prometheus metrics is necessary to pass all the labels present in the metric.
// For this reason we use the current metric state to filter those metrics using the IP
// address and port of the endpoint to be removed.
// Ref: https://godoc.org/github.com/prometheus/client_golang/prometheus#CounterVec.DeleteLabelValues
func (sc *SocketCollector) RemoveMetrics(metrics string, endpoints []string) {
	for _, endpoint := range endpoints {
		for metricName, promHistogram := range sc.metricMapping {
			glog.Infof("Removing prometheus metric from histogram %v for endpoint %v", metricName, endpoint)

			em := extractMetrics(endpoint, metricName, metrics)
			for _, metric := range em {
				l := parseLabels(metric)
				if len(l) == 0 {
					continue
				}

				promHistogram.DeleteLabelValues(l...)
			}
		}
	}
}

const packetSize = 1024 * 65

// handleMessages process the content received in a network connection
func handleMessages(conn net.Conn, fn func([]byte)) {
	defer conn.Close()

	msg := make([]byte, packetSize)
	s, err := conn.Read(msg[0:])
	if err != nil {
		return
	}

	fn(msg[0:s])
}

var execCommand = exec.Command

// extractMetrics filters the metrics using the endpoint information and the name
// of the histogram from where we need to remove the metric.
// Each item in the array contains a complete metric.
func extractMetrics(endpoint, metricName, metrics string) []string {
	cmd := execCommand("/ingress-controller/extract-metrics.sh", endpoint, metricName)
	cmd.Stdin = bytes.NewBufferString(metrics)
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Warningf("unexpected error extracting metrics: %v", err)
		return nil
	}

	s := strings.SplitAfter(string(out), "|")
	var r []string
	for _, s := range s {
		if s == "" || s == "\n" {
			continue
		}
		r = append(r, strings.TrimSpace(s))
	}
	glog.V(3).Infof("Extracted metrics (%v) for endpoint %v: %v", metricName, endpoint, r)
	return r
}

// regex to parse a single label k=v
var kvRegex = regexp.MustCompile(`(\w+)="(.*)"`)

// parse a string containing a prometheus metric returning an array
// with the label values to be used in the delete operation
func parseLabels(input string) []string {
	tokens := strings.Split(input, ",")

	data := make(map[string]string, 11)
	for _, pairs := range tokens {
		kv := kvRegex.FindStringSubmatch(pairs)
		if len(kv) != 3 {
			continue
		}

		data[kv[1]] = kv[2]
	}

	return []string{
		data["host"],
		data["status"],
		data["protocol"],
		data["method"],
		data["path"],
		data["upstream_name"],
		data["upstream_ip"],
		data["upstream_status"],
		data["namespace"],
		data["ingress"],
		data["service"],
	}
}
