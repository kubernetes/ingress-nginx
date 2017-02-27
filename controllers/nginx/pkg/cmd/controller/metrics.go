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

package main

import (
	"path/filepath"

	"github.com/golang/glog"

	common "github.com/ncabatoff/process-exporter"
	"github.com/ncabatoff/process-exporter/proc"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
)

type exeMatcher struct {
	name string
	args []string
}

func (em exeMatcher) MatchAndName(nacl common.NameAndCmdline) (bool, string) {
	if len(nacl.Cmdline) == 0 {
		return false, ""
	}
	cmd := filepath.Base(nacl.Cmdline[0])
	return em.name == cmd, ""
}

func (n *NGINXController) setupMonitor(args []string) {

	// TODO fix true
	pc, err := newProcessCollector(true, exeMatcher{"nginx", args}, false)
	if err != nil {
		glog.Warningf("unexpected error registering nginx collector: %v", err)
	}
	n.namedProcessCollector = pc

	err = prometheus.Register(pc)
	if err != nil {
		glog.Warningf("unexpected error registering nginx collector: %v", err)
	}

}

func (n *NGINXController) reloadMonitor(enableVts *bool) {
	n.namedProcessCollector.vtsCollector = enableVts
}

var (
	// descriptions borrow from https://github.com/vozlt/nginx-module-vts

	cpuSecsDesc = prometheus.NewDesc(
		"nginx_cpu_seconds_total",
		"Cpu usage in seconds",
		nil, nil)

	numprocsDesc = prometheus.NewDesc(
		"nginx_num_procs",
		"number of processes",
		nil, nil)

	memResidentbytesDesc = prometheus.NewDesc(
		"nginx_resident_memory_bytes",
		"number of bytes of memory in use",
		nil, nil)

	memVirtualbytesDesc = prometheus.NewDesc(
		"nginx_virtual_memory_bytes",
		"number of bytes of memory in use",
		nil, nil)

	readBytesDesc = prometheus.NewDesc(
		"nginx_read_bytes_total",
		"number of bytes read",
		nil, nil)

	startTimeDesc = prometheus.NewDesc(
		"nginx_oldest_start_time_seconds",
		"start time in seconds since 1970/01/01",
		nil, nil)

	writeBytesDesc = prometheus.NewDesc(
		"nginx_write_bytes_total",
		"number of bytes written",
		nil, nil)

	//vts metrics
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
	scrapeRequest struct {
		results chan<- prometheus.Metric
		done    chan struct{}
	}

	namedProcessCollector struct {
		scrapeChan chan scrapeRequest
		*proc.Grouper
		fs           *proc.FS
		vtsCollector *bool
	}
)

func newProcessCollector(
	children bool,
	n common.MatchNamer,
	vtsCollector bool) (*namedProcessCollector, error) {

	//fs, err := proc.NewFS("/proc")
	//if err != nil {
	//	return nil, err
	//}
	p := &namedProcessCollector{
		scrapeChan:   make(chan scrapeRequest),
		Grouper:      proc.NewGrouper(children, n),
		//fs:           fs,
		vtsCollector: &vtsCollector,
	}

	//_, err = p.Update(p.fs.AllProcs())
	//if err != nil {
	//	return nil, err
	//}

	go p.start()

	return p, nil
}

// Describe implements prometheus.Collector.
func (p *namedProcessCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- cpuSecsDesc
	ch <- numprocsDesc
	ch <- readBytesDesc
	ch <- writeBytesDesc
	ch <- memResidentbytesDesc
	ch <- memVirtualbytesDesc
	ch <- startTimeDesc

	if p.vtsCollector == true {

		ch <- vtsBytesDesc
		ch <- vtsCacheDesc
		ch <- vtsConnectionsDesc
		ch <- readBytesDesc
		ch <- vtsRequestDesc
		ch <- vtsResponseDesc
		ch <- writeBytesDesc
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

}

// Collect implements prometheus.Collector.
func (p *namedProcessCollector) Collect(ch chan<- prometheus.Metric) {
	req := scrapeRequest{results: ch, done: make(chan struct{})}
	p.scrapeChan <- req
	<-req.done
}

func (p *namedProcessCollector) start() {

	//glog.Warningf("OOO %v", p.configmap.Data)

	for req := range p.scrapeChan {
		ch := req.results
		p.scrapeNginxStatus(ch)

		if &p.vtsCollector {
			p.scrapeVts(ch)
		}

		req.done <- struct{}{}
	}
}

func (p *namedProcessCollector) scrapeNginxStatus(ch chan<- prometheus.Metric) {
	s, err := getNginxStatus()
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx status info: %v", err)
		return
	}

	p.scrapeProcs(ch)

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

func (p *namedProcessCollector) scrapeVts(ch chan<- prometheus.Metric) {

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

			serverZone = strings.Replace(serverZone, "country::", "", 1)

			reflectMetrics(&zone.Responses, vtsFilterZoneResponseDesc, ch, serverZone, country)
			reflectMetrics(&zone.Cache, vtsFilterZoneCacheDesc, ch, serverZone, country)

			ch <- prometheus.MustNewConstMetric(vtsFilterZoneBytesDesc,
				prometheus.CounterValue, float64(zone.InBytes), serverZone, country, "in")

			ch <- prometheus.MustNewConstMetric(vtsFilterZoneBytesDesc,
				prometheus.CounterValue, float64(zone.OutBytes), serverZone, country, "out")

		}

	}

}

func (p *namedProcessCollector) scrapeProcs(ch chan<- prometheus.Metric) {

	_, err := p.Update(p.fs.AllProcs())
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx process info: %v", err)
		return
	}

	for gname, gcounts := range p.Groups() {
		glog.Infof("%v", gname)
		glog.Infof("%v", gcounts)
		ch <- prometheus.MustNewConstMetric(numprocsDesc,
			prometheus.GaugeValue, float64(gcounts.Procs))
		ch <- prometheus.MustNewConstMetric(memResidentbytesDesc,
			prometheus.GaugeValue, float64(gcounts.Memresident))
		ch <- prometheus.MustNewConstMetric(memVirtualbytesDesc,
			prometheus.GaugeValue, float64(gcounts.Memvirtual))
		ch <- prometheus.MustNewConstMetric(startTimeDesc,
			prometheus.GaugeValue, float64(gcounts.OldestStartTime.Unix()))
		ch <- prometheus.MustNewConstMetric(cpuSecsDesc,
			prometheus.CounterValue, gcounts.Cpu)
		ch <- prometheus.MustNewConstMetric(readBytesDesc,
			prometheus.CounterValue, float64(gcounts.ReadBytes))
		ch <- prometheus.MustNewConstMetric(writeBytesDesc,
			prometheus.CounterValue, float64(gcounts.WriteBytes))
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
