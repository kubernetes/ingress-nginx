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

package dataplane

import (
	"net"
	"sync"
	"time"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	ngxtemplate "k8s.io/ingress-nginx/internal/ingress/controller/template"
	"k8s.io/ingress-nginx/internal/ingress/metric"
	"k8s.io/ingress-nginx/internal/ingress/metric/collectors"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/ingress-nginx/pkg/dataplane/grpcclient"
	"k8s.io/ingress-nginx/pkg/tcpproxy"
)

// Configuration contains all the settings required by an Ingress controller
type Configuration struct {
	ResyncPeriod time.Duration

	HealthCheckHost string

	DisableServiceExternalName bool

	EnableSSLPassthrough bool

	EnableProfiling bool

	EnableMetrics       bool
	MetricsPerHost      bool
	MetricsBuckets      *collectors.HistogramBuckets
	ReportStatusClasses bool

	FakeCertificate *ingress.SSLCert

	SyncRateLimit float32

	MaxWorkerOpenFiles int

	MaxmindEditionFiles *[]string

	MonitorMaxBatchSize int

	PostShutdownGracePeriod int
	ShutdownGracePeriod     int

	DeepInspector bool

	DynamicConfigurationRetries int

	GRPCAddress string

	ListenPorts *ngx_config.ListenPorts
}

// NGINXConfigurer describes a NGINX Ingress controller.
type NGINXConfigurer struct {
	cfg *Configuration

	// configureLock is a mutex to avoid race conditions during configuration that needs to change nginx.conf
	configureLock *sync.Mutex

	// stopLock is used to enforce that only a single call to Stop send at
	// a given time. We allow stopping through an HTTP endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock *sync.Mutex

	stopCh chan struct{}
	// ngxErrCh is used to detect errors with the NGINX processes
	ngxErrCh chan error

	// grpcErrCh is used to detect errors with gRPC connection
	grpcErrCh chan error

	// templateConfig is the config that was received via gRPC and will be applied
	templateConfig *config.TemplateConfig

	t ngxtemplate.Writer

	resolver []net.IP

	isIPV6Enabled bool

	isShuttingDown bool

	Proxy *tcpproxy.TCPProxy

	metricCollector metric.Collector

	GRPCClient *grpcclient.Client

	BacklogSize int

	command NginxExecTester
}
