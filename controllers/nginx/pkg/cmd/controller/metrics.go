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
		glog.Fatalf("unexpected error registering nginx collector: %v", err)
	}
	err = prometheus.Register(pc)
	if err != nil {
		glog.Warningf("unexpected error registering nginx collector: %v", err)
	}
}

var (
	numprocsDesc = prometheus.NewDesc(
		"nginx_num_procs",
		"number of processes",
		nil, nil)

	cpuSecsDesc = prometheus.NewDesc(
		"nginx_cpu_seconds_total",
		"Cpu usage in seconds",
		nil, nil)

	readBytesDesc = prometheus.NewDesc(
		"nginx_read_bytes_total",
		"number of bytes read",
		nil, nil)

	writeBytesDesc = prometheus.NewDesc(
		"nginx_write_bytes_total",
		"number of bytes written",
		nil, nil)

	memResidentbytesDesc = prometheus.NewDesc(
		"nginx_resident_memory_bytes",
		"number of bytes of memory in use",
		nil, nil)

	memVirtualbytesDesc = prometheus.NewDesc(
		"nginx_virtual_memory_bytes",
		"number of bytes of memory in use",
		nil, nil)

	startTimeDesc = prometheus.NewDesc(
		"nginx_oldest_start_time_seconds",
		"start time in seconds since 1970/01/01",
		nil, nil)

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
	ch <- numprocsDesc
	ch <- readBytesDesc
	ch <- writeBytesDesc
	ch <- memResidentbytesDesc
	ch <- memVirtualbytesDesc
	ch <- startTimeDesc
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

func (p *namedProcessCollector) scrape(ch chan<- prometheus.Metric) {
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
