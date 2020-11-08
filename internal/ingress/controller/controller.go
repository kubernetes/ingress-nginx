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

// getPublishService returns the Service used to set the load-balancer status of Ingresses.
func (n NGINXController) getPublishService() *apiv1.Service {
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
	// 1. rate limite sync
	n.syncRateLimiter.Accept()
	// 2. skip sync during shutdown
	if n.syncQueue.IsShuttingDown() {
		return nil
	}

	// 3. obtain local copy of ingresses
	ings := n.store.ListIngresses()

	// 4. build nginx configuration
	pcfg := n.buildConfiguration(ings)

	// 5. update metrics
	n.metricCollector.SetSSLExpireTime(pcfg.Servers)

	// 6. check if the configuration changed
	if n.runningConfig.Equal(pcfg) {
		klog.V(3).Infof("No configuration change detected, skipping backend reload")
		return nil
	}

	// 7. update prometheus (hosts) metrics
	n.metricCollector.SetHosts(sets.NewString(pcfg.Hostnames...))

	// 8. can we skip a reload?
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

			klog.ErrorS(err, "unexpected failure reloading the backend")
			n.recorder.Eventf(k8s.IngressNGINXPod, apiv1.EventTypeWarning, "RELOAD", fmt.Sprintf("Error reloading NGINX: %v", err))

			return err
		}

		klog.InfoS("Backend successfully reloaded")
		n.metricCollector.ConfigSuccess(hash, true)
		n.metricCollector.IncReloadCount()

		n.recorder.Eventf(k8s.IngressNGINXPod, apiv1.EventTypeNormal, "RELOAD", "NGINX reload triggered due to a change in configuration")
	}

	// 9. to avoid multiple syncs, wait if this is the first one
	isFirstSync := n.runningConfig.Equal(&ingress.Configuration{})
	if isFirstSync {
		// For the initial sync it always takes some time for NGINX to start listening
		// For large configurations it might take a while so we loop and back off
		klog.InfoS("Initial sync, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	// 10. try to update the configuration
	err := wait.ExponentialBackoff(wait.Backoff{
		Steps:    15,
		Duration: 1 * time.Second,
		Factor:   0.8,
		Jitter:   0.1,
	}, func() (bool, error) {
		err := n.configureDynamically(pcfg)
		if err == nil {
			klog.V(2).InfoS("Dynamic reconfiguration succeeded")
			return true, nil
		}

		klog.ErrorS(err, "dynamic reconfiguration failed")
		return false, err
	})
	if err != nil {
		klog.Errorf("Unexpected failure reconfiguring NGINX:\n%v", err)
		return err
	}

	// 11. determine ingresses and endpoints removed after the update
	ri := getRemovedIngresses(n.runningConfig, pcfg)
	re := getRemovedHosts(n.runningConfig, pcfg)
	n.metricCollector.RemoveMetrics(ri, re)

	// 12. update copy of running configuration
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

	pcfg := n.buildConfiguration(ings)

	err := checkOverlap(ing, allIngresses, pcfg.Servers, n.store.GetIngress)
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

func buildPassthroughBackends(servers []*ingress.Server) []*ingress.SSLPassthroughBackend {
	passUpstreams := []*ingress.SSLPassthroughBackend{}
	for _, server := range servers {
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

	return passUpstreams
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

// buildConfiguration
func (n *NGINXController) buildConfiguration(ingresses []*ingress.Ingress) *ingress.Configuration {
	// 1. default upstream
	du := n.getDefaultUpstream()

	// 2. build list of servernames
	hostnames := extractServers(ingresses)

	// 3. build default proxy location configuration
	proxyConfig := buildProxyConfig(n.store.GetDefaultBackend())

	// 4. build server locations
	enableAccessLogForDefaultBackend := n.store.GetBackendConfiguration().EnableAccessLogForDefaultBackend
	serverLocations := buildServerLocations(enableAccessLogForDefaultBackend, proxyConfig, du, hostnames, ingresses, n.store)

	// 5. update server configuration defined in ingresses
	proxySSLLocationOnly := n.store.GetBackendConfiguration().ProxySSLLocationOnly
	serverConfiguration := buildServerConfiguration(proxySSLLocationOnly, serverLocations, ingresses, n.store)

	// 6. build upstreams
	upstreams := n.createUpstreams(ingresses, du)
	upstreamsWithDefaultBackends := buildUpstreamsWithDefaultBackend(upstreams, serverConfiguration, n.store)
	upstreamsWithDefaultBackends = updateUpstreamsWithCanaries(upstreamsWithDefaultBackends)

	klog.InfoS("Servers", "hostnames", hostnames)
	klog.InfoS("Server Locations", "locations", serverLocations)
	klog.InfoS("Server Configurations", "configurations", serverConfiguration)

	// assemble final configuration
	servers := []*ingress.Server{}
	for _, server := range serverConfiguration {
		servers = append(servers, server)
	}

	return &ingress.Configuration{
		Hostnames:             hostnames,
		Backends:              upstreamsWithDefaultBackends,
		Servers:               servers,
		TCPEndpoints:          n.getStreamServices(n.cfg.TCPConfigMapName, apiv1.ProtocolTCP),
		UDPEndpoints:          n.getStreamServices(n.cfg.UDPConfigMapName, apiv1.ProtocolUDP),
		PassthroughBackends:   buildPassthroughBackends(servers),
		BackendConfigChecksum: n.store.GetBackendConfiguration().Checksum,
	}
}

// createUpstreams creates the NGINX upstreams (Endpoints) for each Service referenced in Ingress rules.
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
				endpoint := n.serviceClusterIP(svcKey, ing.Spec.Backend)
				if endpoint != nil {
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
					endpoint := n.serviceClusterIP(svcKey, &path.Backend)
					if endpoint != nil {
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
func serviceClusterEndpoint(service *apiv1.Service, backend *networking.IngressBackend) (*ingress.Endpoint, error) {
	if len(service.Spec.ClusterIP) == 0 || service.Spec.ClusterIP == apiv1.ClusterIPNone {
		return nil, fmt.Errorf("No ClusterIP found for Service %q", k8s.MetaNamespaceKey(service))
	}

	// if the Service port is referenced by name in the Ingress, lookup the
	// actual port in the service spec
	if backend.ServicePort.Type != intstr.String {
		return &ingress.Endpoint{
			Address: service.Spec.ClusterIP,
			Port:    backend.ServicePort.String(),
		}, nil
	}

	var port int32 = -1
	for _, svcPort := range service.Spec.Ports {
		if svcPort.Name == backend.ServicePort.String() {
			port = svcPort.Port
			break
		}
	}

	if port == -1 {
		return nil, fmt.Errorf("service %q does not have a port named %q", service.Name, backend.ServicePort)
	}

	return &ingress.Endpoint{
		Address: service.Spec.ClusterIP,
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

func (n *NGINXController) serviceClusterIP(serviceKey string, backend *networking.IngressBackend) *ingress.Endpoint {
	service, err := n.store.GetService(serviceKey)
	if err != nil {
		klog.ErrorS(err, "searching service", "service", serviceKey)
		return nil
	}

	endpoint, err := serviceClusterEndpoint(service, backend)
	if err != nil {
		klog.ErrorS(err, "failed to determine a suitable ClusterIP Endpoint", "service", serviceKey)
		return nil
	}

	return endpoint
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

			if !oldIngresses.Has(location.Ingress) {
				oldIngresses.Insert(location.Ingress)
			}
		}
	}

	for _, server := range newcfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == "" {
				continue
			}

			if !newIngresses.Has(location.Ingress) {
				newIngresses.Insert(location.Ingress)
			}
		}
	}

	return oldIngresses.Difference(newIngresses).List()
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
