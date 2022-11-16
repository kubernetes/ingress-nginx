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
	"fmt"
	"math/rand" // #nosec
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/controller"
	"k8s.io/ingress-nginx/internal/ingress/metric"
	"k8s.io/ingress-nginx/internal/nginx"
	ingressflags "k8s.io/ingress-nginx/pkg/flags"
	"k8s.io/ingress-nginx/pkg/metrics"
	"k8s.io/ingress-nginx/pkg/util/file"
	"k8s.io/ingress-nginx/pkg/util/process"
	"k8s.io/ingress-nginx/version"
)

func main() {
	klog.InitFlags(nil)

	rand.Seed(time.Now().UnixNano())

	fmt.Println(version.String())
	var err error
	showVersion, conf, err := ingressflags.ParseFlags()
	if showVersion {
		os.Exit(0)
	}

	if err != nil {
		klog.Fatal(err)
	}

	err = file.CreateRequiredDirectories()
	if err != nil {
		klog.Fatal(err)
	}

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		PidFn:        func() (int, error) { return os.Getpid(), nil },
		ReportErrors: true,
	}))

	mc := metric.NewDummyCollector()
	if conf.EnableMetrics {
		// TODO: Ingress class is not a part of dataplane anymore
		mc, err = metric.NewCollector(conf.MetricsPerHost, conf.ReportStatusClasses, reg, conf.IngressClassConfiguration.Controller, *conf.MetricsBuckets)
		if err != nil {
			klog.Fatalf("Error creating prometheus collector:  %v", err)
		}
	}
	// Pass the ValidationWebhook status to determine if we need to start the collector
	// for the admissionWebhook
	// TODO: Dataplane does not contain validation webhook so the MetricCollector should not receive
	// this as an argument
	mc.Start(conf.ValidationWebhook)

	if conf.EnableProfiling {
		go metrics.RegisterProfiler(nginx.ProfilerAddress, nginx.ProfilerPort)
	}

	ngx := controller.NewNGINXController(conf, mc)

	mux := http.NewServeMux()
	metrics.RegisterHealthz(nginx.HealthPath, mux)
	metrics.RegisterMetrics(reg, mux)

	go metrics.StartHTTPServer(conf.HealthCheckHost, conf.ListenPorts.Health, mux)
	go ngx.Start()

	process.HandleSigterm(ngx, conf.PostShutdownGracePeriod, func(code int) {
		os.Exit(code)
	})
}
