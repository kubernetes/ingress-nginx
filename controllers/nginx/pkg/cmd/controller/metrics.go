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

func (n *NGINXController) setupMonitor(sm statusModule) {
	csm := n.statusModule
	if csm != sm {
		prometheus
		n.statusModule = sm
	}
}

type statsCollector struct {
	process prometheus.Collector
	basic   prometheus.Collector
	vts     prometheus.Collector
}

func newStatsCollector() (*statsCollector, error) {
	pc, err := collector.NewNamedProcess(true, collector.BinaryNameMatcher{"nginx", n.cmdArgs})
	if err != nil {
		return nil, err
	}
	err = prometheus.Register(pc)
	if err != nil {
		glog.Fatalf("unexpected error registering nginx collector: %v", err)
	}

	return nil, &statsCollector{
		process: pc,
	}
}
