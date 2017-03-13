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

import "github.com/prometheus/client_golang/prometheus"

// Stopable defines a prometheus collector that can be stopped
type Stopable interface {
	prometheus.Collector
	Stop()
}

type scrapeRequest struct {
	results chan<- prometheus.Metric
	done    chan struct{}
}
