/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/controller"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
)

func parseFlags() (bool, *controller.Configuration, error) {
	var (
		flags = pflag.NewFlagSet("", pflag.ExitOnError)

		apiserverHost = flags.String("apiserver-host", "", "The address of the Kubernetes Apiserver "+
			"to connect to in the format of protocol://address:port, e.g., "+
			"http://localhost:8080. If not specified, the assumption is that the binary runs inside a "+
			"Kubernetes cluster and local discovery is attempted.")
		kubeConfigFile = flags.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")

		defaultSvc = flags.String("default-backend-service", "",
			`Service used to serve a 404 page for the default backend. Takes the form
		namespace/name. The controller uses the first node port of this Service for
		the default backend.`)

		ingressClass = flags.String("ingress-class", "",
			`Name of the ingress class to route through this controller.`)

		configMap = flags.String("configmap", "",
			`Name of the ConfigMap that contains the custom configuration to use`)

		publishSvc = flags.String("publish-service", "",
			`Service fronting the ingress controllers. Takes the form namespace/name.
		The controller will set the endpoint records on the ingress objects to reflect those on the service.`)

		tcpConfigMapName = flags.String("tcp-services-configmap", "",
			`Name of the ConfigMap that contains the definition of the TCP services to expose.
		The key in the map indicates the external port to be used. The value is the name of the
		service with the format namespace/serviceName and the port of the service could be a
		number of the name of the port.
		The ports 80 and 443 are not allowed as external ports. This ports are reserved for the backend`)

		udpConfigMapName = flags.String("udp-services-configmap", "",
			`Name of the ConfigMap that contains the definition of the UDP services to expose.
		The key in the map indicates the external port to be used. The value is the name of the
		service with the format namespace/serviceName and the port of the service could be a
		number of the name of the port.`)

		resyncPeriod = flags.Duration("sync-period", 600*time.Second,
			`Relist and confirm cloud resources this often. Default is 10 minutes`)

		watchNamespace = flags.String("watch-namespace", apiv1.NamespaceAll,
			`Namespace to watch for Ingress. Default is to watch all namespaces`)

		profiling = flags.Bool("profiling", true, `Enable profiling via web interface host:port/debug/pprof/`)

		defSSLCertificate = flags.String("default-ssl-certificate", "", `Name of the secret
		that contains a SSL certificate to be used as default for a HTTPS catch-all server.
		Takes the form <namespace>/<secret name>.`)

		defHealthzURL = flags.String("health-check-path", "/healthz", `Defines
		the URL to be used as health check inside in the default server in NGINX.`)

		updateStatus = flags.Bool("update-status", true, `Indicates if the
		ingress controller should update the Ingress status IP/hostname. Default is true`)

		electionID = flags.String("election-id", "ingress-controller-leader", `Election id to use for status update.`)

		forceIsolation = flags.Bool("force-namespace-isolation", false,
			`Force namespace isolation. This flag is required to avoid the reference of secrets or
		configmaps located in a different namespace than the specified in the flag --watch-namespace.`)

		updateStatusOnShutdown = flags.Bool("update-status-on-shutdown", true, `Indicates if the
		ingress controller should update the Ingress status IP/hostname when the controller
		is being stopped. Default is true`)

		sortBackends = flags.Bool("sort-backends", false,
			`Defines if backends and its endpoints should be sorted`)

		useNodeInternalIP = flags.Bool("report-node-internal-ip-address", false,
			`Defines if the nodes IP address to be returned in the ingress status should be the internal instead of the external IP address`)

		showVersion = flags.Bool("version", false,
			`Shows release information about the NGINX Ingress controller`)

		enableSSLPassthrough = flags.Bool("enable-ssl-passthrough", false, `Enable SSL passthrough feature. Default is disabled`)

		httpPort      = flags.Int("http-port", 80, `Indicates the port to use for HTTP traffic`)
		httpsPort     = flags.Int("https-port", 443, `Indicates the port to use for HTTPS traffic`)
		statusPort    = flags.Int("status-port", 18080, `Indicates the TCP port to use for exposing the nginx status page`)
		sslProxyPort  = flags.Int("ssl-passtrough-proxy-port", 442, `Default port to use internally for SSL when SSL Passthgough is enabled`)
		defServerPort = flags.Int("default-server-port", 8181, `Default port to use for exposing the default server (catch all)`)
		healthzPort   = flags.Int("healthz-port", 10254, "port for healthz endpoint.")

		annotationsPrefix = flags.String("annotations-prefix", "nginx.ingress.kubernetes.io", `Prefix of the ingress annotations.`)

		enableSSLChainCompletion = flags.Bool("enable-ssl-chain-completion", true,
			`Defines if the nginx ingress controller should check the secrets for missing intermediate CA certificates.
		If the certificate contain issues chain issues is not possible to enable OCSP.
		Default is true.`)

		syncRateLimit = flags.Float32("sync-rate-limit", 0.3,
			`Define the sync frequency upper limit`)

		publishStatusAddress = flags.String("publish-status-address", "",
			`User customized address to be set in the status of ingress resources. The controller will set the
		endpoint records on the ingress using this address.`)

		dynamicConfigurationEnabled = flags.Bool("enable-dynamic-configuration", false,
			`When enabled controller will try to avoid Nginx reloads as much as possible by using Lua. Disabled by default.`)
	)

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	pflag.VisitAll(func(flag *pflag.Flag) {
		glog.V(2).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})

	if *showVersion {
		return true, nil, nil
	}

	if *defaultSvc == "" {
		return false, nil, fmt.Errorf("Please specify --default-backend-service")
	}

	if *ingressClass != "" {
		glog.Infof("Watching for ingress class: %s", *ingressClass)

		if *ingressClass != class.DefaultClass {
			glog.Warningf("only Ingress with class \"%v\" will be processed by this ingress controller", *ingressClass)
		}

		class.IngressClass = *ingressClass
	}

	parser.AnnotationsPrefix = *annotationsPrefix

	// check port collisions
	if !ing_net.IsPortAvailable(*httpPort) {
		return false, nil, fmt.Errorf("Port %v is already in use. Please check the flag --http-port", *httpPort)
	}

	if !ing_net.IsPortAvailable(*httpsPort) {
		return false, nil, fmt.Errorf("Port %v is already in use. Please check the flag --https-port", *httpsPort)
	}

	if !ing_net.IsPortAvailable(*statusPort) {
		return false, nil, fmt.Errorf("Port %v is already in use. Please check the flag --status-port", *statusPort)
	}

	if !ing_net.IsPortAvailable(*defServerPort) {
		return false, nil, fmt.Errorf("Port %v is already in use. Please check the flag --default-server-port", *defServerPort)
	}

	if *enableSSLPassthrough && !ing_net.IsPortAvailable(*sslProxyPort) {
		return false, nil, fmt.Errorf("Port %v is already in use. Please check the flag --ssl-passtrough-proxy-port", *sslProxyPort)
	}

	if !*enableSSLChainCompletion {
		glog.Warningf("Check of SSL certificate chain is disabled (--enable-ssl-chain-completion=false)")
	}

	// LuaJIT is not available on arch s390x and ppc64le
	disableLua := false
	if runtime.GOARCH == "s390x" || runtime.GOARCH == "ppc64le" {
		disableLua = true
		if *dynamicConfigurationEnabled {
			*dynamicConfigurationEnabled = false
			glog.Warningf("Disabling dynamic configuration feature (LuaJIT is not available in s390x and ppc64le)")
		}
	}

	config := &controller.Configuration{
		APIServerHost:               *apiserverHost,
		KubeConfigFile:              *kubeConfigFile,
		UpdateStatus:                *updateStatus,
		ElectionID:                  *electionID,
		EnableProfiling:             *profiling,
		EnableSSLPassthrough:        *enableSSLPassthrough,
		EnableSSLChainCompletion:    *enableSSLChainCompletion,
		ResyncPeriod:                *resyncPeriod,
		DefaultService:              *defaultSvc,
		Namespace:                   *watchNamespace,
		ConfigMapName:               *configMap,
		TCPConfigMapName:            *tcpConfigMapName,
		UDPConfigMapName:            *udpConfigMapName,
		DefaultSSLCertificate:       *defSSLCertificate,
		DefaultHealthzURL:           *defHealthzURL,
		PublishService:              *publishSvc,
		PublishStatusAddress:        *publishStatusAddress,
		ForceNamespaceIsolation:     *forceIsolation,
		UpdateStatusOnShutdown:      *updateStatusOnShutdown,
		SortBackends:                *sortBackends,
		UseNodeInternalIP:           *useNodeInternalIP,
		SyncRateLimit:               *syncRateLimit,
		DynamicConfigurationEnabled: *dynamicConfigurationEnabled,
		DisableLua:                  disableLua,
		ListenPorts: &ngx_config.ListenPorts{
			Default:  *defServerPort,
			Health:   *healthzPort,
			HTTP:     *httpPort,
			HTTPS:    *httpsPort,
			SSLProxy: *sslProxyPort,
			Status:   *statusPort,
		},
	}

	return false, config, nil
}
