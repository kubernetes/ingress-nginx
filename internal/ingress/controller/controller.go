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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure"
	apiv1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxy"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/store"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/klog/v2"
)

const (
	defUpstreamName = "upstream-default-backend"
	defServerName   = "_"
	rootLocation    = "/"
)

var (
	pathTypeExact  = networking.PathTypeExact
	pathTypePrefix = networking.PathTypePrefix
)

// Configuration contains all the settings required by an Ingress controller
type Configuration struct {
	APIServerHost string
	RootCAFile    string

	KubeConfigFile string

	Client clientset.Interface

	ResyncPeriod time.Duration

	ConfigMapName  string
	DefaultService string

	Namespace string

	// +optional
	TCPConfigMapName string
	// +optional
	UDPConfigMapName string

	DefaultSSLCertificate string

	// +optional
	PublishService       string
	PublishStatusAddress string

	UpdateStatus           bool
	UseNodeInternalIP      bool
	ElectionID             string
	UpdateStatusOnShutdown bool

	ListenPorts *ngx_config.ListenPorts

	EnableSSLPassthrough bool

	EnableProfiling bool

	EnableMetrics  bool
	MetricsPerHost bool

	FakeCertificate *ingress.SSLCert

	SyncRateLimit float32

	DisableCatchAll bool

	ValidationWebhook         string
	ValidationWebhookCertPath string
	ValidationWebhookKeyPath  string

	GlobalExternalAuth  *ngx_config.GlobalExternalAuth
	MaxmindEditionFiles []string

	MonitorMaxBatchSize int
}

// GetPublishService returns the Service used to set the load-balancer status of Ingresses.
func (n NGINXController) GetPublishService() *apiv1.Service {
	s, err := n.store.GetService(n.cfg.PublishService)
	if err != nil {
		return nil
	}

	return s
}

// syncIngress collects all the pieces required to assemble the NGINX
// configuration file and passes the resulting data structures to the backend
// (OnUpdate) when a reload is deemed necessary.
func (n *NGINXController) syncIngress(interface{}) error {
	n.syncRateLimiter.Accept()

	if n.syncQueue.IsShuttingDown() {
		return nil
	}

	ings := n.store.ListIngresses()
	hosts, servers, pcfg := n.getConfiguration(ings)

	n.metricCollector.SetSSLExpireTime(servers)

	if n.runningConfig.Equal(pcfg) {
		klog.V(3).Infof("No configuration change detected, skipping backend reload")
		return nil
	}

	n.metricCollector.SetHosts(hosts)

	if !n.IsDynamicConfigurationEnough(pcfg) {
		klog.InfoS("Configuration changes detected, backend reload required")

		hash, _ := hashstructure.Hash(pcfg, &hashstructure.HashOptions{
			TagName: "json",
		})

		pcfg.ConfigurationChecksum = fmt.Sprintf("%v", hash)

		err := n.OnUpdate(*pcfg)
		if err != nil {
			n.metricCollector.IncReloadErrorCount()
			n.metricCollector.ConfigSuccess(hash, false)
			klog.Errorf("Unexpected failure reloading the backend:\n%v", err)
			n.recorder.Eventf(k8s.IngressNGINXPod, apiv1.EventTypeWarning, "RELOAD", fmt.Sprintf("Error reloading NGINX: %v", err))
			return err
		}

		klog.InfoS("Backend successfully reloaded")
		n.metricCollector.ConfigSuccess(hash, true)
		n.metricCollector.IncReloadCount()

		n.recorder.Eventf(k8s.IngressNGINXPod, apiv1.EventTypeNormal, "RELOAD", "NGINX reload triggered due to a change in configuration")
	}

	isFirstSync := n.runningConfig.Equal(&ingress.Configuration{})
	if isFirstSync {
		// For the initial sync it always takes some time for NGINX to start listening
		// For large configurations it might take a while so we loop and back off
		klog.InfoS("Initial sync, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	retry := wait.Backoff{
		Steps:    15,
		Duration: 1 * time.Second,
		Factor:   0.8,
		Jitter:   0.1,
	}

	err := wait.ExponentialBackoff(retry, func() (bool, error) {
		err := n.configureDynamically(pcfg)
		if err == nil {
			klog.V(2).Infof("Dynamic reconfiguration succeeded.")
			return true, nil
		}

		klog.Warningf("Dynamic reconfiguration failed: %v", err)
		return false, err
	})
	if err != nil {
		klog.Errorf("Unexpected failure reconfiguring NGINX:\n%v", err)
		return err
	}

	ri := getRemovedIngresses(n.runningConfig, pcfg)
	re := getRemovedHosts(n.runningConfig, pcfg)
	n.metricCollector.RemoveMetrics(ri, re)

	n.runningConfig = pcfg

	return nil
}

// CheckIngress returns an error in case the provided ingress, when added
// to the current configuration, generates an invalid configuration
func (n *NGINXController) CheckIngress(ing *networking.Ingress) error {
	if ing == nil {
		// no ingress to add, no state change
		return nil
	}

	if !class.IsValid(ing) {
		klog.Warningf("ignoring ingress %v in %v based on annotation %v", ing.Name, ing.ObjectMeta.Namespace, class.IngressKey)
		return nil
	}

	if n.cfg.Namespace != "" && ing.ObjectMeta.Namespace != n.cfg.Namespace {
		klog.Warningf("ignoring ingress %v in namespace %v different from the namespace watched %s", ing.Name, ing.ObjectMeta.Namespace, n.cfg.Namespace)
		return nil
	}

	if parser.AnnotationsPrefix != parser.DefaultAnnotationsPrefix {
		for key := range ing.ObjectMeta.GetAnnotations() {
			if strings.HasPrefix(key, fmt.Sprintf("%s/", parser.DefaultAnnotationsPrefix)) {
				return fmt.Errorf("This deployment has a custom annotation prefix defined. Use '%s' instead of '%s'", parser.AnnotationsPrefix, parser.DefaultAnnotationsPrefix)
			}
		}
	}

	k8s.SetDefaultNGINXPathType(ing)

	allIngresses := n.store.ListIngresses()

	filter := func(toCheck *ingress.Ingress) bool {
		return toCheck.ObjectMeta.Namespace == ing.ObjectMeta.Namespace &&
			toCheck.ObjectMeta.Name == ing.ObjectMeta.Name
	}
	ings := store.FilterIngresses(allIngresses, filter)
	ings = append(ings, &ingress.Ingress{
		Ingress:           *ing,
		ParsedAnnotations: annotations.NewAnnotationExtractor(n.store).Extract(ing),
	})

	cfg := n.store.GetBackendConfiguration()
	cfg.Resolver = n.resolver

	_, servers, pcfg := n.getConfiguration(ings)

	err := checkOverlap(ing, allIngresses, servers, n.store.GetIngress)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}

	content, err := n.generateTemplate(cfg, *pcfg)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}

	err = n.testTemplate(content)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}

	n.metricCollector.IncCheckCount(ing.ObjectMeta.Namespace, ing.Name)
	return nil
}

func (n *NGINXController) getStreamServices(configmapName string, proto apiv1.Protocol) []ingress.L4Service {
	if configmapName == "" {
		return []ingress.L4Service{}
	}
	klog.V(3).Infof("Obtaining information about %v stream services from ConfigMap %q", proto, configmapName)
	_, _, err := k8s.ParseNameNS(configmapName)
	if err != nil {
		klog.Warningf("Error parsing ConfigMap reference %q: %v", configmapName, err)
		return []ingress.L4Service{}
	}
	configmap, err := n.store.GetConfigMap(configmapName)
	if err != nil {
		klog.Warningf("Error getting ConfigMap %q: %v", configmapName, err)
		return []ingress.L4Service{}
	}

	var svcs []ingress.L4Service
	var svcProxyProtocol ingress.ProxyProtocol

	rp := []int{
		n.cfg.ListenPorts.HTTP,
		n.cfg.ListenPorts.HTTPS,
		n.cfg.ListenPorts.SSLProxy,
		n.cfg.ListenPorts.Health,
		n.cfg.ListenPorts.Default,
		nginx.ProfilerPort,
		nginx.StatusPort,
		nginx.StreamPort,
	}

	reserverdPorts := sets.NewInt(rp...)
	// svcRef format: <(str)namespace>/<(str)service>:<(intstr)port>[:<("PROXY")decode>:<("PROXY")encode>]
	for port, svcRef := range configmap.Data {
		externalPort, err := strconv.Atoi(port)
		if err != nil {
			klog.Warningf("%q is not a valid %v port number", port, proto)
			continue
		}
		if reserverdPorts.Has(externalPort) {
			klog.Warningf("Port %d cannot be used for %v stream services. It is reserved for the Ingress controller.", externalPort, proto)
			continue
		}
		nsSvcPort := strings.Split(svcRef, ":")
		if len(nsSvcPort) < 2 {
			klog.Warningf("Invalid Service reference %q for %v port %d", svcRef, proto, externalPort)
			continue
		}
		nsName := nsSvcPort[0]
		svcPort := nsSvcPort[1]
		svcProxyProtocol.Decode = false
		svcProxyProtocol.Encode = false
		// Proxy Protocol is only compatible with TCP Services
		if len(nsSvcPort) >= 3 && proto == apiv1.ProtocolTCP {
			if len(nsSvcPort) >= 3 && strings.ToUpper(nsSvcPort[2]) == "PROXY" {
				svcProxyProtocol.Decode = true
			}
			if len(nsSvcPort) == 4 && strings.ToUpper(nsSvcPort[3]) == "PROXY" {
				svcProxyProtocol.Encode = true
			}
		}
		svcNs, svcName, err := k8s.ParseNameNS(nsName)
		if err != nil {
			klog.Warningf("%v", err)
			continue
		}
		svc, err := n.store.GetService(nsName)
		if err != nil {
			klog.Warningf("Error getting Service %q: %v", nsName, err)
			continue
		}
		var endps []ingress.Endpoint
		targetPort, err := strconv.Atoi(svcPort)
		if err != nil {
			// not a port number, fall back to using port name
			klog.V(3).Infof("Searching Endpoints with %v port name %q for Service %q", proto, svcPort, nsName)
			for _, sp := range svc.Spec.Ports {
				if sp.Name == svcPort {
					if sp.Protocol == proto {
						endps = getEndpoints(svc, &sp, proto, n.store.GetServiceEndpoints)
						break
					}
				}
			}
		} else {
			klog.V(3).Infof("Searching Endpoints with %v port number %d for Service %q", proto, targetPort, nsName)
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(targetPort) {
					if sp.Protocol == proto {
						endps = getEndpoints(svc, &sp, proto, n.store.GetServiceEndpoints)
						break
					}
				}
			}
		}
		// stream services cannot contain empty upstreams and there is
		// no default backend equivalent
		if len(endps) == 0 {
			klog.Warningf("Service %q does not have any active Endpoint for %v port %v", nsName, proto, svcPort)
			continue
		}
		svcs = append(svcs, ingress.L4Service{
			Port: externalPort,
			Backend: ingress.L4Backend{
				Name:          svcName,
				Namespace:     svcNs,
				Port:          intstr.FromString(svcPort),
				Protocol:      proto,
				ProxyProtocol: svcProxyProtocol,
			},
			Endpoints: endps,
			Service:   svc,
		})
	}
	// Keep upstream order sorted to reduce unnecessary nginx config reloads.
	sort.SliceStable(svcs, func(i, j int) bool {
		return svcs[i].Port < svcs[j].Port
	})
	return svcs
}

// getDefaultUpstream returns the upstream associated with the default backend.
// Configures the upstream to return HTTP code 503 in case of error.
func (n *NGINXController) getDefaultUpstream() *ingress.Backend {
	upstream := &ingress.Backend{
		Name: defUpstreamName,
	}
	svcKey := n.cfg.DefaultService

	if len(svcKey) == 0 {
		upstream.Endpoints = append(upstream.Endpoints, n.DefaultEndpoint())
		return upstream
	}

	svc, err := n.store.GetService(svcKey)
	if err != nil {
		klog.Warningf("Error getting default backend %q: %v", svcKey, err)
		upstream.Endpoints = append(upstream.Endpoints, n.DefaultEndpoint())
		return upstream
	}

	endps := getEndpoints(svc, &svc.Spec.Ports[0], apiv1.ProtocolTCP, n.store.GetServiceEndpoints)
	if len(endps) == 0 {
		klog.Warningf("Service %q does not have any active Endpoint", svcKey)
		endps = []ingress.Endpoint{n.DefaultEndpoint()}
	}

	upstream.Service = svc
	upstream.Endpoints = append(upstream.Endpoints, endps...)
	return upstream
}

// getConfiguration returns the configuration matching the standard kubernetes ingress
func (n *NGINXController) getConfiguration(ingresses []*ingress.Ingress) (sets.String, []*ingress.Server, *ingress.Configuration) {
	upstreams, servers := n.backendConfiguration(ingresses)
	var passUpstreams []*ingress.SSLPassthroughBackend

	hosts := sets.NewString()

	for _, server := range servers {
		if !hosts.Has(server.Hostname) {
			hosts.Insert(server.Hostname)
		}

		for _, alias := range server.Aliases {
			if !hosts.Has(alias) {
				hosts.Insert(alias)
			}
		}

		if !server.SSLPassthrough {
			continue
		}

		for _, loc := range server.Locations {
			if loc.Path != rootLocation {
				klog.Warningf("Ignoring SSL Passthrough for location %q in server %q", loc.Path, server.Hostname)
				continue
			}
			passUpstreams = append(passUpstreams, &ingress.SSLPassthroughBackend{
				Backend:  loc.Backend,
				Hostname: server.Hostname,
				Service:  loc.Service,
				Port:     loc.Port,
			})
			break
		}
	}

	return hosts, servers, &ingress.Configuration{
		Backends:              upstreams,
		Servers:               servers,
		TCPEndpoints:          n.getStreamServices(n.cfg.TCPConfigMapName, apiv1.ProtocolTCP),
		UDPEndpoints:          n.getStreamServices(n.cfg.UDPConfigMapName, apiv1.ProtocolUDP),
		PassthroughBackends:   passUpstreams,
		BackendConfigChecksum: n.store.GetBackendConfiguration().Checksum,
	}
}

func buildProxyConfig(defaults defaults.Backend) proxy.Config {
	return proxy.Config{
		BodySize:             defaults.ProxyBodySize,
		ConnectTimeout:       defaults.ProxyConnectTimeout,
		SendTimeout:          defaults.ProxySendTimeout,
		ReadTimeout:          defaults.ProxyReadTimeout,
		BuffersNumber:        defaults.ProxyBuffersNumber,
		BufferSize:           defaults.ProxyBufferSize,
		CookieDomain:         defaults.ProxyCookieDomain,
		CookiePath:           defaults.ProxyCookiePath,
		NextUpstream:         defaults.ProxyNextUpstream,
		NextUpstreamTimeout:  defaults.ProxyNextUpstreamTimeout,
		NextUpstreamTries:    defaults.ProxyNextUpstreamTries,
		RequestBuffering:     defaults.ProxyRequestBuffering,
		ProxyRedirectFrom:    defaults.ProxyRedirectFrom,
		ProxyBuffering:       defaults.ProxyBuffering,
		ProxyHTTPVersion:     defaults.ProxyHTTPVersion,
		ProxyMaxTempFileSize: defaults.ProxyMaxTempFileSize,
	}
}

// getBackendServers returns a list of Upstream and Server to be used by the
// backend.  An upstream can be used in multiple servers if the namespace,
// service name and port are the same.
func (n *NGINXController) backendConfiguration(ingresses []*ingress.Ingress) ([]*ingress.Backend, []*ingress.Server) {
	du := n.getDefaultUpstream()
	upstreams := n.createUpstreams(ingresses, du)

	hostnames := extractServers(ingresses)

	proxyConfig := buildProxyConfig(n.store.GetDefaultBackend())

	enableAccessLogForDefaultBackend := n.store.GetBackendConfiguration().EnableAccessLogForDefaultBackend
	serverLocations := buildServerLocations(enableAccessLogForDefaultBackend, proxyConfig, du, hostnames, ingresses, n.store.GetIngressAnnotations)

	proxySSLLocationOnly := n.store.GetBackendConfiguration().ProxySSLLocationOnly
	serverConfiguration := buildServerConfiguration(proxySSLLocationOnly, serverLocations, ingresses, n.store.GetIngressAnnotations)

	upstreamsWithDefaultBackends := buildUpstreamsWithDefaultBackend(upstreams, serverConfiguration, n.store.GetServiceEndpoints)

	klog.InfoS("Servers", "hostnames", hostnames)
	klog.InfoS("Server Locations", "locations", serverLocations)
	klog.InfoS("Server Configurations", "configurations", serverConfiguration)

	// assemble final configuration
	servers := []*ingress.Server{}
	for _, server := range serverConfiguration {
		servers = append(servers, server)
	}

	return upstreamsWithDefaultBackends, servers
}

// createUpstreams creates the NGINX upstreams (Endpoints) for each Service
// referenced in Ingress rules.
func (n *NGINXController) createUpstreams(data []*ingress.Ingress, du *ingress.Backend) map[string]*ingress.Backend {
	upstreams := make(map[string]*ingress.Backend)
	upstreams[defUpstreamName] = du

	for _, ing := range data {
		anns := ing.ParsedAnnotations

		var defBackend string
		if ing.Spec.Backend != nil {
			defBackend = upstreamName(ing.Namespace, ing.Spec.Backend.ServiceName, ing.Spec.Backend.ServicePort)

			klog.V(3).Infof("Creating upstream %q", defBackend)
			upstreams[defBackend] = newUpstream(defBackend)

			upstreams[defBackend].UpstreamHashBy.UpstreamHashBy = anns.UpstreamHashBy.UpstreamHashBy
			upstreams[defBackend].UpstreamHashBy.UpstreamHashBySubset = anns.UpstreamHashBy.UpstreamHashBySubset
			upstreams[defBackend].UpstreamHashBy.UpstreamHashBySubsetSize = anns.UpstreamHashBy.UpstreamHashBySubsetSize

			upstreams[defBackend].LoadBalancing = anns.LoadBalancing
			if upstreams[defBackend].LoadBalancing == "" {
				upstreams[defBackend].LoadBalancing = n.store.GetBackendConfiguration().LoadBalancing
			}

			svcKey := fmt.Sprintf("%v/%v", ing.Namespace, ing.Spec.Backend.ServiceName)

			// add the service ClusterIP as a single Endpoint instead of individual Endpoints
			if anns.ServiceUpstream {
				endpoint, err := n.serviceClusterEndpoint(svcKey, ing.Spec.Backend)
				if err != nil {
					klog.ErrorS(err, "failed to determine a suitable ClusterIP Endpoint", "service", svcKey)
				} else {
					upstreams[defBackend].Endpoints = []ingress.Endpoint{*endpoint}
				}
			}

			// configure traffic shaping for canary
			if anns.Canary.Enabled {
				upstreams[defBackend].NoServer = true
				upstreams[defBackend].TrafficShapingPolicy = ingress.TrafficShapingPolicy{
					Weight:        anns.Canary.Weight,
					Header:        anns.Canary.Header,
					HeaderValue:   anns.Canary.HeaderValue,
					HeaderPattern: anns.Canary.HeaderPattern,
					Cookie:        anns.Canary.Cookie,
				}
			}

			if len(upstreams[defBackend].Endpoints) == 0 {
				endps, err := n.serviceEndpoints(svcKey, ing.Spec.Backend.ServicePort.String())
				upstreams[defBackend].Endpoints = append(upstreams[defBackend].Endpoints, endps...)
				if err != nil {
					klog.ErrorS(err, "creating upstream", "name", defBackend)
				}
			}

			s, err := n.store.GetService(svcKey)
			if err != nil {
				klog.Warningf("Error obtaining Service %q: %v", svcKey, err)
			}
			upstreams[defBackend].Service = s
		}

		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				name := upstreamName(ing.Namespace, path.Backend.ServiceName, path.Backend.ServicePort)

				if _, ok := upstreams[name]; ok {
					continue
				}

				klog.V(3).Infof("Creating upstream %q", name)
				upstreams[name] = newUpstream(name)
				upstreams[name].Port = path.Backend.ServicePort

				upstreams[name].UpstreamHashBy.UpstreamHashBy = anns.UpstreamHashBy.UpstreamHashBy
				upstreams[name].UpstreamHashBy.UpstreamHashBySubset = anns.UpstreamHashBy.UpstreamHashBySubset
				upstreams[name].UpstreamHashBy.UpstreamHashBySubsetSize = anns.UpstreamHashBy.UpstreamHashBySubsetSize

				upstreams[name].LoadBalancing = anns.LoadBalancing
				if upstreams[name].LoadBalancing == "" {
					upstreams[name].LoadBalancing = n.store.GetBackendConfiguration().LoadBalancing
				}

				svcKey := fmt.Sprintf("%v/%v", ing.Namespace, path.Backend.ServiceName)

				// add the service ClusterIP as a single Endpoint instead of individual Endpoints
				if anns.ServiceUpstream {
					endpoint, err := n.serviceClusterEndpoint(svcKey, &path.Backend)
					if err != nil {
						klog.Errorf("Failed to determine a suitable ClusterIP Endpoint for Service %q: %v", svcKey, err)
					} else {
						upstreams[name].Endpoints = []ingress.Endpoint{*endpoint}
					}
				}

				// configure traffic shaping for canary
				if anns.Canary.Enabled {
					upstreams[name].NoServer = true
					upstreams[name].TrafficShapingPolicy = ingress.TrafficShapingPolicy{
						Weight:        anns.Canary.Weight,
						Header:        anns.Canary.Header,
						HeaderValue:   anns.Canary.HeaderValue,
						HeaderPattern: anns.Canary.HeaderPattern,
						Cookie:        anns.Canary.Cookie,
					}
				}

				if len(upstreams[name].Endpoints) == 0 {
					endp, err := n.serviceEndpoints(svcKey, path.Backend.ServicePort.String())
					if err != nil {
						klog.Warningf("Error obtaining Endpoints for Service %q: %v", svcKey, err)
						continue
					}
					upstreams[name].Endpoints = endp
				}

				s, err := n.store.GetService(svcKey)
				if err != nil {
					klog.Warningf("Error obtaining Service %q: %v", svcKey, err)
					continue
				}

				upstreams[name].Service = s
			}
		}
	}

	return upstreams
}

// getServiceClusterEndpoint returns an Endpoint corresponding to the ClusterIP
// field of a Service.
func (n *NGINXController) serviceClusterEndpoint(svcKey string, backend *networking.IngressBackend) (*ingress.Endpoint, error) {
	svc, err := n.store.GetService(svcKey)
	if err != nil {
		return nil, fmt.Errorf("service %q does not exist", svcKey)
	}

	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		return nil, fmt.Errorf("no ClusterIP found for Service %q", svcKey)
	}

	// if the Service port is referenced by name in the Ingress, lookup the
	// actual port in the service spec
	if backend.ServicePort.Type != intstr.String {
		return &ingress.Endpoint{
			Address: svc.Spec.ClusterIP,
			Port:    backend.ServicePort.String(),
		}, nil
	}

	var port int32 = -1
	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Name == backend.ServicePort.String() {
			port = svcPort.Port
			break
		}
	}

	if port == -1 {
		return nil, fmt.Errorf("service %q does not have a port named %q", svc.Name, backend.ServicePort)

	}

	return &ingress.Endpoint{
		Address: svc.Spec.ClusterIP,
		Port:    fmt.Sprintf("%d", port),
	}, nil
}

// serviceEndpoints returns the upstream servers (Endpoints) associated with a Service.
func (n *NGINXController) serviceEndpoints(svcKey, backendPort string) ([]ingress.Endpoint, error) {
	var upstreams []ingress.Endpoint

	svc, err := n.store.GetService(svcKey)
	if err != nil {
		return upstreams, err
	}

	klog.V(3).Infof("Obtaining ports information for Service %q", svcKey)

	// Ingress with an ExternalName Service and no port defined for that Service
	if svc.Spec.Type == apiv1.ServiceTypeExternalName {
		servicePort := externalNamePorts(backendPort, svc)
		endps := getEndpoints(svc, servicePort, apiv1.ProtocolTCP, n.store.GetServiceEndpoints)
		if len(endps) == 0 {
			klog.Warningf("Service %q does not have any active Endpoint.", svcKey)
			return upstreams, nil
		}

		upstreams = append(upstreams, endps...)
		return upstreams, nil
	}

	for _, servicePort := range svc.Spec.Ports {
		// targetPort could be a string, use either the port name or number (int)
		if strconv.Itoa(int(servicePort.Port)) == backendPort ||
			servicePort.TargetPort.String() == backendPort ||
			servicePort.Name == backendPort {

			endps := getEndpoints(svc, &servicePort, apiv1.ProtocolTCP, n.store.GetServiceEndpoints)
			if len(endps) == 0 {
				klog.Warningf("Service %q does not have any active Endpoint.", svcKey)
			}

			upstreams = append(upstreams, endps...)
			break
		}
	}

	return upstreams, nil
}

func (n *NGINXController) getDefaultSSLCertificate() *ingress.SSLCert {
	// read custom default SSL certificate, fall back to generated default certificate
	if n.cfg.DefaultSSLCertificate != "" {
		certificate, err := n.store.GetLocalSSLCert(n.cfg.DefaultSSLCertificate)
		if err == nil {
			return certificate
		}

		klog.Warningf("Error loading custom default certificate, falling back to generated default:\n%v", err)
	}

	return n.cfg.FakeCertificate
}

func locationApplyAnnotations(loc *ingress.Location, anns *annotations.Ingress) {
	loc.BasicDigestAuth = anns.BasicDigestAuth
	loc.ClientBodyBufferSize = anns.ClientBodyBufferSize
	loc.ConfigurationSnippet = anns.ConfigurationSnippet
	loc.CorsConfig = anns.CorsConfig
	loc.ExternalAuth = anns.ExternalAuth
	loc.EnableGlobalAuth = anns.EnableGlobalAuth
	loc.HTTP2PushPreload = anns.HTTP2PushPreload
	loc.Opentracing = anns.Opentracing
	loc.Proxy = anns.Proxy
	loc.ProxySSL = anns.ProxySSL
	loc.RateLimit = anns.RateLimit
	loc.Redirect = anns.Redirect
	loc.Rewrite = anns.Rewrite
	loc.UpstreamVhost = anns.UpstreamVhost
	loc.Whitelist = anns.Whitelist
	loc.Denied = anns.Denied
	loc.XForwardedPrefix = anns.XForwardedPrefix
	loc.UsePortInRedirects = anns.UsePortInRedirects
	loc.Connection = anns.Connection
	loc.Logs = anns.Logs
	loc.InfluxDB = anns.InfluxDB
	loc.DefaultBackend = anns.DefaultBackend
	loc.BackendProtocol = anns.BackendProtocol
	loc.FastCGI = anns.FastCGI
	loc.CustomHTTPErrors = anns.CustomHTTPErrors
	loc.ModSecurity = anns.ModSecurity
	loc.Satisfy = anns.Satisfy
	loc.Mirror = anns.Mirror

	loc.DefaultBackendUpstreamName = defUpstreamName
}

// OK to merge canary ingresses iff there exists one or more ingresses to potentially merge into
func nonCanaryIngressExists(ingresses []*ingress.Ingress, canaryIngresses []*ingress.Ingress) bool {
	return len(ingresses)-len(canaryIngresses) > 0
}

// ensure that the following conditions are met
// 1) names of backends do not match and canary doesn't merge into itself
// 2) primary name is not the default upstream
// 3) the primary has a server
func canMergeBackend(primary *ingress.Backend, alternative *ingress.Backend) bool {
	return alternative != nil && primary.Name != alternative.Name && primary.Name != defUpstreamName && !primary.NoServer
}

// Performs the merge action and checks to ensure that one two alternative backends do not merge into each other
func mergeAlternativeBackend(priUps *ingress.Backend, altUps *ingress.Backend) bool {
	if priUps.NoServer {
		klog.Warningf("unable to merge alternative backend %v into primary backend %v because %v is a primary backend",
			altUps.Name, priUps.Name, priUps.Name)
		return false
	}

	for _, ab := range priUps.AlternativeBackends {
		if ab == altUps.Name {
			klog.V(2).Infof("skip merge alternative backend %v into %v, it's already present", altUps.Name, priUps.Name)
			return true
		}
	}

	priUps.AlternativeBackends =
		append(priUps.AlternativeBackends, altUps.Name)

	return true
}

// Compares an Ingress of a potential alternative backend's rules with each existing server and finds matching host + path pairs.
// If a match is found, we know that this server should back the alternative backend and add the alternative backend
// to a backend's alternative list.
// If no match is found, then the serverless backend is deleted.
func mergeAlternativeBackends(ing *ingress.Ingress, upstreams map[string]*ingress.Backend,
	servers map[string]*ingress.Server) {

	// merge catch-all alternative backends
	if ing.Spec.Backend != nil {
		upsName := upstreamName(ing.Namespace, ing.Spec.Backend.ServiceName, ing.Spec.Backend.ServicePort)

		altUps := upstreams[upsName]

		if altUps == nil {
			klog.Warningf("alternative backend %s has already been removed", upsName)
		} else {

			merged := false
			altEqualsPri := false

			for _, loc := range servers[defServerName].Locations {
				priUps := upstreams[loc.Backend]
				altEqualsPri = altUps.Name == priUps.Name
				if altEqualsPri {
					klog.Warningf("alternative upstream %s in Ingress %s/%s is primary upstream in Other Ingress for location %s%s!",
						altUps.Name, ing.Namespace, ing.Name, servers[defServerName].Hostname, loc.Path)
					break
				}

				if canMergeBackend(priUps, altUps) {
					klog.V(2).Infof("matching backend %v found for alternative backend %v",
						priUps.Name, altUps.Name)

					merged = mergeAlternativeBackend(priUps, altUps)
				}
			}

			if !altEqualsPri && !merged {
				klog.Warningf("unable to find real backend for alternative backend %v. Deleting.", altUps.Name)
				delete(upstreams, altUps.Name)
			}
		}
	}

	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			upsName := upstreamName(ing.Namespace, path.Backend.ServiceName, path.Backend.ServicePort)

			altUps := upstreams[upsName]

			if altUps == nil {
				klog.Warningf("alternative backend %s has already been removed", upsName)
				continue
			}

			merged := false
			altEqualsPri := false

			server, ok := servers[rule.Host]
			if !ok {
				klog.Errorf("cannot merge alternative backend %s into hostname %s that does not exist",
					altUps.Name,
					rule.Host)

				continue
			}

			// find matching paths
			for _, loc := range server.Locations {
				priUps := upstreams[loc.Backend]
				altEqualsPri = altUps.Name == priUps.Name
				if altEqualsPri {
					klog.Warningf("alternative upstream %s in Ingress %s/%s is primary upstream in Other Ingress for location %s%s!",
						altUps.Name, ing.Namespace, ing.Name, server.Hostname, loc.Path)
					break
				}

				if canMergeBackend(priUps, altUps) && loc.Path == path.Path && *loc.PathType == *path.PathType {
					klog.V(2).Infof("matching backend %v found for alternative backend %v",
						priUps.Name, altUps.Name)

					merged = mergeAlternativeBackend(priUps, altUps)
				}
			}

			if !altEqualsPri && !merged {
				klog.Warningf("unable to find real backend for alternative backend %v. Deleting.", altUps.Name)
				delete(upstreams, altUps.Name)
			}
		}
	}
}

// extractTLSSecretName returns the name of the Secret containing a SSL
// certificate for the given host name, or an empty string.
func extractTLSSecretName(host string, ing *ingress.Ingress,
	getLocalSSLCert func(string) (*ingress.SSLCert, error)) string {

	if ing == nil {
		return ""
	}

	// naively return Secret name from TLS spec if host name matches
	lowercaseHost := toLowerCaseASCII(host)
	for _, tls := range ing.Spec.TLS {
		for _, tlsHost := range tls.Hosts {
			if toLowerCaseASCII(tlsHost) == lowercaseHost {
				return tls.SecretName
			}
		}
	}

	// no TLS host matching host name, try each TLS host for matching SAN or CN
	for _, tls := range ing.Spec.TLS {

		if tls.SecretName == "" {
			// There's no secretName specified, so it will never be available
			continue
		}

		secrKey := fmt.Sprintf("%v/%v", ing.Namespace, tls.SecretName)

		cert, err := getLocalSSLCert(secrKey)
		if err != nil {
			klog.Warningf("Error getting SSL certificate %q: %v", secrKey, err)
			continue
		}

		if cert == nil { // for tests
			continue
		}

		err = cert.Certificate.VerifyHostname(host)
		if err != nil {
			continue
		}
		klog.V(3).Infof("Found SSL certificate matching host %q: %q", host, secrKey)
		return tls.SecretName
	}

	return ""
}

// getRemovedHosts returns a list of the hostsnames
// that are not associated anymore to the NGINX configuration.
func getRemovedHosts(rucfg, newcfg *ingress.Configuration) []string {
	old := sets.NewString()
	new := sets.NewString()

	for _, s := range rucfg.Servers {
		if !old.Has(s.Hostname) {
			old.Insert(s.Hostname)
		}
	}

	for _, s := range newcfg.Servers {
		if !new.Has(s.Hostname) {
			new.Insert(s.Hostname)
		}
	}

	return old.Difference(new).List()
}

func getRemovedIngresses(rucfg, newcfg *ingress.Configuration) []string {
	oldIngresses := sets.NewString()
	newIngresses := sets.NewString()

	for _, server := range rucfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == "" {
				continue
			}

			ingKey := k8s.MetaNamespaceKey(location.Ingress)
			if !oldIngresses.Has(ingKey) {
				oldIngresses.Insert(ingKey)
			}
		}
	}

	for _, server := range newcfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == "" {
				continue
			}

			ingKey := k8s.MetaNamespaceKey(location.Ingress)
			if !newIngresses.Has(ingKey) {
				newIngresses.Insert(ingKey)
			}
		}
	}

	return oldIngresses.Difference(newIngresses).List()
}

// checks conditions for whether or not an upstream should be created for a custom default backend
func shouldCreateUpstreamForLocationDefaultBackend(upstream *ingress.Backend, location *ingress.Location) bool {
	return (upstream.Name == location.Backend) &&
		(len(upstream.Endpoints) == 0 || len(location.CustomHTTPErrors) != 0) &&
		location.DefaultBackend != nil
}

func externalNamePorts(name string, svc *apiv1.Service) *apiv1.ServicePort {
	port, err := strconv.Atoi(name)
	if err != nil {
		// not a number. check port names.
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Name != name {
				continue
			}

			tp := svcPort.TargetPort
			if tp.IntValue() == 0 {
				tp = intstr.FromInt(int(svcPort.Port))
			}

			return &apiv1.ServicePort{
				Protocol:   "TCP",
				Port:       svcPort.Port,
				TargetPort: tp,
			}
		}
	}

	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Port != int32(port) {
			continue
		}

		tp := svcPort.TargetPort
		if tp.IntValue() == 0 {
			tp = intstr.FromInt(port)
		}

		return &apiv1.ServicePort{
			Protocol:   "TCP",
			Port:       svcPort.Port,
			TargetPort: svcPort.TargetPort,
		}
	}

	// ExternalName without port
	return &apiv1.ServicePort{
		Protocol:   "TCP",
		Port:       int32(port),
		TargetPort: intstr.FromInt(port),
	}
}

func checkOverlap(ing *networking.Ingress, ingresses []*ingress.Ingress, servers []*ingress.Server, ingressByKey func(key string) (*networking.Ingress, error)) error {
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		if rule.Host == "" {
			rule.Host = defServerName
		}

		for _, path := range rule.HTTP.Paths {
			if path.Path == "" {
				path.Path = rootLocation
			}

			existingIngresses := ingressForHostPath(rule.Host, path.Path, servers, ingressByKey)

			// no previous ingress
			if len(existingIngresses) == 0 {
				continue
			}

			// same ingress
			skipValidation := false
			for _, existing := range existingIngresses {
				if existing.ObjectMeta.Namespace == ing.ObjectMeta.Namespace && existing.ObjectMeta.Name == ing.ObjectMeta.Name {
					return nil
				}
			}

			if skipValidation {
				continue
			}

			// path overlap. Check if one of the ingresses has a canary annotation
			isCanaryEnabled, annotationErr := parser.GetBoolAnnotation("canary", ing)
			for _, existing := range existingIngresses {
				isExistingCanaryEnabled, existingAnnotationErr := parser.GetBoolAnnotation("canary", existing)

				if isCanaryEnabled && isExistingCanaryEnabled {
					return fmt.Errorf(`host "%s" and path "%s" is already defined in ingress %s/%s`, rule.Host, path.Path, existing.Namespace, existing.Name)
				}

				if annotationErr == errors.ErrMissingAnnotations && existingAnnotationErr == existingAnnotationErr {
					return fmt.Errorf(`host "%s" and path "%s" is already defined in ingress %s/%s`, rule.Host, path.Path, existing.Namespace, existing.Name)
				}
			}

			// no overlap
			return nil
		}
	}

	return nil
}

func ingressForHostPath(hostname, path string, servers []*ingress.Server, ingressByKey func(key string) (*networking.Ingress, error)) []*networking.Ingress {
	ingresses := make([]*networking.Ingress, 0)

	for _, server := range servers {
		if hostname != server.Hostname {
			continue
		}

		for _, location := range server.Locations {
			if location.Path != path {
				continue
			}

			ing, _ := ingressByKey(location.Ingress)
			ingresses = append(ingresses, ing)
		}
	}

	return ingresses
}
