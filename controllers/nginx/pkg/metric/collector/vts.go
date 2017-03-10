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
	"reflect"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	vtsBytesDesc = prometheus.NewDesc(
		"nginx_vts_bytes_total",
		"Nginx bytes count",
		[]string{"server_zone", "direction"}, nil)

	vtsCacheDesc = prometheus.NewDesc(
		"nginx_vts_cache_total",
		"Nginx cache count",
		[]string{"server_zone", "type"}, nil)

	vtsConnectionsDesc = prometheus.NewDesc(
		"nginx_vts_connections_total",
		"Nginx connections count",
		[]string{"type"}, nil)

	vtsResponseDesc = prometheus.NewDesc(
		"nginx_vts_responses_total",
		"The number of responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
		[]string{"server_zone", "status_code"}, nil)

	vtsRequestDesc = prometheus.NewDesc(
		"nginx_vts_requests_total",
		"The total number of requested client connections.",
		[]string{"server_zone"}, nil)

	vtsFilterZoneBytesDesc = prometheus.NewDesc(
		"nginx_vts_filterzone_bytes_total",
		"Nginx bytes count",
		[]string{"server_zone", "country", "direction"}, nil)

	vtsFilterZoneResponseDesc = prometheus.NewDesc(
		"nginx_vts_filterzone_responses_total",
		"The number of responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
		[]string{"server_zone", "country", "status_code"}, nil)

	vtsFilterZoneCacheDesc = prometheus.NewDesc(
		"nginx_vts_filterzone_cache_total",
		"Nginx cache count",
		[]string{"server_zone", "country", "type"}, nil)

	vtsUpstreamBackupDesc = prometheus.NewDesc(
		"nginx_vts_upstream_backup",
		"Current backup setting of the server.",
		[]string{"upstream", "server"}, nil)

	vtsUpstreamBytesDesc = prometheus.NewDesc(
		"nginx_vts_upstream_bytes_total",
		"The total number of bytes sent to this server.",
		[]string{"upstream", "server", "direction"}, nil)

	vtsUpstreamDownDesc = prometheus.NewDesc(
		"nginx_vts_upstream_down_total",
		"Current down setting of the server.",
		[]string{"upstream", "server"}, nil)

	vtsUpstreamFailTimeoutDesc = prometheus.NewDesc(
		"nginx_vts_upstream_fail_timeout",
		"Current fail_timeout setting of the server.",
		[]string{"upstream", "server"}, nil)

	vtsUpstreamMaxFailsDesc = prometheus.NewDesc(
		"nginx_vts_upstream_maxfails",
		"Current max_fails setting of the server.",
		[]string{"upstream", "server"}, nil)

	vtsUpstreamResponsesDesc = prometheus.NewDesc(
		"nginx_vts_upstream_responses_total",
		"The number of upstream responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
		[]string{"upstream", "server", "status_code"}, nil)

	vtsUpstreamRequestDesc = prometheus.NewDesc(
		"nginx_vts_upstream_requests_total",
		"The total number of client connections forwarded to this server.",
		[]string{"upstream", "server"}, nil)

	vtsUpstreamResponseMsecDesc = prometheus.NewDesc(
		"nginx_vts_upstream_response_msecs_avg",
		"The average of only upstream response processing times in milliseconds.",
		[]string{"upstream", "server"}, nil)

	vtsUpstreamWeightDesc = prometheus.NewDesc(
		"nginx_vts_upstream_weight",
		"Current upstream weight setting of the server.",
		[]string{"upstream", "server"}, nil)
)

type vtsCollector struct {
		scrapeChan chan scrapeRequest
	}

func NewNGINXVTSCollector() (prometheus.Collector, error) {
	p := vtsCollector{
		scrapeChan: make(chan scrapeRequest),
	}

	go p.start()

	return p, nil
}

// Describe implements prometheus.Collector.
func (p *vtsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- vtsBytesDesc
	ch <- vtsCacheDesc
	ch <- vtsConnectionsDesc
	ch <- vtsRequestDesc
	ch <- vtsResponseDesc
	ch <- vtsUpstreamBackupDesc
	ch <- vtsUpstreamBytesDesc
	ch <- vtsUpstreamDownDesc
	ch <- vtsUpstreamFailTimeoutDesc
	ch <- vtsUpstreamMaxFailsDesc
	ch <- vtsUpstreamRequestDesc
	ch <- vtsUpstreamResponseMsecDesc
	ch <- vtsUpstreamResponsesDesc
	ch <- vtsUpstreamWeightDesc
	ch <- vtsFilterZoneBytesDesc
	ch <- vtsFilterZoneCacheDesc
	ch <- vtsFilterZoneResponseDesc
}

// Collect implements prometheus.Collector.
func (p *vtsCollector) Collect(ch chan<- prometheus.Metric) {
	req := scrapeRequest{results: ch, done: make(chan struct{})}
	p.scrapeChan <- req
	<-req.done
}

func (p *vtsCollector) start() {
	for req := range p.scrapeChan {
		ch := req.results
		p.scrapeVts(ch)
		req.done <- struct{}{}
	}
}

func (p *vtsCollector) Stop() {
	close(p.scrapeChan)
}

// scrapeVts scrape nginx vts metrics
func (p *vtsCollector) scrapeVts(ch chan<- prometheus.Metric) {
	nginxMetrics, err := getNginxVtsMetrics()
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx status info: %v", err)
		return
	}

	reflectMetrics(&nginxMetrics.Connections, vtsConnectionsDesc, ch)

	for name, zones := range nginxMetrics.UpstreamZones {
		for pos, value := range zones {
			reflectMetrics(&zones[pos].Responses, vtsUpstreamResponsesDesc, ch, name, value.Server)

			ch <- prometheus.MustNewConstMetric(vtsUpstreamRequestDesc,
				prometheus.CounterValue, float64(zones[pos].RequestCounter), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamDownDesc,
				prometheus.CounterValue, float64(zones[pos].Down), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamWeightDesc,
				prometheus.CounterValue, float64(zones[pos].Weight), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamResponseMsecDesc,
				prometheus.CounterValue, float64(zones[pos].ResponseMsec), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamBackupDesc,
				prometheus.CounterValue, float64(zones[pos].Backup), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamFailTimeoutDesc,
				prometheus.CounterValue, float64(zones[pos].FailTimeout), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamMaxFailsDesc,
				prometheus.CounterValue, float64(zones[pos].MaxFails), name, value.Server)
			ch <- prometheus.MustNewConstMetric(vtsUpstreamBytesDesc,
				prometheus.CounterValue, float64(zones[pos].InBytes), name, value.Server, "in")
			ch <- prometheus.MustNewConstMetric(vtsUpstreamBytesDesc,
				prometheus.CounterValue, float64(zones[pos].OutBytes), name, value.Server, "out")
		}
	}

	for name, zone := range nginxMetrics.ServerZones {
		reflectMetrics(&zone.Responses, vtsResponseDesc, ch, name)
		reflectMetrics(&zone.Cache, vtsCacheDesc, ch, name)

		ch <- prometheus.MustNewConstMetric(vtsRequestDesc,
			prometheus.CounterValue, float64(zone.RequestCounter), name)
		ch <- prometheus.MustNewConstMetric(vtsBytesDesc,
			prometheus.CounterValue, float64(zone.InBytes), name, "in")
		ch <- prometheus.MustNewConstMetric(vtsBytesDesc,
			prometheus.CounterValue, float64(zone.OutBytes), name, "out")
	}

	for serverZone, countries := range nginxMetrics.FilterZones {
		for country, zone := range countries {
			reflectMetrics(&zone.Responses, vtsFilterZoneResponseDesc, ch, serverZone, country)
			reflectMetrics(&zone.Cache, vtsFilterZoneCacheDesc, ch, serverZone, country)

			ch <- prometheus.MustNewConstMetric(vtsFilterZoneBytesDesc,
				prometheus.CounterValue, float64(zone.InBytes), serverZone, country, "in")
			ch <- prometheus.MustNewConstMetric(vtsFilterZoneBytesDesc,
				prometheus.CounterValue, float64(zone.OutBytes), serverZone, country, "out")
		}
	}
}

func reflectMetrics(value interface{}, desc *prometheus.Desc, ch chan<- prometheus.Metric, labels ...string) {
	val := reflect.ValueOf(value).Elem()

	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag
		labels := append(labels, tag.Get("json"))
		ch <- prometheus.MustNewConstMetric(desc,
			prometheus.CounterValue, float64(val.Field(i).Interface().(float64)),
			labels...)
	}
}
