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
	pc, err := newProcessCollector(true, exeMatcher{"nginx", args})
	if err != nil {
		glog.Warningf("unexpected error registering nginx collector: %v", err)
	}
	err = prometheus.Register(pc)
	if err != nil {
		glog.Warningf("unexpected error registering nginx collector: %v", err)
	}
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

	//vts metrics
	bytesDesc = prometheus.NewDesc(
		"nginx_bytes_total",
		"Nginx bytes count",
		[]string{"server_zones", "direction"}, nil)

	cacheDesc = prometheus.NewDesc(
		"nginx_cache_total",
		"Nginx cache count",
		[]string{"server_zones", "type"}, nil)

	connectionsDesc = prometheus.NewDesc(
		"nginx_connections_total",
		"Nginx connections count",
		[]string{"type"}, nil)

	startTimeDesc = prometheus.NewDesc(
		"nginx_oldest_start_time_seconds",
		"start time in seconds since 1970/01/01",
		nil, nil)

	writeBytesDesc = prometheus.NewDesc(
		"nginx_write_bytes_total",
		"number of bytes written",
		nil, nil)

	responseDesc = prometheus.NewDesc(
		"nginx_responses_total",
		"The number of responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
		[]string{"server_zones", "status_code"}, nil)

	requestDesc = prometheus.NewDesc(
		"nginx_requests_total",
		"The total number of requested client connections.",
		[]string{"server_zones"}, nil)

	upstreamBackupDesc = prometheus.NewDesc(
		"nginx_upstream_backup",
		"Current backup setting of the server.",
		[]string{"upstream", "server"}, nil)

	upstreamBytesDesc = prometheus.NewDesc(
		"nginx_upstream_bytes_total",
		"The total number of bytes sent to this server.",
		[]string{"upstream", "server", "direction"}, nil)

	upstreamDownDesc = prometheus.NewDesc(
		"nginx_upstream_down_total",
		"Current down setting of the server.",
		[]string{"upstream", "server"}, nil)

	upstreamFailTimeoutDesc = prometheus.NewDesc(
		"nginx_upstream_fail_timeout",
		"Current fail_timeout setting of the server.",
		[]string{"upstream", "server"}, nil)

	upstreamMaxFailsDesc = prometheus.NewDesc(
		"nginx_upstream_maxfails",
		"Current max_fails setting of the server.",
		[]string{"upstream", "server"}, nil)

	upstreamResponsesDesc = prometheus.NewDesc(
		"nginx_upstream_responses_total",
		"The number of upstream responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
		[]string{"upstream", "server", "status_code"}, nil)

	upstreamRequestDesc = prometheus.NewDesc(
		"nginx_upstream_requests_total",
		"The total number of client connections forwarded to this server.",
		[]string{"upstream", "server"}, nil)

	upstreamResponseMsecDesc = prometheus.NewDesc(
		"nginx_upstream_response_msecs_avg",
		"The average of only upstream response processing times in milliseconds.",
		[]string{"upstream", "server"}, nil)

	upstreamWeightDesc = prometheus.NewDesc(
		"nginx_upstream_weight",
		"Current upstream weight setting of the server.",
		[]string{"upstream", "server"}, nil)
)

type (
	scrapeRequest struct {
		results chan<- prometheus.Metric
		done    chan struct{}
	}

	namedProcessCollector struct {
		scrapeChan chan scrapeRequest
		*proc.Grouper
		fs *proc.FS
	}
)

func newProcessCollector(
	children bool,
	n common.MatchNamer) (*namedProcessCollector, error) {

	fs, err := proc.NewFS("/proc")
	if err != nil {
		return nil, err
	}
	p := &namedProcessCollector{
		scrapeChan: make(chan scrapeRequest),
		Grouper:    proc.NewGrouper(children, n),
		fs:         fs,
	}
	_, err = p.Update(p.fs.AllProcs())
	if err != nil {
		return nil, err
	}

	go p.start()

	return p, nil
}

// Describe implements prometheus.Collector.
func (p *namedProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cpuSecsDesc
	ch <- memResidentbytesDesc
	ch <- memVirtualbytesDesc
	ch <- startTimeDesc

	ch <- bytesDesc
	ch <- cacheDesc
	ch <- connectionsDesc
	ch <- readBytesDesc
	ch <- requestDesc
	ch <- responseDesc
	ch <- writeBytesDesc
	ch <- upstreamBackupDesc
	ch <- upstreamBytesDesc
	ch <- upstreamDownDesc
	ch <- upstreamFailTimeoutDesc
	ch <- upstreamMaxFailsDesc
	ch <- upstreamRequestDesc
	ch <- upstreamResponseMsecDesc
	ch <- upstreamResponsesDesc
	ch <- upstreamWeightDesc

	ch <- numprocsDesc


}

// Collect implements prometheus.Collector.
func (p *namedProcessCollector) Collect(ch chan<- prometheus.Metric) {
	req := scrapeRequest{results: ch, done: make(chan struct{})}
	p.scrapeChan <- req
	<-req.done
}

func (p *namedProcessCollector) start() {
	for req := range p.scrapeChan {
		ch := req.results
		p.scrape(ch)
		req.done <- struct{}{}
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

func (p *namedProcessCollector) scrape(ch chan<- prometheus.Metric) {

	nginxMetrics, err := getNginxMetrics()
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx status info: %v", err)
		return
	}



	reflectMetrics(&nginxMetrics.Connections, connectionsDesc, ch)

	for name, zones := range nginxMetrics.UpstreamZones {

		for pos, value := range zones {

			reflectMetrics(&zones[pos].Responses, upstreamResponsesDesc, ch, name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamRequestDesc,
				prometheus.CounterValue, float64(zones[pos].RequestCounter), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamDownDesc,
				prometheus.CounterValue, float64(zones[pos].Down), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamWeightDesc,
				prometheus.CounterValue, float64(zones[pos].Weight), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamResponseMsecDesc,
				prometheus.CounterValue, float64(zones[pos].ResponseMsec), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamBackupDesc,
				prometheus.CounterValue, float64(zones[pos].Backup), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamFailTimeoutDesc,
				prometheus.CounterValue, float64(zones[pos].FailTimeout), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamMaxFailsDesc,
				prometheus.CounterValue, float64(zones[pos].MaxFails), name, value.Server)

			ch <- prometheus.MustNewConstMetric(upstreamBytesDesc,
				prometheus.CounterValue, float64(zones[pos].InBytes), name, value.Server, "in")

			ch <- prometheus.MustNewConstMetric(upstreamBytesDesc,
				prometheus.CounterValue, float64(zones[pos].OutBytes), name, value.Server, "out")

		}
	}

	for name, zone := range nginxMetrics.ServerZones {

		reflectMetrics(&zone.Responses, responseDesc, ch, name)

		ch <- prometheus.MustNewConstMetric(requestDesc,
			prometheus.CounterValue, float64(zone.RequestCounter), name)

		ch <- prometheus.MustNewConstMetric(bytesDesc,
			prometheus.CounterValue, float64(zone.InBytes), name, "in")

		ch <- prometheus.MustNewConstMetric(bytesDesc,
			prometheus.CounterValue, float64(zone.OutBytes), name, "out")

		//cache
		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheBypass), name, "bypass")

		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheExpired), name, "expired")

		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheHit), name, "hit")

		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheRevalidated), name, "revalidated")

		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheUpdating), name, "updating")

		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheStale), name, "stale")

		ch <- prometheus.MustNewConstMetric(cacheDesc,
			prometheus.CounterValue, float64(zone.Responses.CacheScarce), name, "scarce")

	}

	_, err = p.Update(p.fs.AllProcs())
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
