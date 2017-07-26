package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/spf13/pflag"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	nginxconfig "k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

type DiscoveryReciever interface {
	GetResponsesToSwap() *unsafe.Pointer
	Start()
}

type EnvoyController struct {
	startedReciever bool
	reciever        DiscoveryReciever
}

func NewEnvoyController(reciever DiscoveryReciever) ingress.Controller {
	return &EnvoyController{
		reciever:        reciever,
		startedReciever: false,
	}
}

var ingressClass = "envoy"

func (ec *EnvoyController) Name() string {
	return "envoy-controller"
}

// TODO(cmaloney): Check fs for "bad health check" flag
// Also send a request to envoy health endpoint
func (ec *EnvoyController) Check(req *http.Request) error {
	return nil
}

func (ec *EnvoyController) OnUpdate(configuration ingress.Configuration) error {
	log.Printf("Received OnUpdate notification")
	newResponses := &Responses{
		Sds: make(map[string]SdsResponse),
	}

	for _, backend := range configuration.Backends {
		sds := SdsResponse{}
		// TODO(cmaloney): sessionAffinity, Secure, SSLPassthrough
		// TODO(cmaloney): How to get the HealthCheckPath?
		sds.Hosts = make([]EnvoyHost, 0, len(backend.Endpoints))
		for _, endpoint := range backend.Endpoints {
			// TODO(cmaloney): MaxFails,  FailTimeout
			port, err := strconv.Atoi(endpoint.Port)
			if err != nil {
				log.Printf("Recieved non-integer port %s for endpoint %+v: %s", endpoint.Port, endpoint, err)
				continue
			}
			sds.Hosts = append(sds.Hosts, EnvoyHost{
				IpAddress: endpoint.Address,
				Port:      port,
				// TODO(cmaloney): Tags, such as "canary"
			})
		}
		newResponses.Cds.Clusters = append(newResponses.Cds.Clusters,
			EnvoyCluster{
				Name: backend.Name,
				Type: "sds",
				// TODO(cmaloney): Make tunable
				ConnectTimeoutMs: 10000,
				// TODO(cmaloney): Make tunable
				LbType: "least_request",

				ServiceName: backend.Name,

				// TODO(cmaloney): Make tunable
				HealthCheck: &EnvoyHealthCheck{
					Type:               "http",
					TimeoutMs:          1000,
					IntervalMs:         5000,
					UnhealthyThreshold: 5,
					HealthyThreshold:   5,
					Path:               "/health-check",
				},
			})
		newResponses.Sds[backend.Name] = sds
	}

	// TODO(cmaloney): Turn on ValidateClusters to ensure we wrote consistent config
	// newResponses.Rds.ValidateClusters = true
	newResponses.Rds.VirtualHosts = make([]EnvoyVirtualHost, 0, len(configuration.Servers))
	for _, server := range configuration.Servers {
		vh := EnvoyVirtualHost{
			Name:    server.Hostname,
			Domains: []string{server.Hostname},
		}
		vh.Routes = make([]EnvoyRoute, 0, len(server.Locations))
		for _, location := range server.Locations {
			vh.Routes = append(vh.Routes, EnvoyRoute{
				Prefix:  location.Path,
				Cluster: location.Backend,
				// TODO(cmaloney): Make configurable
				RetryPolicy: EnvoyRetryPolicy{
					RetryOn:    "5xx",
					NumRetries: 3,
					// TODO(cmaloney): Lower client request times / speed this up.
					// 20min max time per request
					PerTryTimeoutMs: int(20 * time.Minute / time.Millisecond),
				},
				// TODO(cmaloney): Lower client request times / speed this up.
				// 20min max time per request
				TimeoutMs: int(20 * time.Minute / time.Millisecond),
				// TODO(cmaloney): BasicDigestAuth, Denied, EnableCors, ExternalAuth,
				// RateLimit, Redirect, Whitelist, Proxy, CertificateAuth, UsePortInRedirects, ConfigurationSnippet
			})
		}
		newResponses.Rds.VirtualHosts = append(newResponses.Rds.VirtualHosts, vh)
	}

	// TODO(cmaloney): Auto-configure based on the actual ingress info
	newResponses.Lds.Listeners = []EnvoyListener{
		EnvoyListener{
			Name:    "default-http",
			Address: "tcp://0.0.0.0:80",
			Filters: []EnvoyFilter{
				EnvoyFilter{
					Type: "read",
					Name: "http_connection_manager",
					Config: HttpConnectionManagerConfig{
						CodecType:  "auto",
						StatPrefix: "default-http",
						Rds: RdsConfig{
							// TODO(cmaloney): Allow multiple RouteConfigName and dispatch properly
							Cluster:         "local_lds",
							RouteConfigName: "default",
							RefreshDelayMs:  1000,
						},
						Filters: []EnvoyFilter{
							EnvoyFilter{
								Type:   "decoder",
								Name:   "router",
								Config: map[string]string{},
							},
						},
					},
				},
			},
			UseProxyProto: false,
		},
	}

	// Add an additional localhost cluster for envoy to use to reference this
	// local_lds, local_cds, local_sds

	// TODO(cmaloney): TCPEndpoints, UDPEndpoints

	// Produce a DiscoveryItems and send it to all "listeners"
	log.Println("Swapping discovery info for reciever")
	if ec.reciever != nil {
		oldResponses := ec.reciever.GetResponsesToSwap()
		atomic.StorePointer(oldResponses, unsafe.Pointer(newResponses))
	}
	if !ec.startedReciever {
		if ec.reciever != nil {
			ec.reciever.Start()
		}
		ec.startedReciever = true
	}
	return nil
}

func (ec *EnvoyController) SetConfig(cfg *api.ConfigMap) {
	log.Printf("Config map %+v", cfg)
}

func (ec *EnvoyController) SetListers(ingress.StoreLister) {}

func (ec *EnvoyController) BackendDefaults() defaults.Backend {
	// Just adopt nginx's default backend config
	return nginxconfig.NewDefault().Backend
}

func (ec *EnvoyController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "envoy-controller",
		Release:    "0.0.0",
		Build:      "git-00000000",
		Repository: "https://github.com/cmaloney/ingress/",
	}
}

func (ec *EnvoyController) ConfigureFlags(*pflag.FlagSet) {
}

func (ec *EnvoyController) OverrideFlags(*pflag.FlagSet) {
}

func (ec *EnvoyController) DefaultIngressClass() string {
	return ingressClass
}

// TODO(cmaloney): Implement this
func (ec *EnvoyController) UpdateIngressStatus(*extensions.Ingress) []api.LoadBalancerIngress {
	return nil
}
