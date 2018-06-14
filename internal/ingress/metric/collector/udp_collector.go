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
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type udpData struct {
	Host   string `json:"host"`   // Label
	Status string `json:"status"` // Label

	RealIPAddress string `json:"realIpAddr"` // Label
	RemoteAddress string `json:"remoteAddr"` // Label
	RemoteUser    string `json:"remoteUser"` // Label

	BytesSent float64 `json:"bytesSent"` // Metric

	Protocol string `json:"protocol"` // Label
	Method   string `json:"method"`   // Label
	URI      string `json:"uri"`      // Label

	RequestLength float64 `json:"requestLength"` // Metric
	RequestTime   float64 `json:"requestTime"`   // Metric

	UpstreamName         string  `json:"upstreamName"`         // Label
	UpstreamIP           string  `json:"upstreamIP"`           // Label
	UpstreamResponseTime float64 `json:"upstreamResponseTime"` // Metric
	UpstreamStatus       string  `json:"upstreamStatus"`       // Label

	Namespace string `json:"namespace"` // Label
	Ingress   string `json:"ingress"`   // Label
	Service   string `json:"service"`   // Label
}

// UDPCollector stores prometheus metrics and ingress meta-data
type UDPCollector struct {
	upstreamResponseTime *prometheus.HistogramVec
	requestTime          *prometheus.HistogramVec
	requestLength        *prometheus.HistogramVec
	bytesSent            *prometheus.HistogramVec
	collectorSuccess     *prometheus.GaugeVec
	collectorSuccessTime *prometheus.GaugeVec
	requests             *prometheus.CounterVec
	listener             *net.UDPConn
	ns                   string
	ingressClass         string
	port                 int
}

// InitUDPCollector creates a new UDPCollector instance
func InitUDPCollector(ns string, class string, port int) error {
	sc := UDPCollector{}

	ns = strings.Replace(ns, "-", "_", -1)

	listener, err := newUDPListener(port)

	if err != nil {
		return err
	}

	sc.listener = listener
	sc.ns = ns
	sc.ingressClass = class
	sc.port = port

	requestTags := []string{"host", "status", "remote_address", "real_ip_address", "remote_user", "protocol", "method", "uri", "upstream_name", "upstream_ip", "upstream_status", "namespace", "ingress", "service"}
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

	return nil
}

func (sc *UDPCollector) handleMessage(msg []byte) {
	glog.V(5).Infof("msg: %v", string(msg))

	collectorSuccess := true

	// Unmarshall bytes
	var stats udpData
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
		"remote_address":  stats.RemoteAddress,
		"real_ip_address": stats.RealIPAddress,
		"remote_user":     stats.RemoteUser,
		"protocol":        stats.Protocol,
		"method":          stats.Method,
		"uri":             stats.URI,
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

// Run adds a message handler to a UDP listener
func (sc *UDPCollector) Run() {
	handleMessages(sc.listener, sc.handleMessage)
}
