package dataplane

import (
	"k8s.io/ingress-nginx/internal/ingress"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/metric/collectors"
)

const (
	emptyUID = "-1"
)

// Configuration contains all the settings required by configurer and NGINX
type Configuration struct {
	HealthCheckHost string
	ListenPorts     *ngx_config.ListenPorts

	DisableServiceExternalName bool

	EnableSSLPassthrough bool

	EnableProfiling bool

	EnableMetrics       bool
	MetricsPerHost      bool
	MetricsBuckets      *collectors.HistogramBuckets
	ReportStatusClasses bool

	FakeCertificate *ingress.SSLCert

	DisableFullValidationTest bool

	MaxmindEditionFiles *[]string

	MonitorMaxBatchSize int

	PostShutdownGracePeriod int
	ShutdownGracePeriod     int

	DynamicConfigurationRetries int
}
