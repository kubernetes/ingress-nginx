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

package collectors

import (
	"fmt"
	"path/filepath"

	"k8s.io/klog"

	common "github.com/ncabatoff/process-exporter"
	"github.com/ncabatoff/process-exporter/proc"
	"github.com/prometheus/client_golang/prometheus"
)

type scrapeRequest struct {
	results chan<- prometheus.Metric
	done    chan struct{}
}

// Stopable defines a prometheus collector that can be stopped
type Stopable interface {
	prometheus.Collector
	Stop()
}

func newBinaryNameMatcher(name, binary string) common.MatchNamer {
	return BinaryNameMatcher{
		Name:   name,
		Binary: binary,
	}
}

// BinaryNameMatcher define a namer using the binary name
type BinaryNameMatcher struct {
	Name   string
	Binary string
}

// MatchAndName returns false if the match failed, otherwise
// true and the resulting name.
func (em BinaryNameMatcher) MatchAndName(nacl common.ProcAttributes) (bool, string) {
	if len(nacl.Cmdline) == 0 {
		return false, ""
	}
	cmd := filepath.Base(em.Binary)
	return em.Name == cmd, ""
}

// String returns the name of the binary to match
func (em BinaryNameMatcher) String() string {
	return fmt.Sprintf("%+v", em.Binary)
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

const subSystem = "nginx_process"

// NGINXProcessCollector defines a process collector interface
type NGINXProcessCollector interface {
	prometheus.Collector

	Start()
	Stop()
}

var name = "nginx"
var binary = "/usr/bin/nginx"

// NewNGINXProcess returns a new prometheus collector for the nginx process
func NewNGINXProcess(pod, namespace, ingressClass string) (NGINXProcessCollector, error) {
	fs, err := proc.NewFS("/proc", false)
	if err != nil {
		return nil, err
	}

	nm := newBinaryNameMatcher(name, binary)

	p := namedProcess{
		scrapeChan: make(chan scrapeRequest),
		Grouper:    proc.NewGrouper(nm, true, false, false, false),
		fs:         fs,
	}

	_, _, err = p.Update(p.fs.AllProcs())
	if err != nil {
		return nil, err
	}

	constLabels := prometheus.Labels{
		"controller_namespace": namespace,
		"controller_class":     ingressClass,
		"controller_pod":       pod,
	}

	p.data = namedProcessData{
		numProcs: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "num_procs"),
			"number of processes",
			nil, constLabels),

		cpuSecs: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "cpu_seconds_total"),
			"Cpu usage in seconds",
			nil, constLabels),

		readBytes: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "read_bytes_total"),
			"number of bytes read",
			nil, constLabels),

		writeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "write_bytes_total"),
			"number of bytes written",
			nil, constLabels),

		memResidentbytes: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "resident_memory_bytes"),
			"number of bytes of memory in use",
			nil, constLabels),

		memVirtualbytes: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "virtual_memory_bytes"),
			"number of bytes of memory in use",
			nil, constLabels),

		startTime: prometheus.NewDesc(
			prometheus.BuildFQName(PrometheusNamespace, subSystem, "oldest_start_time_seconds"),
			"start time in seconds since 1970/01/01",
			nil, constLabels),
	}

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

func (p namedProcess) Start() {
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
	_, groups, err := p.Update(p.fs.AllProcs())
	if err != nil {
		klog.Warningf("unexpected error obtaining nginx process info: %v", err)
		return
	}

	for _, gcounts := range groups {
		ch <- prometheus.MustNewConstMetric(p.data.numProcs,
			prometheus.GaugeValue, float64(gcounts.Procs))
		ch <- prometheus.MustNewConstMetric(p.data.memResidentbytes,
			prometheus.GaugeValue, float64(gcounts.Memory.ResidentBytes))
		ch <- prometheus.MustNewConstMetric(p.data.memVirtualbytes,
			prometheus.GaugeValue, float64(gcounts.Memory.VirtualBytes))
		ch <- prometheus.MustNewConstMetric(p.data.startTime,
			prometheus.GaugeValue, float64(gcounts.OldestStartTime.Unix()))
		ch <- prometheus.MustNewConstMetric(p.data.cpuSecs,
			prometheus.CounterValue, gcounts.CPUSystemTime)
		ch <- prometheus.MustNewConstMetric(p.data.readBytes,
			prometheus.CounterValue, float64(gcounts.ReadBytes))
		ch <- prometheus.MustNewConstMetric(p.data.writeBytes,
			prometheus.CounterValue, float64(gcounts.WriteBytes))
	}
}
