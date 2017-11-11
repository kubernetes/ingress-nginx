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
}

var (
	reloadOperation = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "success",
			Help:      "Cumulative number of Ingress controller reload operations",
		},
		[]string{operation},
	)
	reloadOperationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "errors",
			Help:      "Cumulative number of Ingress controller errors during reload operations",
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

func incReloadCount() {
	reloadOperation.WithLabelValues(reloadLabel).Inc()
}

func incReloadErrorCount() {
	reloadOperationErrors.WithLabelValues(reloadLabel).Inc()
}

func setSSLExpireTime(servers []*ingress.Server) {
	for _, s := range servers {
		if s.Hostname != defServerName {
			sslExpireTime.WithLabelValues(s.Hostname).Set(float64(s.SSLExpireTime.Unix()))
		}
	}
}
