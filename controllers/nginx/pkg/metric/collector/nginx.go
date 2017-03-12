/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	nginxStatusCollector struct {
		scrapeChan    chan scrapeRequest
		ngxHealthPort int
		ngxVtsPath    string
		data          *nginxStatusData
	}

	nginxStatusData struct {
		active   *prometheus.Desc
		accepted *prometheus.Desc
		handled  *prometheus.Desc
		requests *prometheus.Desc
		reading  *prometheus.Desc
		writing  *prometheus.Desc
		waiting  *prometheus.Desc
	}
)

func buildNS(namespace, class string) string {
	if namespace == "" {
		namespace = "all"
	}
	if class == "" {
		class = "all"
	}

	return fmt.Sprintf("%v_%v", namespace, class)
}

// NewNginxStatus returns a new prometheus collector the default nginx status module
func NewNginxStatus(namespace, class string, ngxHealthPort int, ngxVtsPath string) Stopable {
	p := nginxStatusCollector{
		scrapeChan:    make(chan scrapeRequest),
		ngxHealthPort: ngxHealthPort,
		ngxVtsPath:    ngxVtsPath,
	}

	ns := buildNS(namespace, class)

	p.data = &nginxStatusData{
		active: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "active_connections"),
			"total number of active connections",
			nil, nil),

		accepted: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "accepted_connections"),
			"total number of accepted client connections",
			nil, nil),

		handled: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "handled_connections"),
			"total number of handled connections",
			nil, nil),

		requests: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "total_requests"),
			"total number of client requests",
			nil, nil),

		reading: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "current_reading_connections"),
			"current number of connections where nginx is reading the request header",
			nil, nil),

		writing: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "current_writing_connections"),
			"current number of connections where nginx is writing the response back to the client",
			nil, nil),

		waiting: prometheus.NewDesc(
			prometheus.BuildFQName(system, ns, "current_waiting_connections"),
			"current number of idle client connections waiting for a request",
			nil, nil),
	}

	go p.start()

	return p
}

// Describe implements prometheus.Collector.
func (p nginxStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.data.active
	ch <- p.data.accepted
	ch <- p.data.handled
	ch <- p.data.requests
	ch <- p.data.reading
	ch <- p.data.writing
	ch <- p.data.waiting
}

// Collect implements prometheus.Collector.
func (p nginxStatusCollector) Collect(ch chan<- prometheus.Metric) {
	req := scrapeRequest{results: ch, done: make(chan struct{})}
	p.scrapeChan <- req
	<-req.done
}

func (p nginxStatusCollector) start() {
	for req := range p.scrapeChan {
		ch := req.results
		p.scrape(ch)
		req.done <- struct{}{}
	}
}

func (p nginxStatusCollector) Stop() {
	close(p.scrapeChan)
}

// nginxStatusCollector scrap the nginx status
func (p nginxStatusCollector) scrape(ch chan<- prometheus.Metric) {
	s, err := getNginxStatus(p.ngxHealthPort, p.ngxVtsPath)
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx status info: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(p.data.active,
		prometheus.GaugeValue, float64(s.Active))
	ch <- prometheus.MustNewConstMetric(p.data.accepted,
		prometheus.GaugeValue, float64(s.Accepted))
	ch <- prometheus.MustNewConstMetric(p.data.handled,
		prometheus.GaugeValue, float64(s.Handled))
	ch <- prometheus.MustNewConstMetric(p.data.requests,
		prometheus.GaugeValue, float64(s.Requests))
	ch <- prometheus.MustNewConstMetric(p.data.reading,
		prometheus.GaugeValue, float64(s.Reading))
	ch <- prometheus.MustNewConstMetric(p.data.writing,
		prometheus.GaugeValue, float64(s.Writing))
	ch <- prometheus.MustNewConstMetric(p.data.waiting,
		prometheus.GaugeValue, float64(s.Waiting))
}
