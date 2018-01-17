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

type (
	nginxStatusCollector struct {
		scrapeChan     chan scrapeRequest
		ngxHealthPort  int
		ngxVtsPath     string
		data           *nginxStatusData
		watchNamespace string
		ingressClass   string
	}

	nginxStatusData struct {
		connectionsTotal *prometheus.Desc
		requestsTotal    *prometheus.Desc
		connections      *prometheus.Desc
	}
)

// NewNginxStatus returns a new prometheus collector the default nginx status module
func NewNginxStatus(watchNamespace, ingressClass string, ngxHealthPort int, ngxVtsPath string) Stopable {

	p := nginxStatusCollector{
		scrapeChan:     make(chan scrapeRequest),
		ngxHealthPort:  ngxHealthPort,
		ngxVtsPath:     ngxVtsPath,
		watchNamespace: watchNamespace,
		ingressClass:   ingressClass,
	}

	p.data = &nginxStatusData{
		connectionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "connections_total"),
			"total number of connections with state {active, accepted, handled}",
			[]string{"ingress_class", "namespace", "state"}, nil),

		requestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "requests_total"),
			"total number of client requests",
			[]string{"ingress_class", "namespace"}, nil),

		connections: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "connnections"),
			"current number of client connections with state {reading, writing, waiting}",
			[]string{"ingress_class", "namespace", "state"}, nil),
	}

	go p.start()

	return p
}

// Describe implements prometheus.Collector.
func (p nginxStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.data.connectionsTotal
	ch <- p.data.requestsTotal
	ch <- p.data.connections
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

// nginxStatusCollector scrape the nginx status
func (p nginxStatusCollector) scrape(ch chan<- prometheus.Metric) {
	s, err := getNginxStatus(p.ngxHealthPort, p.ngxVtsPath)
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx status info: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(p.data.connectionsTotal,
		prometheus.CounterValue, float64(s.Active), p.ingressClass, p.watchNamespace, "active")
	ch <- prometheus.MustNewConstMetric(p.data.connectionsTotal,
		prometheus.CounterValue, float64(s.Accepted), p.ingressClass, p.watchNamespace, "accepted")
	ch <- prometheus.MustNewConstMetric(p.data.connectionsTotal,
		prometheus.CounterValue, float64(s.Handled), p.ingressClass, p.watchNamespace, "handled")
	ch <- prometheus.MustNewConstMetric(p.data.requestsTotal,
		prometheus.CounterValue, float64(s.Requests), p.ingressClass, p.watchNamespace)
	ch <- prometheus.MustNewConstMetric(p.data.connections,
		prometheus.GaugeValue, float64(s.Reading), p.ingressClass, p.watchNamespace, "reading")
	ch <- prometheus.MustNewConstMetric(p.data.connections,
		prometheus.GaugeValue, float64(s.Writing), p.ingressClass, p.watchNamespace, "writing")
	ch <- prometheus.MustNewConstMetric(p.data.connections,
		prometheus.GaugeValue, float64(s.Waiting), p.ingressClass, p.watchNamespace, "waiting")
}
