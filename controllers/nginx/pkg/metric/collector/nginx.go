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
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	activeDesc = prometheus.NewDesc(
		"nginx_active_connections",
		"total number of active connections",
		nil, nil)

	acceptedDesc = prometheus.NewDesc(
		"nginx_accepted_connections",
		"total number of accepted client connections",
		nil, nil)

	handledDesc = prometheus.NewDesc(
		"nginx_handled_connections",
		"total number of handled connections",
		nil, nil)

	requestsDesc = prometheus.NewDesc(
		"nginx_total_requests",
		"total number of client requests",
		nil, nil)

	readingDesc = prometheus.NewDesc(
		"nginx_current_reading_connections",
		"current number of connections where nginx is reading the request header",
		nil, nil)

	writingDesc = prometheus.NewDesc(
		"nginx_current_writing_connections",
		"current number of connections where nginx is writing the response back to the client",
		nil, nil)

	waitingDesc = prometheus.NewDesc(
		"nginx_current_waiting_connections",
		"current number of idle client connections waiting for a request",
		nil, nil)
)

type (
	nginxStatusCollector struct {
		scrapeChan chan scrapeRequest
	}
)

func NewNginxStatus() (prometheus.Collector, error) {
	p := nginxStatusCollector{
		scrapeChan: make(chan scrapeRequest),
	}

	go p.start()

	return p, nil
}

// Describe implements prometheus.Collector.
func (p nginxStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- activeDesc
	ch <- acceptedDesc
	ch <- handledDesc
	ch <- requestsDesc
	ch <- readingDesc
	ch <- writingDesc
	ch <- waitingDesc
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
	s, err := getNginxStatus()
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx status info: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(activeDesc,
		prometheus.GaugeValue, float64(s.Active))
	ch <- prometheus.MustNewConstMetric(acceptedDesc,
		prometheus.GaugeValue, float64(s.Accepted))
	ch <- prometheus.MustNewConstMetric(handledDesc,
		prometheus.GaugeValue, float64(s.Handled))
	ch <- prometheus.MustNewConstMetric(requestsDesc,
		prometheus.GaugeValue, float64(s.Requests))
	ch <- prometheus.MustNewConstMetric(readingDesc,
		prometheus.GaugeValue, float64(s.Reading))
	ch <- prometheus.MustNewConstMetric(writingDesc,
		prometheus.GaugeValue, float64(s.Writing))
	ch <- prometheus.MustNewConstMetric(waitingDesc,
		prometheus.GaugeValue, float64(s.Waiting))

}
