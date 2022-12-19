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
	"flag"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/metric/collectors"
	"k8s.io/ingress-nginx/pkg/dataplane"
	"k8s.io/ingress-nginx/pkg/util/maxmind"
	"k8s.io/klog/v2"
)

// ParseDataplaneFlags is the function to get the required config structure from passed flags
func ParseDataplaneFlags() (bool, *dataplane.Configuration, error) {

	var (
		flags       = pflag.NewFlagSet("", pflag.ExitOnError)
		grpcAddress = flags.String("grpc-host", "ingress-nginx:10000", "Address to connect to gRPC Control plane")

		showVersion = flags.Bool("version", false,
			`Show release information about the NGINX Ingress controller and exit.`)

		// Ports:
		httpPort      = flags.Int("http-port", 80, `Port to use for servicing HTTP traffic.`)
		httpsPort     = flags.Int("https-port", 443, `Port to use for servicing HTTPS traffic.`)
		sslProxyPort  = flags.Int("ssl-passthrough-proxy-port", 442, `Port to use internally for SSL Passthrough.`)
		defServerPort = flags.Int("default-server-port", 8181, `Port to use for exposing the default server (catch-all).`)
		healthzPort   = flags.Int("healthz-port", 10254, "Port to use for the healthz endpoint.")

		enableMetrics = flags.Bool("enable-metrics", true,
			`Enables the collection of NGINX metrics`)
		metricsPerHost = flags.Bool("metrics-per-host", true,
			`Export metrics per-host`)
		reportStatusClasses = flags.Bool("report-status-classes", false,
			`Use status classes (2xx, 3xx, 4xx and 5xx) instead of status codes in metrics`)

		timeBuckets         = flags.Float64Slice("time-buckets", prometheus.DefBuckets, "Set of buckets which will be used for prometheus histogram metrics such as RequestTime, ResponseTime")
		lengthBuckets       = flags.Float64Slice("length-buckets", prometheus.LinearBuckets(10, 10, 10), "Set of buckets which will be used for prometheus histogram metrics such as RequestLength, ResponseLength")
		sizeBuckets         = flags.Float64Slice("size-buckets", prometheus.ExponentialBuckets(10, 10, 7), "Set of buckets which will be used for prometheus histogram metrics such as BytesSent")
		monitorMaxBatchSize = flags.Int("monitor-max-batch-size", 10000, "Max batch size of NGINX metrics")

		maxmindEditionIDs = flags.String("maxmind-edition-ids", "GeoLite2-City,GeoLite2-ASN", `Maxmind edition ids to download GeoLite2 Databases.`)

		maxmindMirror         = flags.String("maxmind-mirror", "", `Maxmind mirror url (example: http://geoip.local/databases`)
		maxmindLicenseKey     = flags.String("maxmind-license-key", "", `Maxmind license key to download GeoLite2 Databases. https://blog.maxmind.com/2019/12/18/significant-changes-to-accessing-and-using-geolite2-databases`)
		maxmindRetriesCount   = flags.Int("maxmind-retries-count", 1, "Number of attempts to download the GeoIP DB.")
		maxmindRetriesTimeout = flags.Duration("maxmind-retries-timeout", time.Second*0, "Maxmind downloading delay between 1st and 2nd attempt, 0s - do not retry to download if something went wrong.")

		/*healthzPort = flags.Int("healthz-port", 10254, "Port to use for the healthz endpoint.")
		healthzHost = flags.String("healthz-host", "", "Address to bind the healthz endpoint.")*/
	)

	flags.Parse(os.Args)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	pflag.VisitAll(func(flag *pflag.Flag) {
		klog.V(2).InfoS("FLAG", flag.Name, flag.Value)
	})

	if *showVersion {
		return true, nil, nil
	}

	var histogramBuckets = &collectors.HistogramBuckets{
		TimeBuckets:   *timeBuckets,
		LengthBuckets: *lengthBuckets,
		SizeBuckets:   *sizeBuckets,
	}

	var err error
	config := &dataplane.Configuration{
		EnableMetrics:       *enableMetrics,
		MetricsPerHost:      *metricsPerHost,
		MetricsBuckets:      histogramBuckets,
		ReportStatusClasses: *reportStatusClasses,
		MonitorMaxBatchSize: *monitorMaxBatchSize,
		GRPCAddress:         *grpcAddress,
		ListenPorts: &ngx_config.ListenPorts{
			Default:  *defServerPort,
			Health:   *healthzPort,
			HTTP:     *httpPort,
			HTTPS:    *httpsPort,
			SSLProxy: *sslProxyPort,
		},
	}

	if *maxmindEditionIDs != "" {
		maxmindConfig := maxmind.Config{
			EditionIDs:     *maxmindEditionIDs,
			LicenseKey:     *maxmindLicenseKey,
			Mirror:         *maxmindMirror,
			RetriesCount:   *maxmindRetriesCount,
			RetriesTimeout: *maxmindRetriesTimeout,
		}
		files, err := maxmind.BootstrapMaxmindFiles(maxmindConfig)
		if err != nil {
			klog.ErrorS(err, "failed bootstrapping maxmind files")
			return false, nil, err
		}

		config.MaxmindEditionFiles = &files
		config.MaxMindEditionIDs = *maxmindEditionIDs
	}

	return false, config, err
}
