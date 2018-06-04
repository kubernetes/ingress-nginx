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
	"net"
	"time"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

const ns = "nginx"

type data struct {
	Host   string `json:"host"`   // Label
	Status int    `json:"status"` // Label

	Time time.Duration `json:"time"` // Metric

	RemoteAddress string `json:"remoteAddr"` // Label
	RemoteUser    string `json:"remoteUser"` // Label

	BytesSent int64 `json:"bytesSent"` // Metric

	Protocol string `json:"protocol"` // Label
	Method   string `json:"method"`   // Label
	Path     string `json:"path"`     // Label

	RequestTime   string `json:"requestTime"`   // Metric
	RequestLength string `json:"requestLength"` // Metric
	Duration      int    `json:"duration"`      // Metric DONE

	UpstreamName         string `json:"upstreamName"`         // Label
	UpstreamIP           string `json:"upstreamIP"`           // Label
	UpstreamResponseTime string `json:"upstreamResponseTime"` // Metric DONE
	UpstreamStatus       string `json:"upstreamStatus"`       // Label

	Namespace string `json:"namespace"` // Label
	Ingress   string `json:"ingress"`   // Label
	Service   string `json:"service"`   // Label
}

type statsCollector struct {
	upstreamResponseTime prometheus.*HistogramVec
	requestDuration prometheus.*HistogramVec
	requestLength prometheus.*HistogramVec
	bytesSent prometheus.*HistogramVec
	listener *net.UDPConn
	ns string
	watchClass string
	port int
}

func NewInstance(ns string, class string, port int) {
	sc := statsCollector{}

	listener, err := newUDPListener(port)

	if err != nil {
		return nil, err
	}

	sc.listener := listener
	sc.ns := ns
	sc.watchClass := class
	sc.port := port

	tags := []string{"host", "status", "remote_address", "remote_user", "protocol", "method", "path", "upstream_name", "upstream_ip", "upstream_response_time", "upstream_status", "namespace", "ingress", "service"}

	sc.upstreamResponseTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "upstream_response_time_seconds",
			Help:    "The time spent on receiving the response from the upstream server",
			Namespace: ns,
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10), // 10 buckets, each 0.1 seconds wide.
		},
		tags
	)

	sc.requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "The request processing time in seconds",
			Namespace: ns,
			Buckets: prometheus.LinearBuckets(0.5, 0.5, 20), // 20 buckets, each 0.5 seconds wide.
		},
		tags
	)

	sc.requestLength := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_length_bytes",
			Help:    "The request length (including request line, header, and request body)",
			Namespace: ns,
			Buckets: prometheus.LinearBuckets(20, 20, 20), // 20 buckets, each 20 bytes wide.
		},
		tags
	)

	sc.bytesSent := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bytes_sent",
			Help:    "The the number of bytes sent to a client",
			Namespace: ns,
			Buckets: prometheus.LinearBuckets(100, 100, 20), // 20 buckets, each 100 bytes wide.
		},
		tags
	)

	prometheus.MustRegister(upstreamResponseTime)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestLength)
	prometheus.MustRegister(bytesSent)
	return &sc
}

func (sc *statsCollector) handleMessage(msg []byte) {
	glog.Infof("msg: %v", string(msg))

	// Unmarshall bytes
	var stats data
	err := json.Unmarshall(msg, &stats)
	if err != nil {
		panic(err)
	}

	// Create Labels Map
	labels := prometheus.Labels{
		"host": stats.Host,
		"status": stats.Status,
		"remote_address": stats.RemoteAddress,
		"remote_user": stats.RemoteUser,
		"protocol": stats.Protocol,
		"method": stats.Method,
		"path": stats.Path,
		"upstream_name": stats.UpstreamName,
		"upstream_ip": stats.UpstreamIP,
		"upstream_status": stats.UpstreamStatus,
		"namespace": status.Namespace,
		"ingress": status.Ingress,
		"service": status.Service
	}

	// Emit metrics
	sc.upstreamResponseTime.GetMetricWith(labels).Observe(stats.UpstreamResponseTime)
	sc.requestDuration.GetMetricWith(labels).Observe(stats.UpstreamResponseTime)
	sc.requestLength.GetMetricWith(labels).Observe(stats.RequestLength)
	sc.bytesSent.GetMetricWith(labels).Observe(stats.BytesSent)

}

func (sc *statsCollector) Run() {
	handleMessages(sc.listener, sc.handleMessage)
}
