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
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/ingress/controllers/nginx/pkg/metric/collector"
)

const (
	ngxStatusPath = "/internal_nginx_status"
	ngxVtsPath    = "/nginx_status/format/json"
)

func (n *NGINXController) setupMonitor(sm statusModule) {
	csm := n.statusModule
	if csm != sm {
		glog.Infof("changing prometheus collector from %v to %v", csm, sm)
		n.stats.stop(csm)
		n.stats.start(sm)
		n.statusModule = sm
	}
}

type statsCollector struct {
	process prometheus.Collector
	basic   collector.Stopable
	vts     collector.Stopable

	namespace  string
	watchClass string
}

func (s *statsCollector) stop(sm statusModule) {
	switch sm {
	case defaultStatusModule:
		s.basic.Stop()
		prometheus.Unregister(s.basic)
		break
	case vtsStatusModule:
		s.vts.Stop()
		prometheus.Unregister(s.vts)
		break
	}
}

func (s *statsCollector) start(sm statusModule) {
	switch sm {
	case defaultStatusModule:
		s.basic = collector.NewNginxStatus(s.namespace, s.watchClass, ngxHealthPort, ngxStatusPath)
		prometheus.Register(s.basic)
		break
	case vtsStatusModule:
		s.vts = collector.NewNGINXVTSCollector(s.namespace, s.watchClass, ngxHealthPort, ngxVtsPath)
		prometheus.Register(s.vts)
		break
	}
}

func newStatsCollector(ns, class, binary string) *statsCollector {
	glog.Infof("starting new nginx stats collector for Ingress controller running in namespace %v (class %v)", ns, class)
	pc, err := collector.NewNamedProcess(true, collector.BinaryNameMatcher{
		Name:   "nginx",
		Binary: binary,
	})
	if err != nil {
		glog.Fatalf("unexpected error registering nginx collector: %v", err)
	}
	err = prometheus.Register(pc)
	if err != nil {
		glog.Fatalf("unexpected error registering nginx collector: %v", err)
	}

	return &statsCollector{
		namespace:  ns,
		watchClass: class,
		process:    pc,
	}
}
