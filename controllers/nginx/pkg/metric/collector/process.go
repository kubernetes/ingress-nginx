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

// BinaryNameMatcher ...
type BinaryNameMatcher struct {
	Name   string
	Binary string
}

// MatchAndName returns false if the match failed, otherwise
// true and the resulting name.
func (em BinaryNameMatcher) MatchAndName(nacl common.NameAndCmdline) (bool, string) {
	if len(nacl.Cmdline) == 0 {
		return false, ""
	}
	cmd := filepath.Base(em.Binary)
	return em.Name == cmd, ""
}

type namedProcessData struct {
	numProcs         *prometheus.Desc
	cpuSecs          *prometheus.Desc
	readBytes        *prometheus.Desc
	writeBytes       *prometheus.Desc
	memResidentbytes *prometheus.Desc
	memVirtualbytes  *prometheus.Desc
	startTime        *prometheus.Desc
}

type namedProcess struct {
	*proc.Grouper

	scrapeChan chan scrapeRequest
	fs         *proc.FS
	data       namedProcessData
}

// NewNamedProcess returns a new prometheus collector for the nginx process
func NewNamedProcess(children bool, mn common.MatchNamer) (prometheus.Collector, error) {
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

	p.data = namedProcessData{
		numProcs: prometheus.NewDesc(
			"num_procs",
			"number of processes",
			nil, nil),

		cpuSecs: prometheus.NewDesc(
			"cpu_seconds_total",
			"Cpu usage in seconds",
			nil, nil),

		readBytes: prometheus.NewDesc(
			"read_bytes_total",
			"number of bytes read",
			nil, nil),

		writeBytes: prometheus.NewDesc(
			"write_bytes_total",
			"number of bytes written",
			nil, nil),

		memResidentbytes: prometheus.NewDesc(
			"resident_memory_bytes",
			"number of bytes of memory in use",
			nil, nil),

		memVirtualbytes: prometheus.NewDesc(
			"virtual_memory_bytes",
			"number of bytes of memory in use",
			nil, nil),

		startTime: prometheus.NewDesc(
			"oldest_start_time_seconds",
			"start time in seconds since 1970/01/01",
			nil, nil),
	}

	go p.start()

	return p, nil
}

// Describe implements prometheus.Collector.
func (p namedProcess) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.data.cpuSecs
	ch <- p.data.numProcs
	ch <- p.data.readBytes
	ch <- p.data.writeBytes
	ch <- p.data.memResidentbytes
	ch <- p.data.memVirtualbytes
	ch <- p.data.startTime
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

	for _, gcounts := range p.Groups() {
		ch <- prometheus.MustNewConstMetric(p.data.numProcs,
			prometheus.GaugeValue, float64(gcounts.Procs))
		ch <- prometheus.MustNewConstMetric(p.data.memResidentbytes,
			prometheus.GaugeValue, float64(gcounts.Memresident))
		ch <- prometheus.MustNewConstMetric(p.data.memVirtualbytes,
			prometheus.GaugeValue, float64(gcounts.Memvirtual))
		ch <- prometheus.MustNewConstMetric(p.data.startTime,
			prometheus.GaugeValue, float64(gcounts.OldestStartTime.Unix()))
		ch <- prometheus.MustNewConstMetric(p.data.cpuSecs,
			prometheus.CounterValue, gcounts.Cpu)
		ch <- prometheus.MustNewConstMetric(p.data.readBytes,
			prometheus.CounterValue, float64(gcounts.ReadBytes))
		ch <- prometheus.MustNewConstMetric(p.data.writeBytes,
			prometheus.CounterValue, float64(gcounts.WriteBytes))
	}
}
