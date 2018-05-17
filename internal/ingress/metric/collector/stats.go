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

package collector

import (
	"reflect"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

const ns = "nginx"

type stat struct {
	Host   string `json:"host"`
	Status int    `json:"status"`

	Time time.Duration `json:"time"`

	RemoteAddress string `json:"remoteAddr"`
	RemoteUser    string `json:"remoteUser"`

	BytesSent int64 `json:"bytesSent"`

	Protocol string `json:"protocol"`
	Method   string `json:"method"`
	Path     string `json:"path"`

	RequestTime   string `json:"requestTime"`
	RequestLength string `json:"requestLength"`
	Duration      int    `json:"duration"`

	UpstreamName         string `json:"upstreamName"`
	UpstreamIP           string `json:"upstreamIP"`
	UpstreamResponseTime string `json:"upstreamResponseTime"`
	UpstreamStatus       string `json:"upstreamStatus"`

	Namespace string `json:"namespace"`
	Ingress   string `json:"ingress"`
	Service   string `json:"service"`
}

type data struct {
	bytes                *prometheus.Desc
	cache                *prometheus.Desc
	connections          *prometheus.Desc
	responses            *prometheus.Desc
	requests             *prometheus.Desc
	filterZoneBytes      *prometheus.Desc
	filterZoneResponses  *prometheus.Desc
	filterZoneCache      *prometheus.Desc
	upstreamBackup       *prometheus.Desc
	upstreamBytes        *prometheus.Desc
	upstreamDown         *prometheus.Desc
	upstreamFailTimeout  *prometheus.Desc
	upstreamMaxFails     *prometheus.Desc
	upstreamResponses    *prometheus.Desc
	upstreamRequests     *prometheus.Desc
	upstreamResponseMsec *prometheus.Desc
	upstreamWeight       *prometheus.Desc
}

type statsCollector struct {
	process prometheus.Collector

	namespace  string
	watchClass string

	port int

	promData *data
}

func (sc *statsCollector) HandleMessage(msg []byte) {
	/*
		reflectMetrics(&nginxMetrics.Connections, p.data.connections, ch, p.ingressClass, p.namespace)

		for name, zones := range nginxMetrics.UpstreamZones {
			for pos, value := range zones {
				reflectMetrics(&zones[pos].Responses, p.data.upstreamResponses, ch, p.ingressClass, p.namespace, name, value.Server)

				ch <- prometheus.MustNewConstMetric(p.data.upstreamRequests,
					prometheus.CounterValue, zones[pos].RequestCounter, p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamDown,
					prometheus.CounterValue, float64(zones[pos].Down), p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamWeight,
					prometheus.CounterValue, zones[pos].Weight, p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamResponseMsec,
					prometheus.CounterValue, zones[pos].ResponseMsec, p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamBackup,
					prometheus.CounterValue, float64(zones[pos].Backup), p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamFailTimeout,
					prometheus.CounterValue, zones[pos].FailTimeout, p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamMaxFails,
					prometheus.CounterValue, zones[pos].MaxFails, p.ingressClass, p.namespace, name, value.Server)
				ch <- prometheus.MustNewConstMetric(p.data.upstreamBytes,
					prometheus.CounterValue, zones[pos].InBytes, p.ingressClass, p.namespace, name, value.Server, "in")
				ch <- prometheus.MustNewConstMetric(p.data.upstreamBytes,
					prometheus.CounterValue, zones[pos].OutBytes, p.ingressClass, p.namespace, name, value.Server, "out")
			}
		}

		for name, zone := range nginxMetrics.ServerZones {
			reflectMetrics(&zone.Responses, p.data.responses, ch, p.ingressClass, p.namespace, name)
			reflectMetrics(&zone.Cache, p.data.cache, ch, p.ingressClass, p.namespace, name)

			ch <- prometheus.MustNewConstMetric(p.data.requests,
				prometheus.CounterValue, zone.RequestCounter, p.ingressClass, p.namespace, name)
			ch <- prometheus.MustNewConstMetric(p.data.bytes,
				prometheus.CounterValue, zone.InBytes, p.ingressClass, p.namespace, name, "in")
			ch <- prometheus.MustNewConstMetric(p.data.bytes,
				prometheus.CounterValue, zone.OutBytes, p.ingressClass, p.namespace, name, "out")
		}

		for serverZone, keys := range nginxMetrics.FilterZones {
			for name, zone := range keys {
				reflectMetrics(&zone.Responses, p.data.filterZoneResponses, ch, p.ingressClass, p.namespace, serverZone, name)
				reflectMetrics(&zone.Cache, p.data.filterZoneCache, ch, p.ingressClass, p.namespace, serverZone, name)

				ch <- prometheus.MustNewConstMetric(p.data.filterZoneBytes,
					prometheus.CounterValue, zone.InBytes, p.ingressClass, p.namespace, serverZone, name, "in")
				ch <- prometheus.MustNewConstMetric(p.data.filterZoneBytes,
					prometheus.CounterValue, zone.OutBytes, p.ingressClass, p.namespace, serverZone, name, "out")
			}
		}
	*/
}

func reflectMetrics(value interface{}, desc *prometheus.Desc, ch chan<- prometheus.Metric, labels ...string) {
	val := reflect.ValueOf(value).Elem()

	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag
		l := append(labels, tag.Get("json"))
		ch <- prometheus.MustNewConstMetric(desc,
			prometheus.CounterValue, val.Field(i).Interface().(float64),
			l...)
	}
}

func NewInstance(ns, class, binary string, port int) (*statsCollector, error) {
	glog.Infof("starting new nginx stats collector for Ingress controller running in namespace %v (class %v)", ns, class)
	glog.Infof("collector extracting information from port %v", port)
	pc, err := newNamedProcess(true, BinaryNameMatcher{
		Name:   "nginx",
		Binary: binary,
	})
	if err != nil {
		glog.Fatalf("unexpected error registering nginx collector: %v", err)
	}
	err = prometheus.Register(pc)
	if err != nil {
		glog.Fatalf("unexpected error registering nginx collector: %v", err)
	}

	promData := &data{
		bytes: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "bytes_total"),
			"Nginx bytes count",
			[]string{"ingress_class", "namespace", "server_zone", "direction"}, nil),

		cache: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "cache_total"),
			"Nginx cache count",
			[]string{"ingress_class", "namespace", "server_zone", "type"}, nil),

		connections: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "connections_total"),
			"Nginx connections count",
			[]string{"ingress_class", "namespace", "type"}, nil),

		responses: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "responses_total"),
			"The number of responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
			[]string{"ingress_class", "namespace", "server_zone", "status_code"}, nil),

		requests: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "requests_total"),
			"The total number of requested client connections.",
			[]string{"ingress_class", "namespace", "server_zone"}, nil),

		filterZoneBytes: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "filterzone_bytes_total"),
			"Nginx bytes count",
			[]string{"ingress_class", "namespace", "server_zone", "key", "direction"}, nil),

		filterZoneResponses: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "filterzone_responses_total"),
			"The number of responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
			[]string{"ingress_class", "namespace", "server_zone", "key", "status_code"}, nil),

		filterZoneCache: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "filterzone_cache_total"),
			"Nginx cache count",
			[]string{"ingress_class", "namespace", "server_zone", "key", "type"}, nil),

		upstreamBackup: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_backup"),
			"Current backup setting of the server.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),

		upstreamBytes: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_bytes_total"),
			"The total number of bytes sent to this server.",
			[]string{"ingress_class", "namespace", "upstream", "server", "direction"}, nil),

		upstreamDown: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "vts_upstream_down_total"),
			"Current down setting of the server.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),

		upstreamFailTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_fail_timeout"),
			"Current fail_timeout setting of the server.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),

		upstreamMaxFails: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_maxfails"),
			"Current max_fails setting of the server.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),

		upstreamResponses: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_responses_total"),
			"The number of upstream responses with status codes 1xx, 2xx, 3xx, 4xx, and 5xx.",
			[]string{"ingress_class", "namespace", "upstream", "server", "status_code"}, nil),

		upstreamRequests: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_requests_total"),
			"The total number of client connections forwarded to this server.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),

		upstreamResponseMsec: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_response_msecs_avg"),
			"The average of only upstream response processing times in milliseconds.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),

		upstreamWeight: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "upstream_weight"),
			"Current upstream weight setting of the server.",
			[]string{"ingress_class", "namespace", "upstream", "server"}, nil),
	}

	sc := &statsCollector{
		namespace:  ns,
		watchClass: class,
		process:    pc,
		port:       port,
		promData:   promData,
	}

	listener, err := newUDPListener(port)
	if err != nil {
		return nil, err
	}

	go handleMessages(listener, sc.HandleMessage)

	return sc, nil
}
