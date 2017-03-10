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
	"path/filepath"

	"github.com/golang/glog"
	common "github.com/ncabatoff/process-exporter"
	"github.com/ncabatoff/process-exporter/proc"
	"github.com/prometheus/client_golang/prometheus"
)

type BinaryNameMatcher struct {
	name string
	args []string
}

func (em BinaryNameMatcher) MatchAndName(nacl common.NameAndCmdline) (bool, string) {
	if len(nacl.Cmdline) == 0 {
		return false, ""
	}
	cmd := filepath.Base(nacl.Cmdline[0])
	return em.name == cmd, ""
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
)

type namedProcess struct {
	scrapeChan chan scrapeRequest
	*proc.Grouper
	fs *proc.FS
}

func NewNamedProcessCollector(children bool, mn common.MatchNamer) (prometheus.Collector, error) {
	fs, err := proc.NewFS("/proc")
	if err != nil {
		return nil, err
	}
	p := namedProcess{
		scrapeChan: make(chan scrapeRequest),
		Grouper:    proc.NewGrouper(children, mn),
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
func (p namedProcess) Describe(ch chan<- *prometheus.Desc) {
	ch <- cpuSecsDesc
	ch <- numprocsDesc
	ch <- readBytesDesc
	ch <- writeBytesDesc
	ch <- memResidentbytesDesc
	ch <- memVirtualbytesDesc
	ch <- startTimeDesc
}

// Collect implements prometheus.Collector.
func (p namedProcess) Collect(ch chan<- prometheus.Metric) {
	req := scrapeRequest{results: ch, done: make(chan struct{})}
	p.scrapeChan <- req
	<-req.done
}

func (p namedProcess) start() {
	for req := range p.scrapeChan {
		ch := req.results
		p.scrape(ch)
		req.done <- struct{}{}
	}
}

func (p namedProcess) Stop() {
	close(p.scrapeChan)
}

func (p namedProcess) scrape(ch chan<- prometheus.Metric) {
	_, err := p.Update(p.fs.AllProcs())
	if err != nil {
		glog.Warningf("unexpected error obtaining nginx process info: %v", err)
		return
	}

	for gname, gcounts := range p.Groups() {
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
