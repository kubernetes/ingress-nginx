/*
Copyright 2022 The Kubernetes Authors.

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

package metrics

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apiserver/pkg/server/healthz"
	klog "k8s.io/klog/v2"
)

func RegisterHealthz(healthPath string, mux *http.ServeMux, checks ...healthz.HealthChecker) {

	healthCheck := []healthz.HealthChecker{healthz.PingHealthz}
	if len(checks) > 0 {
		healthCheck = append(healthCheck, checks...)
	}
	// expose health check endpoint (/healthz)
	healthz.InstallPathHandler(mux,
		healthPath,
		healthCheck...,
	)
}

func RegisterMetrics(reg *prometheus.Registry, mux *http.ServeMux) {
	mux.Handle(
		"/metrics",
		promhttp.InstrumentMetricHandler(
			reg,
			promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		),
	)
}

func RegisterProfiler(host string, port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", host, port),
		//G112 (CWE-400): Potential Slowloris Attack
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           mux,
	}
	klog.Fatal(server.ListenAndServe())
}

func StartHTTPServer(host string, port int, mux *http.ServeMux) {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%v", host, port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	klog.Fatal(server.ListenAndServe())
}
