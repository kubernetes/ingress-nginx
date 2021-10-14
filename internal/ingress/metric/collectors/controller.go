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

package collectors

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/version"
	"k8s.io/klog/v2"
)

var (
	operation        = []string{"controller_namespace", "controller_class", "controller_pod"}
	ingressOperation = []string{"controller_namespace", "controller_class", "controller_pod", "namespace", "ingress"}
	sslLabelHost     = []string{"namespace", "class", "host"}
)

// Controller defines base metrics about the ingress controller
type Controller struct {
	prometheus.Collector

	configHash        prometheus.Gauge
	configSuccess     prometheus.Gauge
	configSuccessTime prometheus.Gauge

	reloadOperation             *prometheus.CounterVec
	reloadOperationErrors       *prometheus.CounterVec
	checkIngressOperation       *prometheus.CounterVec
	checkIngressOperationErrors *prometheus.CounterVec
	sslExpireTime               *prometheus.GaugeVec

	constLabels prometheus.Labels
	labels      prometheus.Labels

	leaderElection *prometheus.GaugeVec

	buildInfo prometheus.Collector
}

// NewController creates a new prometheus collector for the
// Ingress controller operations
func NewController(pod, namespace, class string) *Controller {
	constLabels := prometheus.Labels{
		"controller_namespace": namespace,
		"controller_class":     class,
		"controller_pod":       pod,
	}

	cm := &Controller{
		constLabels: constLabels,

		labels: prometheus.Labels{
			"namespace": namespace,
			"class":     class,
		},

		buildInfo: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: PrometheusNamespace,
				Name:      "build_info",
				Help:      "A metric with a constant '1' labeled with information about the build.",
				ConstLabels: prometheus.Labels{
					"controller_namespace": namespace,
					"controller_class":     class,
					"controller_pod":       pod,
					"release":              version.RELEASE,
					"build":                version.COMMIT,
					"repository":           version.REPO,
				},
			},
			func() float64 { return 1 },
		),

		configHash: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   PrometheusNamespace,
				Name:        "config_hash",
				Help:        "Running configuration hash actually running",
				ConstLabels: constLabels,
			},
		),
		configSuccess: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   PrometheusNamespace,
				Name:        "config_last_reload_successful",
				Help:        "Whether the last configuration reload attempt was successful",
				ConstLabels: constLabels,
			}),
		configSuccessTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   PrometheusNamespace,
				Name:        "config_last_reload_successful_timestamp_seconds",
				Help:        "Timestamp of the last successful configuration reload.",
				ConstLabels: constLabels,
			}),
		reloadOperation: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: PrometheusNamespace,
				Name:      "success",
				Help:      `Cumulative number of Ingress controller reload operations`,
			},
			operation,
		),
		reloadOperationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: PrometheusNamespace,
				Name:      "errors",
				Help:      `Cumulative number of Ingress controller errors during reload operations`,
			},
			operation,
		),
		checkIngressOperationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: PrometheusNamespace,
				Name:      "check_errors",
				Help:      `Cumulative number of Ingress controller errors during syntax check operations`,
			},
			ingressOperation,
		),
		checkIngressOperation: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: PrometheusNamespace,
				Name:      "check_success",
				Help:      `Cumulative number of Ingress controller syntax check operations`,
			},
			ingressOperation,
		),
		sslExpireTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: PrometheusNamespace,
				Name:      "ssl_expire_time_seconds",
				Help: `Number of seconds since 1970 to the SSL Certificate expire.
			An example to check if this certificate will expire in 10 days is: "nginx_ingress_controller_ssl_expire_time_seconds < (time() + (10 * 24 * 3600))"`,
			},
			sslLabelHost,
		),
		leaderElection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   PrometheusNamespace,
				Name:        "leader_election_status",
				Help:        "Gauge reporting status of the leader election, 0 indicates follower, 1 indicates leader. 'name' is the string used to identify the lease",
				ConstLabels: constLabels,
			},
			[]string{"name"},
		),
	}

	return cm
}

// IncReloadCount increment the reload counter
func (cm *Controller) IncReloadCount() {
	cm.reloadOperation.With(cm.constLabels).Inc()
}

// IncReloadErrorCount increment the reload error counter
func (cm *Controller) IncReloadErrorCount() {
	cm.reloadOperationErrors.With(cm.constLabels).Inc()
}

// OnStartedLeading indicates the pod was elected as the leader
func (cm *Controller) OnStartedLeading(electionID string) {
	cm.leaderElection.WithLabelValues(electionID).Set(1.0)
}

// OnStoppedLeading indicates the pod stopped being the leader
func (cm *Controller) OnStoppedLeading(electionID string) {
	cm.leaderElection.WithLabelValues(electionID).Set(0)
}

// IncCheckCount increment the check counter
func (cm *Controller) IncCheckCount(namespace, name string) {
	labels := prometheus.Labels{
		"namespace": namespace,
		"ingress":   name,
	}
	cm.checkIngressOperation.MustCurryWith(cm.constLabels).With(labels).Inc()
}

// IncCheckErrorCount increment the check error counter
func (cm *Controller) IncCheckErrorCount(namespace, name string) {
	labels := prometheus.Labels{
		"namespace": namespace,
		"ingress":   name,
	}
	cm.checkIngressOperationErrors.MustCurryWith(cm.constLabels).With(labels).Inc()
}

// ConfigSuccess set a boolean flag according to the output of the controller configuration reload
func (cm *Controller) ConfigSuccess(hash uint64, success bool) {
	if success {
		cm.configSuccessTime.Set(float64(time.Now().Unix()))
		cm.configSuccess.Set(1)

		cm.configHash.Set(float64(hash))

		return
	}

	cm.configSuccess.Set(0)
	cm.configHash.Set(0)
}

// Describe implements prometheus.Collector
func (cm Controller) Describe(ch chan<- *prometheus.Desc) {
	cm.configHash.Describe(ch)
	cm.configSuccess.Describe(ch)
	cm.configSuccessTime.Describe(ch)
	cm.reloadOperation.Describe(ch)
	cm.reloadOperationErrors.Describe(ch)
	cm.checkIngressOperation.Describe(ch)
	cm.checkIngressOperationErrors.Describe(ch)
	cm.sslExpireTime.Describe(ch)
	cm.leaderElection.Describe(ch)
	cm.buildInfo.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (cm Controller) Collect(ch chan<- prometheus.Metric) {
	cm.configHash.Collect(ch)
	cm.configSuccess.Collect(ch)
	cm.configSuccessTime.Collect(ch)
	cm.reloadOperation.Collect(ch)
	cm.reloadOperationErrors.Collect(ch)
	cm.checkIngressOperation.Collect(ch)
	cm.checkIngressOperationErrors.Collect(ch)
	cm.sslExpireTime.Collect(ch)
	cm.leaderElection.Collect(ch)
	cm.buildInfo.Collect(ch)
}

// SetSSLExpireTime sets the expiration time of SSL Certificates
func (cm *Controller) SetSSLExpireTime(servers []*ingress.Server) {
	for _, s := range servers {
		if s.Hostname != "" && s.SSLCert != nil && s.SSLCert.ExpireTime.Unix() > 0 {
			labels := make(prometheus.Labels, len(cm.labels)+1)
			for k, v := range cm.labels {
				labels[k] = v
			}
			labels["host"] = s.Hostname

			cm.sslExpireTime.With(labels).Set(float64(s.SSLCert.ExpireTime.Unix()))
		}
	}
}

// RemoveMetrics removes metrics for hostnames not available anymore
func (cm *Controller) RemoveMetrics(hosts []string, registry prometheus.Gatherer) {
	cm.removeSSLExpireMetrics(true, hosts, registry)
}

// RemoveAllSSLExpireMetrics removes metrics for expiration of SSL Certificates
func (cm *Controller) RemoveAllSSLExpireMetrics(registry prometheus.Gatherer) {
	cm.removeSSLExpireMetrics(false, []string{}, registry)
}

func (cm *Controller) removeSSLExpireMetrics(onlyDefinedHosts bool, hosts []string, registry prometheus.Gatherer) {
	mfs, err := registry.Gather()
	if err != nil {
		klog.ErrorS(err, "Error gathering metrics")
		return
	}

	toRemove := sets.NewString(hosts...)

	for _, mf := range mfs {
		metricName := mf.GetName()
		if fmt.Sprintf("%v_ssl_expire_time_seconds", PrometheusNamespace) != metricName {
			continue
		}

		for _, m := range mf.GetMetric() {
			labels := make(map[string]string, len(m.GetLabel()))
			for _, labelPair := range m.GetLabel() {
				labels[*labelPair.Name] = *labelPair.Value
			}

			// remove labels that are constant
			deleteConstants(labels)

			host, ok := labels["host"]
			if !ok {
				continue
			}

			if onlyDefinedHosts && !toRemove.Has(host) {
				continue
			}

			klog.V(2).InfoS("Removing prometheus metric", "gauge", metricName, "host", host)
			removed := cm.sslExpireTime.Delete(labels)
			if !removed {
				klog.V(2).InfoS("metric removed", "metric", metricName, "host", host, "labels", labels)
			}
		}
	}
}
