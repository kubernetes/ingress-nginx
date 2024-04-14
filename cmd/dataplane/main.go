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

package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"k8s.io/klog/v2"

	dataplanenginx "k8s.io/ingress-nginx/cmd/dataplane/pkg/nginx"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/metrics"
)

func main() {
	klog.InitFlags(nil)

	//fmt.Println(version.String())
	//var err error

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector())
	// TODO: Below is supported just on Linux, do not register if OS is not Linux
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		PidFn:        func() (int, error) { return os.Getpid(), nil },
		ReportErrors: true,
	}))

	//mc := metric.NewDummyCollector()
	go metrics.RegisterProfiler(nginx.ProfilerAddress, nginx.ProfilerPort)
	
	mux := http.NewServeMux()
	metrics.RegisterHealthz(nginx.HealthPath, mux)
	metrics.RegisterMetrics(reg, mux)

	errCh := make(chan error)
	stopCh := make(chan bool)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	// TODO: Turn delay configurable
	n := dataplanenginx.NewNGINXExecutor(mux, 10, errCh, stopCh)
	//go executor.Start()
	go metrics.StartHTTPServer("127.0.0.1", 12345, mux)

	// TODO: deal with OS signals
	select {
	case err := <- errCh:
		klog.ErrorS(err, "error executing NGINX")
		os.Exit(1)
	
	case <- stopCh:
		klog.Warning("received request to stop")
		os.Exit(0)

	case <- signalChan:
		klog.InfoS("Received SIGTERM, shutting down")
		exitCode := 0
		if err := n.Stop(); err != nil {
			klog.Warningf("Error during sigterm shutdown: %v", err)
			exitCode = 1
		}
		klog.InfoS("Exiting", "code", exitCode)
		os.Exit(exitCode)
	}

}
