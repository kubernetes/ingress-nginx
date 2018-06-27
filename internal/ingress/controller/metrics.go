/*
Copyright 2015 The Kubernetes Authors.

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

package controller

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/ingress-nginx/internal/ingress"
)

const (
	ns             = "ingress_controller"
	operation      = "count"
	reloadLabel    = "reloads"
	sslLabelExpire = "ssl_expire_time_seconds"
	sslLabelHost   = "host"
)

func init() {
	prometheus.MustRegister(reloadOperation)
	prometheus.MustRegister(reloadOperationErrors)
	prometheus.MustRegister(sslExpireTime)
	prometheus.MustRegister(configSuccess)
	prometheus.MustRegister(configSuccessTime)
}

var (
	configSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "config_last_reload_successfull",
		Help: `Whether the last configuration reload attemp was successful.
		Prometheus alert example:
		alert: IngressControllerFailedReload 
		expr: ingress_controller_config_last_reload_successfull == 0
		for: 10m`,
	})
	configSuccessTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "config_last_reload_successfull_timestamp_seconds",
		Help:      "Timestamp of the last successful configuration reload.",
	})
	// TODO depreciate this metrics in favor of ingress_controller_config_last_reload_successfull_timestamp_seconds
	reloadOperation = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "success",
			Help: `DEPRECATED: use ingress_controller_config_last_reload_successfull_timestamp_seconds or ingress_controller_config_last_reload_successfull instead.
			 Cumulative number of Ingress controller reload operations`,
		},
		[]string{operation},
	)
	// TODO depreciate this metrics in favor of ingress_controller_config_last_reload_successfull_timestamp_seconds
	reloadOperationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "errors",
			Help: `DEPRECATED: use ingress_controller_config_last_reload_successfull_timestamp_seconds or ingress_controller_config_last_reload_successfull instead.
			 Cumulative number of Ingress controller errors during reload operations`,
		},
		[]string{operation},
	)
	sslExpireTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      sslLabelExpire,
			Help: "Number of seconds since 1970 to the SSL Certificate expire. An example to check if this " +
				"certificate will expire in 10 days is: \"ingress_controller_ssl_expire_time_seconds < (time() + (10 * 24 * 3600))\"",
		},
		[]string{sslLabelHost},
	)
)

// IncReloadCount increment the reload counter
func IncReloadCount() {
	reloadOperation.WithLabelValues(reloadLabel).Inc()
}

// IncReloadErrorCount increment the reload error counter
func IncReloadErrorCount() {
	reloadOperationErrors.WithLabelValues(reloadLabel).Inc()
}

// ConfigSuccess set a boolean flag according to the output of the controller configuration reload
func ConfigSuccess(success bool) {
	if success {
		ConfigSuccessTime()
		configSuccess.Set(1)
	} else {
		configSuccess.Set(0)
	}
}

// ConfigSuccessTime set the current timestamp when the controller is successfully reloaded
func ConfigSuccessTime() {
	configSuccessTime.Set(float64(time.Now().Unix()))
}

func setSSLExpireTime(servers []*ingress.Server) {
	for _, s := range servers {
		if s.Hostname != defServerName {
			sslExpireTime.WithLabelValues(s.Hostname).Set(float64(s.SSLCert.ExpireTime.Unix()))
		}
	}
}
