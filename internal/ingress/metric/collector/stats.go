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
	"encoding/json"
	"net"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type udpData struct {
	Host   string `json:"host"`   // Label
	Status string `json:"status"` // Label

	Time string `json:"time"` // Metric

	RemoteAddress string `json:"remoteAddr"` // Label
	RemoteUser    string `json:"remoteUser"` // Label

	BytesSent float64 `json:"bytesSent,string"` // Metric

	Protocol string `json:"protocol"` // Label
	Method   string `json:"method"`   // Label
	Path     string `json:"path"`     // Label

	RequestTime   string  `json:"requestTime"`          // Metric
	RequestLength float64 `json:"requestLength,string"` // Metric
	Duration      float64 `json:"duration,string"`      // Metric

	UpstreamName         string  `json:"upstreamName"`                // Label
	UpstreamIP           string  `json:"upstreamIP"`                  // Label
	UpstreamResponseTime float64 `json:"upstreamResponseTime,string"` // Metric
	UpstreamStatus       string  `json:"upstreamStatus"`              // Label

	Namespace string `json:"namespace"` // Label
	Ingress   string `json:"ingress"`   // Label
	Service   string `json:"service"`   // Label
}

type StatsCollector struct {
	upstreamResponseTime *prometheus.HistogramVec
	requestDuration      *prometheus.HistogramVec
	requestLength        *prometheus.HistogramVec
	bytesSent            *prometheus.HistogramVec
	listener             *net.UDPConn
	ns                   string
	watchClass           string
	port                 int
}

func NewInstance(ns string, class string, port int) (*StatsCollector, error) {
	sc := StatsCollector{}

	listener, err := newUDPListener(port)

	if err != nil {
		return nil, err
	}

	sc.listener = listener
	sc.ns = ns
	sc.watchClass = class
	sc.port = port

	tags := []string{"host", "status", "remote_address", "remote_user", "protocol", "method", "path", "upstream_name", "upstream_ip", "upstream_status", "namespace", "ingress", "service"}

	sc.upstreamResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "upstream_response_time_seconds",
			Help:      "The time spent on receiving the response from the upstream server",
			Namespace: ns,
			Buckets:   prometheus.LinearBuckets(0.1, 0.1, 10), // 10 buckets, each 0.1 seconds wide.
		},
		tags,
	)

	sc.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_duration_seconds",
			Help:      "The request processing time in seconds",
			Namespace: ns,
			Buckets:   prometheus.LinearBuckets(0.5, 0.5, 20), // 20 buckets, each 0.5 seconds wide.
		},
		tags,
	)

	sc.requestLength = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_length_bytes",
			Help:      "The request length (including request line, header, and request body)",
			Namespace: ns,
			Buckets:   prometheus.LinearBuckets(20, 20, 20), // 20 buckets, each 20 bytes wide.
		},
		tags,
	)

	sc.bytesSent = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "bytes_sent",
			Help:      "The the number of bytes sent to a client",
			Namespace: ns,
			Buckets:   prometheus.LinearBuckets(100, 100, 20), // 20 buckets, each 100 bytes wide.
		},
		tags,
	)

	prometheus.MustRegister(sc.upstreamResponseTime)
	prometheus.MustRegister(sc.requestDuration)
	prometheus.MustRegister(sc.requestLength)
	prometheus.MustRegister(sc.bytesSent)
	return &sc, nil
}

func (sc *StatsCollector) handleMessage(msg []byte) {
	glog.Infof("msg: %v", string(msg))

	// Unmarshall bytes
	var stats udpData
	err := json.Unmarshal(msg, &stats)
	if err != nil {
		panic(err)
	}

	// Create Labels Map
	labels := prometheus.Labels{
		"host":            stats.Host,
		"status":          stats.Status,
		"remote_address":  stats.RemoteAddress,
		"remote_user":     stats.RemoteUser,
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

	// Emit metrics
	upstreamResponseTimeMetric, err := sc.upstreamResponseTime.GetMetricWith(labels)
	if err != nil {
		glog.Errorf("Error fetching upstream response time metric: %v", err)
	}
	upstreamResponseTimeMetric.Observe(stats.UpstreamResponseTime)

	requestDurationMetric, err := sc.requestDuration.GetMetricWith(labels)
	if err != nil {
		glog.Errorf("Error fetching request duration metric: %v", err)
	}
	requestDurationMetric.Observe(stats.Duration)

	requestLengthMetric, err := sc.requestLength.GetMetricWith(labels)
	if err != nil {
		glog.Errorf("Error fetching request length metric: %v", err)
	}
	requestLengthMetric.Observe(stats.RequestLength)

	bytesSentMetric, err := sc.bytesSent.GetMetricWith(labels)
	if err != nil {
		glog.Errorf("Error fetching bytes sent metric: %v", err)
	}
	bytesSentMetric.Observe(stats.BytesSent)

}

func (sc *StatsCollector) Run() {
	handleMessages(sc.listener, sc.handleMessage)
}
