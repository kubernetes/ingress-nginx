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
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/healthcheck"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxy"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/k8s"
)

const (
	defUpstreamName = "upstream-default-backend"
	defServerName   = "_"
	rootLocation    = "/"
)

// Configuration contains all the settings required by an Ingress controller
type Configuration struct {
	APIServerHost  string
	KubeConfigFile string
	Client         clientset.Interface

	ResyncPeriod time.Duration

	ConfigMapName  string
	DefaultService string

	Namespace string

	ForceNamespaceIsolation bool

	// optional
	TCPConfigMapName string
	// optional
	UDPConfigMapName string

	DefaultHealthzURL     string
	DefaultSSLCertificate string

	// optional
	PublishService       string
	PublishStatusAddress string

	UpdateStatus           bool
	UseNodeInternalIP      bool
	ElectionID             string
	UpdateStatusOnShutdown bool

	SortBackends bool

	ListenPorts *ngx_config.ListenPorts

	EnableSSLPassthrough bool

	EnableProfiling bool

	EnableSSLChainCompletion bool

	FakeCertificatePath string
	FakeCertificateSHA  string

	SyncRateLimit float32

	DynamicConfigurationEnabled bool

	DisableLua bool
}

// GetPublishService returns the configured service used to set ingress status
func (n NGINXController) GetPublishService() *apiv1.Service {
	s, err := n.store.GetService(n.cfg.PublishService)
	if err != nil {
		return nil
	}

	return s
}

// sync collects all the pieces required to assemble the configuration file and
// then sends the content to the backend (OnUpdate) receiving the populated
// template as response reloading the backend if is required.
func (n *NGINXController) syncIngress(interface{}) error {
	n.syncRateLimiter.Accept()

	if n.syncQueue.IsShuttingDown() {
		return nil
	}

	// Sort ingress rules using the ResourceVersion field
	ings := n.store.ListIngresses()
	sort.SliceStable(ings, func(i, j int) bool {
		ir := ings[i].ResourceVersion
		jr := ings[j].ResourceVersion
		return ir < jr
	})

	upstreams, servers := n.getBackendServers(ings)
	var passUpstreams []*ingress.SSLPassthroughBackend

	for _, server := range servers {
		if !server.SSLPassthrough {
			continue
		}

		for _, loc := range server.Locations {
			if loc.Path != rootLocation {
				glog.Warningf("ignoring path %v of ssl passthrough host %v", loc.Path, server.Hostname)
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

	pcfg := ingress.Configuration{
		Backends:            upstreams,
		Servers:             servers,
		TCPEndpoints:        n.getStreamServices(n.cfg.TCPConfigMapName, apiv1.ProtocolTCP),
		UDPEndpoints:        n.getStreamServices(n.cfg.UDPConfigMapName, apiv1.ProtocolUDP),
		PassthroughBackends: passUpstreams,
	}

	if !n.isForceReload() && n.runningConfig.Equal(&pcfg) {
		glog.V(3).Infof("skipping backend reload (no changes detected)")
		return nil
	}

	if n.cfg.DynamicConfigurationEnabled && n.IsDynamicConfigurationEnough(&pcfg) && !n.isForceReload() {
		glog.Infof("skipping reload")
	} else {
		glog.Infof("backend reload required")

		err := n.OnUpdate(pcfg)
		if err != nil {
			incReloadErrorCount()
			glog.Errorf("unexpected failure restarting the backend: \n%v", err)
			return err
		}

		glog.Infof("ingress backend successfully reloaded...")
		incReloadCount()
		setSSLExpireTime(servers)
	}

	if n.cfg.DynamicConfigurationEnabled {
		isFirstSync := n.runningConfig.Equal(&ingress.Configuration{})
		go func(isFirstSync bool) {
			if isFirstSync {
				glog.Infof("first sync of Nginx configuration")

				// it takes time for Nginx to start listening on the port
				time.Sleep(1 * time.Second)
			}
			err := n.ConfigureDynamically(&pcfg)
			if err == nil {
				glog.Infof("dynamic reconfiguration succeeded")
			} else {
				glog.Warningf("could not dynamically reconfigure: %v", err)
			}
		}(isFirstSync)
	}

	n.runningConfig = &pcfg
	n.SetForceReload(false)

	return nil
}

func (n *NGINXController) getStreamServices(configmapName string, proto apiv1.Protocol) []ingress.L4Service {
	glog.V(3).Infof("obtaining information about stream services of type %v located in configmap %v", proto, configmapName)
	if configmapName == "" {
		// no configmap configured
		return []ingress.L4Service{}
	}

	_, _, err := k8s.ParseNameNS(configmapName)
	if err != nil {
		glog.Errorf("unexpected error reading configmap %v: %v", configmapName, err)
		return []ingress.L4Service{}
	}

	configmap, err := n.store.GetConfigMap(configmapName)
	if err != nil {
		glog.Errorf("unexpected error reading configmap %v: %v", configmapName, err)
		return []ingress.L4Service{}
	}

	var svcs []ingress.L4Service
	var svcProxyProtocol ingress.ProxyProtocol
	// k -> port to expose
	// v -> <namespace>/<service name>:<port from service to be used>

	rp := []int{
		n.cfg.ListenPorts.HTTP,
		n.cfg.ListenPorts.HTTPS,
		n.cfg.ListenPorts.SSLProxy,
		n.cfg.ListenPorts.Status,
		n.cfg.ListenPorts.Health,
		n.cfg.ListenPorts.Default,
	}
	reserverdPorts := sets.NewInt(rp...)

	for k, v := range configmap.Data {
		externalPort, err := strconv.Atoi(k)
		if err != nil {
			glog.Warningf("%v is not valid as a TCP/UDP port", k)
			continue
		}

		if reserverdPorts.Has(externalPort) {
			glog.Warningf("port %v cannot be used for TCP or UDP services. It is reserved for the Ingress controller", k)
			continue
		}

		nsSvcPort := strings.Split(v, ":")
		if len(nsSvcPort) < 2 {
			glog.Warningf("invalid format (namespace/name:port:[PROXY]:[PROXY]) '%v'", k)
			continue
		}

		nsName := nsSvcPort[0]
		svcPort := nsSvcPort[1]
		svcProxyProtocol.Decode = false
		svcProxyProtocol.Encode = false

		// Proxy protocol is possible if the service is TCP
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
			glog.Warningf("%v", err)
			continue
		}

		svc, err := n.store.GetService(nsName)
		if err != nil {
			glog.Warningf("error getting service %v: %v", nsName, err)
			continue
		}

		var endps []ingress.Endpoint
		targetPort, err := strconv.Atoi(svcPort)
		if err != nil {
			glog.V(3).Infof("searching service %v endpoints using the name '%v'", svcNs, svcName, svcPort)
			for _, sp := range svc.Spec.Ports {
				if sp.Name == svcPort {
					if sp.Protocol == proto {
						endps = getEndpoints(svc, &sp, proto, &healthcheck.Config{}, n.store.GetServiceEndpoints)
						break
					}
				}
			}
		} else {
			// we need to use the TargetPort (where the endpoints are running)
			glog.V(3).Infof("searching service %v/%v endpoints using the target port '%v'", svcNs, svcName, targetPort)
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(targetPort) {
					if sp.Protocol == proto {
						endps = getEndpoints(svc, &sp, proto, &healthcheck.Config{}, n.store.GetServiceEndpoints)
						break
					}
				}
			}
		}

		// stream services cannot contain empty upstreams and there is no
		// default backend equivalent
		if len(endps) == 0 {
			glog.Warningf("service %v/%v does not have any active endpoints for port %v and protocol %v", svcNs, svcName, svcPort, proto)
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
		})
	}

	return svcs
}

// getDefaultUpstream returns an upstream associated with the
// default backend service. In case of error retrieving information
// configure the upstream to return http code 503.
func (n *NGINXController) getDefaultUpstream() *ingress.Backend {
	upstream := &ingress.Backend{
		Name: defUpstreamName,
	}
	svcKey := n.cfg.DefaultService
	svc, err := n.store.GetService(svcKey)
	if err != nil {
		glog.Warningf("unexpected error searching the default backend %v: %v", n.cfg.DefaultService, err)
		upstream.Endpoints = append(upstream.Endpoints, n.DefaultEndpoint())
		return upstream
	}

	endps := getEndpoints(svc, &svc.Spec.Ports[0], apiv1.ProtocolTCP, &healthcheck.Config{}, n.store.GetServiceEndpoints)
	if len(endps) == 0 {
		glog.Warningf("service %v does not have any active endpoints", svcKey)
		endps = []ingress.Endpoint{n.DefaultEndpoint()}
	}

	upstream.Service = svc
	upstream.Endpoints = append(upstream.Endpoints, endps...)
	return upstream
}

// getBackendServers returns a list of Upstream and Server to be used by the backend
// An upstream can be used in multiple servers if the namespace, service name and port are the same
func (n *NGINXController) getBackendServers(ingresses []*extensions.Ingress) ([]*ingress.Backend, []*ingress.Server) {
	du := n.getDefaultUpstream()
	upstreams := n.createUpstreams(ingresses, du)
	servers := n.createServers(ingresses, upstreams, du)

	for _, ing := range ingresses {
		anns, err := n.store.GetIngressAnnotations(ing)
		if err != nil {
			glog.Errorf("unexpected error reading ingress annotations: %v", err)
		}

		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = defServerName
			}
			server := servers[host]
			if server == nil {
				server = servers[defServerName]
			}

			if rule.HTTP == nil &&
				host != defServerName {
				glog.V(3).Infof("ingress rule %v/%v does not contain HTTP rules, using default backend", ing.Namespace, ing.Name)
				continue
			}

			if server.AuthTLSError == "" && anns.CertificateAuth.AuthTLSError != "" {
				server.AuthTLSError = anns.CertificateAuth.AuthTLSError
			}

			if server.CertificateAuth.CAFileName == "" {
				server.CertificateAuth = anns.CertificateAuth
				// It is possible that no CAFileName is found in the secret
				if server.CertificateAuth.CAFileName == "" {
					glog.V(3).Infof("secret %v does not contain 'ca.crt', mutual authentication not enabled - ingress rule %v/%v.", server.CertificateAuth.Secret, ing.Namespace, ing.Name)
				}
			} else {
				glog.V(3).Infof("server %v already contains a mutual authentication configuration - ingress rule %v/%v", server.Hostname, ing.Namespace, ing.Name)
			}

			for _, path := range rule.HTTP.Paths {
				upsName := fmt.Sprintf("%v-%v-%v",
					ing.GetNamespace(),
					path.Backend.ServiceName,
					path.Backend.ServicePort.String())

				ups := upstreams[upsName]

				// if there's no path defined we assume /
				nginxPath := rootLocation
				if path.Path != "" {
					nginxPath = path.Path
				}

				addLoc := true
				for _, loc := range server.Locations {
					if loc.Path == nginxPath {
						addLoc = false

						if !loc.IsDefBackend {
							glog.V(3).Infof("avoiding replacement of ingress rule %v/%v location %v upstream %v (%v)", ing.Namespace, ing.Name, loc.Path, ups.Name, loc.Backend)
							break
						}

						glog.V(3).Infof("replacing ingress rule %v/%v location %v upstream %v (%v)", ing.Namespace, ing.Name, loc.Path, ups.Name, loc.Backend)
						loc.Backend = ups.Name
						loc.IsDefBackend = false
						loc.Backend = ups.Name
						loc.Port = ups.Port
						loc.Service = ups.Service
						loc.Ingress = ing
						loc.BasicDigestAuth = anns.BasicDigestAuth
						loc.ClientBodyBufferSize = anns.ClientBodyBufferSize
						loc.ConfigurationSnippet = anns.ConfigurationSnippet
						loc.CorsConfig = anns.CorsConfig
						loc.ExternalAuth = anns.ExternalAuth
						loc.Proxy = anns.Proxy
						loc.RateLimit = anns.RateLimit
						loc.Redirect = anns.Redirect
						loc.Rewrite = anns.Rewrite
						loc.UpstreamVhost = anns.UpstreamVhost
						loc.VtsFilterKey = anns.VtsFilterKey
						loc.Whitelist = anns.Whitelist
						loc.Denied = anns.Denied
						loc.XForwardedPrefix = anns.XForwardedPrefix
						loc.UsePortInRedirects = anns.UsePortInRedirects
						loc.Connection = anns.Connection
						loc.Logs = anns.Logs
						loc.GRPC = anns.GRPC
						loc.LuaRestyWAF = anns.LuaRestyWAF

						if loc.Redirect.FromToWWW {
							server.RedirectFromToWWW = true
						}
						break
					}
				}
				// is a new location
				if addLoc {
					glog.V(3).Infof("adding location %v in ingress rule %v/%v upstream %v", nginxPath, ing.Namespace, ing.Name, ups.Name)
					loc := &ingress.Location{
						Path:                 nginxPath,
						Backend:              ups.Name,
						IsDefBackend:         false,
						Service:              ups.Service,
						Port:                 ups.Port,
						Ingress:              ing,
						BasicDigestAuth:      anns.BasicDigestAuth,
						ClientBodyBufferSize: anns.ClientBodyBufferSize,
						ConfigurationSnippet: anns.ConfigurationSnippet,
						CorsConfig:           anns.CorsConfig,
						ExternalAuth:         anns.ExternalAuth,
						Proxy:                anns.Proxy,
						RateLimit:            anns.RateLimit,
						Redirect:             anns.Redirect,
						Rewrite:              anns.Rewrite,
						UpstreamVhost:        anns.UpstreamVhost,
						VtsFilterKey:         anns.VtsFilterKey,
						Whitelist:            anns.Whitelist,
						Denied:               anns.Denied,
						XForwardedPrefix:     anns.XForwardedPrefix,
						UsePortInRedirects:   anns.UsePortInRedirects,
						Connection:           anns.Connection,
						Logs:                 anns.Logs,
						GRPC:                 anns.GRPC,
						LuaRestyWAF:          anns.LuaRestyWAF,
					}

					if loc.Redirect.FromToWWW {
						server.RedirectFromToWWW = true
					}
					server.Locations = append(server.Locations, loc)
				}

				if ups.SessionAffinity.AffinityType == "" {
					ups.SessionAffinity.AffinityType = anns.SessionAffinity.Type
				}

				if anns.SessionAffinity.Type == "cookie" {
					ups.SessionAffinity.CookieSessionAffinity.Name = anns.SessionAffinity.Cookie.Name
					ups.SessionAffinity.CookieSessionAffinity.Hash = anns.SessionAffinity.Cookie.Hash

					locs := ups.SessionAffinity.CookieSessionAffinity.Locations
					if _, ok := locs[host]; !ok {
						locs[host] = []string{}
					}

					locs[host] = append(locs[host], path.Path)
				}
			}
		}
	}

	aUpstreams := make([]*ingress.Backend, 0, len(upstreams))

	for _, upstream := range upstreams {
		isHTTPSfrom := []*ingress.Server{}
		for _, server := range servers {
			for _, location := range server.Locations {
				if upstream.Name == location.Backend {
					if len(upstream.Endpoints) == 0 {
						glog.V(3).Infof("upstream %v does not have any active endpoints.", upstream.Name)
						location.Backend = ""

						// check if the location contains endpoints and a custom default backend
						if location.DefaultBackend != nil {
							sp := location.DefaultBackend.Spec.Ports[0]
							endps := getEndpoints(location.DefaultBackend, &sp, apiv1.ProtocolTCP, &healthcheck.Config{}, n.store.GetServiceEndpoints)
							if len(endps) > 0 {
								glog.V(3).Infof("using custom default backend in server %v location %v (service %v/%v)",
									server.Hostname, location.Path, location.DefaultBackend.Namespace, location.DefaultBackend.Name)
								nb := upstream.DeepCopy()
								name := fmt.Sprintf("custom-default-backend-%v", upstream.Name)
								nb.Name = name
								nb.Endpoints = endps
								aUpstreams = append(aUpstreams, nb)
								location.Backend = name
							}
						}
					}

					// Configure Backends[].SSLPassthrough
					if server.SSLPassthrough {
						if location.Path == rootLocation {
							if location.Backend == defUpstreamName {
								glog.Warningf("ignoring ssl passthrough of %v as it doesn't have a default backend (root context)", server.Hostname)
								continue
							}

							isHTTPSfrom = append(isHTTPSfrom, server)
						}
					}
				}
			}
		}

		if len(isHTTPSfrom) > 0 {
			upstream.SSLPassthrough = true
		}
	}

	// create the list of upstreams and skip those without endpoints
	for _, upstream := range upstreams {
		if len(upstream.Endpoints) == 0 {
			continue
		}
		aUpstreams = append(aUpstreams, upstream)
	}

	if n.cfg.SortBackends {
		sort.SliceStable(aUpstreams, func(a, b int) bool {
			return aUpstreams[a].Name < aUpstreams[b].Name
		})
	}

	aServers := make([]*ingress.Server, 0, len(servers))
	for _, value := range servers {
		sort.SliceStable(value.Locations, func(i, j int) bool {
			return value.Locations[i].Path > value.Locations[j].Path
		})
		aServers = append(aServers, value)
	}

	sort.SliceStable(aServers, func(i, j int) bool {
		return aServers[i].Hostname < aServers[j].Hostname
	})

	return aUpstreams, aServers
}

// createUpstreams creates the NGINX upstreams for each service referenced in
// Ingress rules. The servers inside the upstream are endpoints.
func (n *NGINXController) createUpstreams(data []*extensions.Ingress, du *ingress.Backend) map[string]*ingress.Backend {
	upstreams := make(map[string]*ingress.Backend)
	upstreams[defUpstreamName] = du

	for _, ing := range data {
		anns, err := n.store.GetIngressAnnotations(ing)
		if err != nil {
			glog.Errorf("unexpected error reading ingress annotations: %v", err)
		}

		var defBackend string
		if ing.Spec.Backend != nil {
			defBackend = fmt.Sprintf("%v-%v-%v",
				ing.GetNamespace(),
				ing.Spec.Backend.ServiceName,
				ing.Spec.Backend.ServicePort.String())

			glog.V(3).Infof("creating upstream %v", defBackend)
			upstreams[defBackend] = newUpstream(defBackend)
			if !upstreams[defBackend].Secure {
				upstreams[defBackend].Secure = anns.SecureUpstream.Secure
			}
			if upstreams[defBackend].SecureCACert.Secret == "" {
				upstreams[defBackend].SecureCACert = anns.SecureUpstream.CACert
			}
			if upstreams[defBackend].UpstreamHashBy == "" {
				upstreams[defBackend].UpstreamHashBy = anns.UpstreamHashBy
			}
			if upstreams[defBackend].LoadBalancing == "" {
				upstreams[defBackend].LoadBalancing = anns.LoadBalancing
			}

			svcKey := fmt.Sprintf("%v/%v", ing.GetNamespace(), ing.Spec.Backend.ServiceName)

			// Add the service cluster endpoint as the upstream instead of individual endpoints
			// if the serviceUpstream annotation is enabled
			if anns.ServiceUpstream {
				endpoint, err := n.getServiceClusterEndpoint(svcKey, ing.Spec.Backend)
				if err != nil {
					glog.Errorf("Failed to get service cluster endpoint for service %s: %v", svcKey, err)
				} else {
					upstreams[defBackend].Endpoints = []ingress.Endpoint{endpoint}
				}
			}

			if len(upstreams[defBackend].Endpoints) == 0 {
				endps, err := n.serviceEndpoints(svcKey, ing.Spec.Backend.ServicePort.String(), &anns.HealthCheck)
				upstreams[defBackend].Endpoints = append(upstreams[defBackend].Endpoints, endps...)
				if err != nil {
					glog.Warningf("error creating upstream %v: %v", defBackend, err)
				}
			}

		}

		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				name := fmt.Sprintf("%v-%v-%v",
					ing.GetNamespace(),
					path.Backend.ServiceName,
					path.Backend.ServicePort.String())

				if _, ok := upstreams[name]; ok {
					continue
				}

				glog.V(3).Infof("creating upstream %v", name)
				upstreams[name] = newUpstream(name)
				upstreams[name].Port = path.Backend.ServicePort

				if !upstreams[name].Secure {
					upstreams[name].Secure = anns.SecureUpstream.Secure
				}

				if upstreams[name].SecureCACert.Secret == "" {
					upstreams[name].SecureCACert = anns.SecureUpstream.CACert
				}

				if upstreams[name].UpstreamHashBy == "" {
					upstreams[name].UpstreamHashBy = anns.UpstreamHashBy
				}

				if upstreams[name].LoadBalancing == "" {
					upstreams[name].LoadBalancing = anns.LoadBalancing
				}

				svcKey := fmt.Sprintf("%v/%v", ing.GetNamespace(), path.Backend.ServiceName)

				// Add the service cluster endpoint as the upstream instead of individual endpoints
				// if the serviceUpstream annotation is enabled
				if anns.ServiceUpstream {
					endpoint, err := n.getServiceClusterEndpoint(svcKey, &path.Backend)
					if err != nil {
						glog.Errorf("failed to get service cluster endpoint for service %s: %v", svcKey, err)
					} else {
						upstreams[name].Endpoints = []ingress.Endpoint{endpoint}
					}
				}

				if len(upstreams[name].Endpoints) == 0 {
					endp, err := n.serviceEndpoints(svcKey, path.Backend.ServicePort.String(), &anns.HealthCheck)
					if err != nil {
						glog.Warningf("error obtaining service endpoints: %v", err)
						continue
					}
					upstreams[name].Endpoints = endp
				}

				s, err := n.store.GetService(svcKey)
				if err != nil {
					glog.Warningf("error obtaining service: %v", err)
					continue
				}

				upstreams[name].Service = s
			}
		}
	}

	return upstreams
}

func (n *NGINXController) getServiceClusterEndpoint(svcKey string, backend *extensions.IngressBackend) (endpoint ingress.Endpoint, err error) {
	svc, err := n.store.GetService(svcKey)
	if err != nil {
		return endpoint, fmt.Errorf("service %v does not exist", svcKey)
	}

	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		return endpoint, fmt.Errorf("No ClusterIP found for service %s", svcKey)
	}

	endpoint.Address = svc.Spec.ClusterIP

	// If the service port in the ingress uses a name, lookup
	// the actual port in the service spec
	if backend.ServicePort.Type == intstr.String {
		var port int32 = -1
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Name == backend.ServicePort.String() {
				port = svcPort.Port
				break
			}
		}
		if port == -1 {
			return endpoint, fmt.Errorf("no port mapped for service %s and port name %s", svc.Name, backend.ServicePort.String())
		}
		endpoint.Port = fmt.Sprintf("%d", port)
	} else {
		endpoint.Port = backend.ServicePort.String()
	}

	return endpoint, err
}

// serviceEndpoints returns the upstream servers (endpoints) associated
// to a service.
func (n *NGINXController) serviceEndpoints(svcKey, backendPort string,
	hz *healthcheck.Config) ([]ingress.Endpoint, error) {
	svc, err := n.store.GetService(svcKey)

	var upstreams []ingress.Endpoint
	if err != nil {
		return upstreams, fmt.Errorf("error getting service %v from the cache: %v", svcKey, err)
	}

	glog.V(3).Infof("obtaining port information for service %v", svcKey)
	for _, servicePort := range svc.Spec.Ports {
		// targetPort could be a string, use the name or the port (int)
		if strconv.Itoa(int(servicePort.Port)) == backendPort ||
			servicePort.TargetPort.String() == backendPort ||
			servicePort.Name == backendPort {

			endps := getEndpoints(svc, &servicePort, apiv1.ProtocolTCP, hz, n.store.GetServiceEndpoints)
			if len(endps) == 0 {
				glog.Warningf("service %v does not have any active endpoints", svcKey)
			}

			if n.cfg.SortBackends {
				sort.SliceStable(endps, func(i, j int) bool {
					iName := endps[i].Address
					jName := endps[j].Address
					if iName != jName {
						return iName < jName
					}

					return endps[i].Port < endps[j].Port
				})
			}
			upstreams = append(upstreams, endps...)
			break
		}
	}

	// Ingress with an ExternalName service and no port defined in the service.
	if len(svc.Spec.Ports) == 0 && svc.Spec.Type == apiv1.ServiceTypeExternalName {
		externalPort, err := strconv.Atoi(backendPort)
		if err != nil {
			glog.Warningf("only numeric ports are allowed in ExternalName services: %v is not valid as a TCP/UDP port", backendPort)
			return upstreams, nil
		}

		servicePort := apiv1.ServicePort{
			Protocol:   "TCP",
			Port:       int32(externalPort),
			TargetPort: intstr.FromString(backendPort),
		}
		endps := getEndpoints(svc, &servicePort, apiv1.ProtocolTCP, hz, n.store.GetServiceEndpoints)
		if len(endps) == 0 {
			glog.Warningf("service %v does not have any active endpoints", svcKey)
			return upstreams, nil
		}

		upstreams = append(upstreams, endps...)
		return upstreams, nil
	}

	if !n.cfg.SortBackends {
		rand.Seed(time.Now().UnixNano())
		for i := range upstreams {
			j := rand.Intn(i + 1)
			upstreams[i], upstreams[j] = upstreams[j], upstreams[i]
		}
	}

	return upstreams, nil
}

// createServers initializes a map that contains information about the list of
// FDQN referenced by ingress rules and the common name field in the referenced
// SSL certificates. Each server is configured with location / using a default
// backend specified by the user or the one inside the ingress spec.
func (n *NGINXController) createServers(data []*extensions.Ingress,
	upstreams map[string]*ingress.Backend,
	du *ingress.Backend) map[string]*ingress.Server {

	servers := make(map[string]*ingress.Server, len(data))
	// If a server has a hostname equivalent to a pre-existing alias, then we
	// remove the alias to avoid conflicts.
	aliases := make(map[string]string, len(data))

	bdef := n.store.GetDefaultBackend()
	ngxProxy := proxy.Config{
		BodySize:          bdef.ProxyBodySize,
		ConnectTimeout:    bdef.ProxyConnectTimeout,
		SendTimeout:       bdef.ProxySendTimeout,
		ReadTimeout:       bdef.ProxyReadTimeout,
		BufferSize:        bdef.ProxyBufferSize,
		CookieDomain:      bdef.ProxyCookieDomain,
		CookiePath:        bdef.ProxyCookiePath,
		NextUpstream:      bdef.ProxyNextUpstream,
		NextUpstreamTries: bdef.ProxyNextUpstreamTries,
		RequestBuffering:  bdef.ProxyRequestBuffering,
		ProxyRedirectFrom: bdef.ProxyRedirectFrom,
		ProxyBuffering:    bdef.ProxyBuffering,
	}

	// generated on Start() with createDefaultSSLCertificate()
	defaultPemFileName := n.cfg.FakeCertificatePath
	defaultPemSHA := n.cfg.FakeCertificateSHA

	// Tries to fetch the default Certificate from nginx configuration.
	// If it does not exists, use the ones generated on Start()
	defaultCertificate, err := n.store.GetLocalSSLCert(n.cfg.DefaultSSLCertificate)
	if err == nil {
		defaultPemFileName = defaultCertificate.PemFileName
		defaultPemSHA = defaultCertificate.PemSHA
	}

	// initialize the default server
	servers[defServerName] = &ingress.Server{
		Hostname:       defServerName,
		SSLCertificate: defaultPemFileName,
		SSLPemChecksum: defaultPemSHA,
		Locations: []*ingress.Location{
			{
				Path:         rootLocation,
				IsDefBackend: true,
				Backend:      du.Name,
				Proxy:        ngxProxy,
				Service:      du.Service,
			},
		}}

	// initialize all the servers
	for _, ing := range data {
		anns, err := n.store.GetIngressAnnotations(ing)
		if err != nil {
			glog.Errorf("unexpected error reading ingress annotations: %v", err)
		}

		// default upstream server
		un := du.Name

		if ing.Spec.Backend != nil {
			// replace default backend
			defUpstream := fmt.Sprintf("%v-%v-%v", ing.GetNamespace(), ing.Spec.Backend.ServiceName, ing.Spec.Backend.ServicePort.String())
			if backendUpstream, ok := upstreams[defUpstream]; ok {
				un = backendUpstream.Name

				// Special case:
				// ingress only with a backend and no rules
				// this case defines a "catch all" server
				defLoc := servers[defServerName].Locations[0]
				if defLoc.IsDefBackend && len(ing.Spec.Rules) == 0 {
					defLoc.IsDefBackend = false
					defLoc.Backend = backendUpstream.Name
					defLoc.Service = backendUpstream.Service
					defLoc.Ingress = ing

					// we need to use the ingress annotations
					defLoc.BasicDigestAuth = anns.BasicDigestAuth
					defLoc.ClientBodyBufferSize = anns.ClientBodyBufferSize
					defLoc.ConfigurationSnippet = anns.ConfigurationSnippet
					defLoc.CorsConfig = anns.CorsConfig
					defLoc.ExternalAuth = anns.ExternalAuth
					defLoc.Proxy = anns.Proxy
					defLoc.RateLimit = anns.RateLimit
					// TODO: Redirect and rewrite can affect the catch all behavior. Don't use this annotations for now
					// defLoc.Redirect = anns.Redirect
					// defLoc.Rewrite = anns.Rewrite
					defLoc.UpstreamVhost = anns.UpstreamVhost
					defLoc.VtsFilterKey = anns.VtsFilterKey
					defLoc.Whitelist = anns.Whitelist
					defLoc.Denied = anns.Denied
					defLoc.GRPC = anns.GRPC
					defLoc.LuaRestyWAF = anns.LuaRestyWAF
				}
			}
		}

		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = defServerName
			}
			if _, ok := servers[host]; ok {
				// server already configured
				continue
			}

			servers[host] = &ingress.Server{
				Hostname: host,
				Locations: []*ingress.Location{
					{
						Path:         rootLocation,
						IsDefBackend: true,
						Backend:      un,
						Proxy:        ngxProxy,
						Service:      &apiv1.Service{},
					},
				},
				SSLPassthrough: anns.SSLPassthrough,
				SSLCiphers:     anns.SSLCiphers,
			}
		}
	}

	// configure default location, alias, and SSL
	for _, ing := range data {
		anns, err := n.store.GetIngressAnnotations(ing)
		if err != nil {
			glog.Errorf("unexpected error reading ingress annotations: %v", err)
		}

		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = defServerName
			}

			// setup server aliases
			if anns.Alias != "" {
				if servers[host].Alias == "" {
					servers[host].Alias = anns.Alias
					if _, ok := aliases["Alias"]; !ok {
						aliases["Alias"] = host
					}
				} else {
					glog.Warningf("ingress %v/%v for host %v contains an Alias but one has already been configured.",
						ing.Namespace, ing.Name, host)
				}
			}

			//notifying the user that it has already been configured.
			if servers[host].ServerSnippet != "" && anns.ServerSnippet != "" {
				glog.Warningf("ingress %v/%v for host %v contains a Server Snippet section that it has already been configured.",
					ing.Namespace, ing.Name, host)
			}

			// only add a server snippet if the server does not have one previously configured
			if servers[host].ServerSnippet == "" && anns.ServerSnippet != "" {
				servers[host].ServerSnippet = anns.ServerSnippet
			}

			// only add ssl ciphers if the server does not have one previously configured
			if servers[host].SSLCiphers == "" && anns.SSLCiphers != "" {
				servers[host].SSLCiphers = anns.SSLCiphers
			}

			// only add a certificate if the server does not have one previously configured
			if servers[host].SSLCertificate != "" {
				continue
			}

			if len(ing.Spec.TLS) == 0 {
				glog.V(3).Infof("ingress %v/%v for host %v does not contains a TLS section", ing.Namespace, ing.Name, host)
				continue
			}

			tlsSecretName := extractTLSSecretName(host, ing, n.store.GetLocalSSLCert)

			if tlsSecretName == "" {
				glog.V(3).Infof("host %v is listed on tls section but secretName is empty. Using default cert", host)
				servers[host].SSLCertificate = defaultPemFileName
				servers[host].SSLPemChecksum = defaultPemSHA
				continue
			}

			key := fmt.Sprintf("%v/%v", ing.Namespace, tlsSecretName)
			cert, err := n.store.GetLocalSSLCert(key)
			if err != nil {
				glog.Warningf("ssl certificate \"%v\" does not exist in local store", key)
				continue
			}

			err = cert.Certificate.VerifyHostname(host)
			if err != nil {
				glog.Warningf("unexpected error validating SSL certificate %v for host %v. Reason: %v", key, host, err)
				glog.Warningf("Validating certificate against DNS names. This will be deprecated in a future version.")
				// check the common name field
				// https://github.com/golang/go/issues/22922
				err := verifyHostname(host, cert.Certificate)
				if err != nil {
					glog.Warningf("ssl certificate %v does not contain a Common Name or Subject Alternative Name for host %v. Reason: %v", key, host, err)
					continue
				}
			}

			servers[host].SSLCertificate = cert.PemFileName
			servers[host].SSLFullChainCertificate = cert.FullChainPemFileName
			servers[host].SSLPemChecksum = cert.PemSHA
			servers[host].SSLExpireTime = cert.ExpireTime

			if cert.ExpireTime.Before(time.Now().Add(240 * time.Hour)) {
				glog.Warningf("ssl certificate for host %v is about to expire in 10 days", host)
			}
		}
	}

	for alias, host := range aliases {
		if _, ok := servers[alias]; ok {
			glog.Warningf("There is a conflict with server hostname '%v' and alias '%v' (in server %v). Removing alias to avoid conflicts.", alias, host)
			servers[host].Alias = ""
		}
	}

	return servers
}

func (n *NGINXController) isForceReload() bool {
	return atomic.LoadInt32(&n.forceReload) != 0
}

// SetForceReload sets if the ingress controller should be reloaded or not
func (n *NGINXController) SetForceReload(shouldReload bool) {
	if shouldReload {
		atomic.StoreInt32(&n.forceReload, 1)
		n.syncQueue.Enqueue(&extensions.Ingress{})
	} else {
		atomic.StoreInt32(&n.forceReload, 0)
	}
}

// extractTLSSecretName returns the name of the secret that
// contains a SSL certificate for a particular hostname.
// In case there is no match, an empty string is returned.
func extractTLSSecretName(host string, ing *extensions.Ingress,
	getLocalSSLCert func(string) (*ingress.SSLCert, error)) string {
	if ing == nil {
		return ""
	}

	for _, tls := range ing.Spec.TLS {
		if sets.NewString(tls.Hosts...).Has(host) {
			return tls.SecretName
		}
	}

	// contains a TLS section but none of the host match or there
	// is no hosts in the TLS section. As last resort we valide
	// the host against the certificate and we use it if is valid
	for _, tls := range ing.Spec.TLS {
		key := fmt.Sprintf("%v/%v", ing.Namespace, tls.SecretName)
		cert, err := getLocalSSLCert(key)
		if err != nil {
			glog.Warningf("ssl certificate \"%v\" does not exist in local store", key)
			continue
		}

		if cert == nil {
			continue
		}

		if sets.NewString(cert.CN...).Has(host) {
			return tls.SecretName
		}
	}

	return ""
}
