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

	"github.com/mitchellh/hashstructure/v2"
	apiv1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/canary"
	"k8s.io/ingress-nginx/internal/ingress/annotations/log"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxy"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/ingressclass"
	"k8s.io/ingress-nginx/internal/ingress/controller/store"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/inspector"
	"k8s.io/ingress-nginx/internal/ingress/metric/collectors"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	utilingress "k8s.io/ingress-nginx/pkg/util/ingress"
	"k8s.io/klog/v2"
)

const (
	defUpstreamName             = "upstream-default-backend"
	defServerName               = "_"
	rootLocation                = "/"
	emptyZone                   = ""
	orphanMetricLabelNoService  = "no-service"
	orphanMetricLabelNoEndpoint = "no-endpoint"
)

// Configuration contains all the settings required by an Ingress controller
type Configuration struct {
	APIServerHost               string
	RootCAFile                  string
	KubeConfigFile              string
	Client                      clientset.Interface
	ResyncPeriod                time.Duration
	ConfigMapName               string
	DefaultService              string
	Namespace                   string
	WatchNamespaceSelector      labels.Selector
	TCPConfigMapName            string
	UDPConfigMapName            string
	DefaultSSLCertificate       string
	PublishService              string
	PublishStatusAddress        string
	UpdateStatus                bool
	UseNodeInternalIP           bool
	ElectionID                  string
	ElectionTTL                 time.Duration
	UpdateStatusOnShutdown      bool
	HealthCheckHost             string
	ListenPorts                 *ngx_config.ListenPorts
	DisableServiceExternalName  bool
	EnableSSLPassthrough        bool
	DisableLeaderElection       bool
	EnableProfiling             bool
	EnableMetrics               bool
	MetricsPerHost              bool
	MetricsPerUndefinedHost     bool
	MetricsBuckets              *collectors.HistogramBuckets
	MetricsBucketFactor         float64
	MetricsMaxBuckets           uint32
	ReportStatusClasses         bool
	ExcludeSocketMetrics        []string
	FakeCertificate             *ingress.SSLCert
	SyncRateLimit               float32
	DisableCatchAll             bool
	IngressClassConfiguration   *ingressclass.Configuration
	ValidationWebhook           string
	ValidationWebhookCertPath   string
	ValidationWebhookKeyPath    string
	DisableFullValidationTest   bool
	GlobalExternalAuth          *ngx_config.GlobalExternalAuth
	MaxmindEditionFiles         *[]string
	MonitorMaxBatchSize         int
	PostShutdownGracePeriod     int
	ShutdownGracePeriod         int
	InternalLoggerAddress       string
	IsChroot                    bool
	DeepInspector               bool
	DynamicConfigurationRetries int
	DisableSyncEvents           bool
	EnableTopologyAwareRouting  bool
	ConfigurationTemplateEngine string
}

func getIngressPodZone(svc *apiv1.Service) string {
	svcKey := k8s.MetaNamespaceKey(svc)
	if svcZoneAnnotation, ok := svc.ObjectMeta.GetAnnotations()[apiv1.AnnotationTopologyMode]; ok {
		if strings.EqualFold(svcZoneAnnotation, "auto") {
			if foundZone, ok := k8s.IngressNodeDetails.GetLabels()[apiv1.LabelTopologyZone]; ok {
				klog.V(3).Infof("Svc has topology aware annotation enabled, try to use zone %q where controller pod is running for Service %q ", foundZone, svcKey)
				return foundZone
			}
		}
	}
	return emptyZone
}

// GetPublishService returns the Service used to set the load-balancer status of Ingresses.
func (n *NGINXController) GetPublishService() *apiv1.Service {
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
	n.metricCollector.SetSSLInfo(servers)

	if n.runningConfig.Equal(pcfg) {
		klog.V(3).Infof("No configuration change detected, skipping backend reload")
		return nil
	}

	n.metricCollector.SetHosts(hosts)

	if !utilingress.IsDynamicConfigurationEnough(pcfg, n.runningConfig) {
		klog.InfoS("Configuration changes detected, backend reload required")

		hash, err := hashstructure.Hash(pcfg, hashstructure.FormatV1, &hashstructure.HashOptions{
			TagName: "json",
		})
		if err != nil {
			klog.Errorf("unexpected error hashing configuration: %v", err)
		}

		pcfg.ConfigurationChecksum = fmt.Sprintf("%v", hash)

		err = n.OnUpdate(*pcfg)
		if err != nil {
			n.metricCollector.IncReloadErrorCount()
			n.metricCollector.ConfigSuccess(hash, false)
			klog.Errorf("Unexpected failure reloading the backend:\n%v", err)
			n.recorder.Eventf(k8s.IngressPodDetails, apiv1.EventTypeWarning, "RELOAD", fmt.Sprintf("Error reloading NGINX: %v", err))
			return err
		}

		klog.InfoS("Backend successfully reloaded")
		n.metricCollector.ConfigSuccess(hash, true)
		n.metricCollector.IncReloadCount()

		n.recorder.Eventf(k8s.IngressPodDetails, apiv1.EventTypeNormal, "RELOAD", "NGINX reload triggered due to a change in configuration")
	}

	isFirstSync := n.runningConfig.Equal(&ingress.Configuration{})
	if isFirstSync {
		// For the initial sync it always takes some time for NGINX to start listening
		// For large configurations it might take a while so we loop and back off
		klog.InfoS("Initial sync, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	retry := wait.Backoff{
		Steps:    1 + n.cfg.DynamicConfigurationRetries,
		Duration: time.Second,
		Factor:   1.3,
		Jitter:   0.1,
	}

	retriesRemaining := retry.Steps
	err := wait.ExponentialBackoff(retry, func() (bool, error) {
		err := n.configureDynamically(pcfg)
		if err == nil {
			klog.V(2).Infof("Dynamic reconfiguration succeeded.")
			return true, nil
		}
		retriesRemaining--
		if retriesRemaining > 0 {
			klog.Warningf("Dynamic reconfiguration failed (retrying; %d retries left): %v", retriesRemaining, err)
			return false, nil
		}
		klog.Warningf("Dynamic reconfiguration failed: %v", err)
		return false, err
	})
	if err != nil {
		klog.Errorf("Unexpected failure reconfiguring NGINX:\n%v", err)
		return err
	}

	ri := utilingress.GetRemovedIngresses(n.runningConfig, pcfg)
	rc := utilingress.GetRemovedCertificateSerialNumbers(n.runningConfig, pcfg)
	n.metricCollector.RemoveMetrics(ri, rc)

	n.runningConfig = pcfg

	return nil
}

// GetWarnings returns a list of warnings an Ingress gets when being created.
// The warnings are going to be used in an admission webhook, and they represent
// a list of messages that users need to be aware (like deprecation notices)
// when creating a new ingress object
func (n *NGINXController) CheckWarning(ing *networking.Ingress) ([]string, error) {
	warnings := make([]string, 0)

	deprecatedAnnotations := sets.NewString()
	deprecatedAnnotations.Insert(
		"enable-influxdb",
		"influxdb-measurement",
		"influxdb-port",
		"influxdb-host",
		"influxdb-server-name",
		"secure-verify-ca-secret",
	)

	// Skip checks if the ingress is marked as deleted
	if !ing.DeletionTimestamp.IsZero() {
		return warnings, nil
	}

	anns := ing.GetAnnotations()
	for k := range anns {
		trimmedkey := strings.TrimPrefix(k, parser.AnnotationsPrefix+"/")
		if deprecatedAnnotations.Has(trimmedkey) {
			warnings = append(warnings, fmt.Sprintf("annotation %s is deprecated", k))
		}
	}

	// Add each validation as a single warning
	// rikatz: I know this is somehow a duplicated code from CheckIngress, but my goal was to deliver fast warning on this behavior. We
	// can and should, tho, simplify this in the near future
	if err := inspector.ValidatePathType(ing); err != nil {
		if errs, is := err.(interface{ Unwrap() []error }); is {
			for _, errW := range errs.Unwrap() {
				warnings = append(warnings, errW.Error())
			}
		} else {
			warnings = append(warnings, err.Error())
		}
	}

	return warnings, nil
}

// CheckIngress returns an error in case the provided ingress, when added
// to the current configuration, generates an invalid configuration
func (n *NGINXController) CheckIngress(ing *networking.Ingress) error {
	startCheck := time.Now().UnixNano() / 1000000

	if ing == nil {
		// no ingress to add, no state change
		return nil
	}

	// Skip checks if the ingress is marked as deleted
	if !ing.DeletionTimestamp.IsZero() {
		return nil
	}

	if n.cfg.DeepInspector {
		if err := inspector.DeepInspect(ing); err != nil {
			return fmt.Errorf("invalid object: %w", err)
		}
	}

	// Do not attempt to validate an ingress that's not meant to be controlled by the current instance of the controller.
	if ingressClass, err := n.store.GetIngressClass(ing, n.cfg.IngressClassConfiguration); ingressClass == "" {
		klog.Warningf("ignoring ingress %v in %v based on annotation %v: %v", ing.Name, ing.ObjectMeta.Namespace, ingressClass, err)
		return nil
	}

	if n.cfg.Namespace != "" && ing.ObjectMeta.Namespace != n.cfg.Namespace {
		klog.Warningf("ignoring ingress %v in namespace %v different from the namespace watched %s", ing.Name, ing.ObjectMeta.Namespace, n.cfg.Namespace)
		return nil
	}

	if n.cfg.DisableCatchAll && ing.Spec.DefaultBackend != nil {
		return fmt.Errorf("this deployment is trying to create a catch-all ingress while DisableCatchAll flag is set to true. Remove '.spec.defaultBackend' or set DisableCatchAll flag to false")
	}
	startRender := time.Now().UnixNano() / 1000000
	cfg := n.store.GetBackendConfiguration()
	cfg.Resolver = n.resolver

	// Adds the pathType Validation
	if cfg.StrictValidatePathType {
		if err := inspector.ValidatePathType(ing); err != nil {
			return fmt.Errorf("ingress contains invalid paths: %w", err)
		}
	}

	var arrayBadWords []string

	if cfg.AnnotationValueWordBlocklist != "" {
		arrayBadWords = strings.Split(strings.TrimSpace(cfg.AnnotationValueWordBlocklist), ",")
	}

	for key, value := range ing.ObjectMeta.GetAnnotations() {
		if parser.AnnotationsPrefix != parser.DefaultAnnotationsPrefix {
			if strings.HasPrefix(key, fmt.Sprintf("%s/", parser.DefaultAnnotationsPrefix)) {
				return fmt.Errorf("this deployment has a custom annotation prefix defined. Use '%s' instead of '%s'", parser.AnnotationsPrefix, parser.DefaultAnnotationsPrefix)
			}
		}

		if strings.HasPrefix(key, fmt.Sprintf("%s/", parser.AnnotationsPrefix)) && len(arrayBadWords) != 0 {
			for _, forbiddenvalue := range arrayBadWords {
				if strings.Contains(value, strings.TrimSpace(forbiddenvalue)) {
					return fmt.Errorf("%s annotation contains invalid word %s", key, forbiddenvalue)
				}
			}
		}

		if !cfg.AllowSnippetAnnotations && strings.HasSuffix(key, "-snippet") {
			return fmt.Errorf("%s annotation cannot be used. Snippet directives are disabled by the Ingress administrator", key)
		}
	}

	k8s.SetDefaultNGINXPathType(ing)

	allIngresses := n.store.ListIngresses()

	filter := func(toCheck *ingress.Ingress) bool {
		return toCheck.ObjectMeta.Namespace == ing.ObjectMeta.Namespace &&
			toCheck.ObjectMeta.Name == ing.ObjectMeta.Name
	}
	ings := store.FilterIngresses(allIngresses, filter)
	parsed, err := annotations.NewAnnotationExtractor(n.store).Extract(ing)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}
	ings = append(ings, &ingress.Ingress{
		Ingress:           *ing,
		ParsedAnnotations: parsed,
	})
	startTest := time.Now().UnixNano() / 1000000
	_, servers, pcfg := n.getConfiguration(ings)

	err = checkOverlap(ing, servers)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}
	testedSize := len(ings)
	if n.cfg.DisableFullValidationTest {
		_, _, pcfg = n.getConfiguration(ings[len(ings)-1:])
		testedSize = 1
	}

	content, err := n.generateTemplate(cfg, *pcfg)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}

	/* Deactivated to mitigate CVE-2025-1974
	// TODO: Implement sandboxing so this test can be done safely
	err = n.testTemplate(content)
	if err != nil {
		n.metricCollector.IncCheckErrorCount(ing.ObjectMeta.Namespace, ing.Name)
		return err
	}
	*/

	n.metricCollector.IncCheckCount(ing.ObjectMeta.Namespace, ing.Name)
	endCheck := time.Now().UnixNano() / 1000000
	n.metricCollector.SetAdmissionMetrics(
		float64(testedSize),
		float64(endCheck-startTest)/1000,
		float64(len(ings)),
		float64(startTest-startRender)/1000,
		float64(len(content)),
		float64(endCheck-startCheck)/1000,
	)
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

	svcs := make([]ingress.L4Service, 0, len(configmap.Data))
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

	reservedPorts := sets.NewInt(rp...)
	// svcRef format: <(str)namespace>/<(str)service>:<(intstr)port>[:<("PROXY")decode>:<("PROXY")encode>]
	for port, svcRef := range configmap.Data {
		externalPort, err := strconv.Atoi(port) // #nosec
		if err != nil {
			klog.Warningf("%q is not a valid %v port number", port, proto)
			continue
		}
		if reservedPorts.Has(externalPort) {
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
			if len(nsSvcPort) >= 3 && strings.EqualFold(nsSvcPort[2], "PROXY") {
				svcProxyProtocol.Decode = true
			}
			if len(nsSvcPort) == 4 && strings.EqualFold(nsSvcPort[3], "PROXY") {
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
		/* #nosec */
		targetPort, err := strconv.Atoi(svcPort) // #nosec
		var zone string
		if n.cfg.EnableTopologyAwareRouting {
			zone = getIngressPodZone(svc)
		} else {
			zone = emptyZone
		}

		if err != nil {
			// not a port number, fall back to using port name
			klog.V(3).Infof("Searching Endpoints with %v port name %q for Service %q", proto, svcPort, nsName)
			for i := range svc.Spec.Ports {
				sp := svc.Spec.Ports[i]
				if sp.Name == svcPort {
					if sp.Protocol == proto {
						endps = getEndpointsFromSlices(svc, &sp, proto, zone, n.store.GetServiceEndpointsSlices)
						break
					}
				}
			}
		} else {
			klog.V(3).Infof("Searching Endpoints with %v port number %d for Service %q", proto, targetPort, nsName)
			for i := range svc.Spec.Ports {
				sp := svc.Spec.Ports[i]
				//nolint:gosec // Ignore G109 error
				if sp.Port == int32(targetPort) {
					if sp.Protocol == proto {
						endps = getEndpointsFromSlices(svc, &sp, proto, zone, n.store.GetServiceEndpointsSlices)
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

	if svcKey == "" {
		upstream.Endpoints = append(upstream.Endpoints, n.DefaultEndpoint())
		return upstream
	}

	svc, err := n.store.GetService(svcKey)
	if err != nil {
		klog.Warningf("Error getting default backend %q: %v", svcKey, err)
		upstream.Endpoints = append(upstream.Endpoints, n.DefaultEndpoint())
		return upstream
	}
	var zone string
	if n.cfg.EnableTopologyAwareRouting {
		zone = getIngressPodZone(svc)
	} else {
		zone = emptyZone
	}
	endps := getEndpointsFromSlices(svc, &svc.Spec.Ports[0], apiv1.ProtocolTCP, zone, n.store.GetServiceEndpointsSlices)
	if len(endps) == 0 {
		klog.Warningf("Service %q does not have any active Endpoint", svcKey)
		endps = []ingress.Endpoint{n.DefaultEndpoint()}
	}

	upstream.Service = svc
	upstream.Endpoints = append(upstream.Endpoints, endps...)
	return upstream
}

// getConfiguration returns the configuration matching the standard kubernetes ingress
func (n *NGINXController) getConfiguration(ingresses []*ingress.Ingress) (sets.Set[string], []*ingress.Server, *ingress.Configuration) {
	upstreams, servers := n.getBackendServers(ingresses)
	var passUpstreams []*ingress.SSLPassthroughBackend

	hosts := sets.New[string]()

	for _, server := range servers {
		// If a location is defined by a prefix string that ends with the slash character, and requests are processed by one of
		// proxy_pass, fastcgi_pass, uwsgi_pass, scgi_pass, memcached_pass, or grpc_pass, then the special processing is performed.
		// In response to a request with URI equal to // this string, but without the trailing slash, a permanent redirect with the
		// code 301 will be returned to the requested URI with the slash appended. If this is not desired, an exact match of the
		// URIand location could be defined like this:
		//
		// location /user/ {
		//     proxy_pass http://user.example.com;
		// }
		// location = /user {
		//     proxy_pass http://login.example.com;
		// }
		server.Locations = updateServerLocations(server.Locations)

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
		DefaultSSLCertificate: n.getDefaultSSLCertificate(),
		StreamSnippets:        n.getStreamSnippets(ingresses),
	}
}

func dropSnippetDirectives(anns *annotations.Ingress, ingKey string) {
	if anns != nil {
		if anns.ConfigurationSnippet != "" {
			klog.V(3).Infof("Ingress %q tried to use configuration-snippet and the annotation is disabled by the admin. Removing the annotation", ingKey)
			anns.ConfigurationSnippet = ""
		}
		if anns.ServerSnippet != "" {
			klog.V(3).Infof("Ingress %q tried to use server-snippet and the annotation is disabled by the admin. Removing the annotation", ingKey)
			anns.ServerSnippet = ""
		}

		if anns.ModSecurity.Snippet != "" {
			klog.V(3).Infof("Ingress %q tried to use modsecurity-snippet and the annotation is disabled by the admin. Removing the annotation", ingKey)
			anns.ModSecurity.Snippet = ""
		}

		if anns.ExternalAuth.AuthSnippet != "" {
			klog.V(3).Infof("Ingress %q tried to use auth-snippet and the annotation is disabled by the admin. Removing the annotation", ingKey)
			anns.ExternalAuth.AuthSnippet = ""
		}

		if anns.StreamSnippet != "" {
			klog.V(3).Infof("Ingress %q tried to use stream-snippet and the annotation is disabled by the admin. Removing the annotation", ingKey)
			anns.StreamSnippet = ""
		}
	}
}

// getBackendServers returns a list of Upstream and Server to be used by the
// backend.  An upstream can be used in multiple servers if the namespace,
// service name and port are the same.
//
//nolint:gocyclo // Ignore function complexity error
func (n *NGINXController) getBackendServers(ingresses []*ingress.Ingress) ([]*ingress.Backend, []*ingress.Server) {
	du := n.getDefaultUpstream()
	upstreams := n.createUpstreams(ingresses, du)
	servers := n.createServers(ingresses, upstreams, du)

	var canaryIngresses []*ingress.Ingress

	for _, ing := range ingresses {
		ingKey := k8s.MetaNamespaceKey(ing)
		anns := ing.ParsedAnnotations

		if !n.store.GetBackendConfiguration().AllowSnippetAnnotations {
			dropSnippetDirectives(anns, ingKey)
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
				klog.V(3).Infof("Ingress %q does not contain any HTTP rule, using default backend", ingKey)
				continue
			}

			if server.AuthTLSError == "" && anns.CertificateAuth.AuthTLSError != "" {
				server.AuthTLSError = anns.CertificateAuth.AuthTLSError
			}

			if server.CertificateAuth.CAFileName == "" {
				server.CertificateAuth = anns.CertificateAuth
				if server.CertificateAuth.Secret != "" && server.CertificateAuth.CAFileName == "" {
					klog.V(3).Infof("Secret %q has no 'ca.crt' key, mutual authentication disabled for Ingress %q",
						server.CertificateAuth.Secret, ingKey)
				}
			} else {
				klog.V(3).Infof("Server %q is already configured for mutual authentication (Ingress %q)",
					server.Hostname, ingKey)
			}

			if !n.store.GetBackendConfiguration().ProxySSLLocationOnly {
				if server.ProxySSL.CAFileName == "" {
					server.ProxySSL = anns.ProxySSL
					if server.ProxySSL.Secret != "" && server.ProxySSL.CAFileName == "" {
						klog.V(3).Infof("Secret %q has no 'ca.crt' key, client cert authentication disabled for Ingress %q",
							server.ProxySSL.Secret, ingKey)
					}
				} else {
					klog.V(3).Infof("Server %q is already configured for client cert authentication (Ingress %q)",
						server.Hostname, ingKey)
				}
			}

			if rule.HTTP == nil {
				klog.V(3).Infof("Ingress %q does not contain any HTTP rule, using default backend", ingKey)
				continue
			}

			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service == nil {
					// skip non-service backends
					klog.V(3).Infof("Ingress %q and path %q does not contain a service backend, using default backend", ingKey, path.Path)
					continue
				}

				upsName := upstreamName(ing.Namespace, path.Backend.Service)

				ups := upstreams[upsName]

				// Backend is not referenced to by a server
				if ups.NoServer {
					continue
				}

				nginxPath := rootLocation
				if path.Path != "" {
					nginxPath = path.Path
				}

				addLoc := true
				for _, loc := range server.Locations {
					if loc.Path != nginxPath {
						continue
					}

					// Same paths but different types are allowed
					// (same type means overlap in the path definition)
					if !apiequality.Semantic.DeepEqual(loc.PathType, path.PathType) {
						break
					}

					addLoc = false

					if !loc.IsDefBackend {
						klog.V(3).Infof("Location %q already configured for server %q with upstream %q (Ingress %q)",
							loc.Path, server.Hostname, loc.Backend, ingKey)
						break
					}

					klog.V(3).Infof("Replacing location %q for server %q with upstream %q to use upstream %q (Ingress %q)",
						loc.Path, server.Hostname, loc.Backend, ups.Name, ingKey)

					loc.Backend = ups.Name
					loc.IsDefBackend = false
					loc.Port = ups.Port
					loc.Service = ups.Service
					loc.Ingress = ing

					locationApplyAnnotations(loc, anns)

					if loc.Redirect.FromToWWW {
						server.RedirectFromToWWW = true
					}

					break
				}

				// new location
				if addLoc {
					klog.V(3).Infof("Adding location %q for server %q with upstream %q (Ingress %q)",
						nginxPath, server.Hostname, ups.Name, ingKey)
					loc := &ingress.Location{
						Path:         nginxPath,
						PathType:     path.PathType,
						Backend:      ups.Name,
						IsDefBackend: false,
						Service:      ups.Service,
						Port:         ups.Port,
						Ingress:      ing,
					}
					locationApplyAnnotations(loc, anns)

					if loc.Redirect.FromToWWW {
						server.RedirectFromToWWW = true
					}
					server.Locations = append(server.Locations, loc)
				}

				if ups.SessionAffinity.AffinityType == "" {
					ups.SessionAffinity.AffinityType = anns.SessionAffinity.Type
				}

				if ups.SessionAffinity.AffinityMode == "" {
					ups.SessionAffinity.AffinityMode = anns.SessionAffinity.Mode
				}

				if anns.SessionAffinity.Type == "cookie" {
					cookiePath := anns.SessionAffinity.Cookie.Path
					if anns.Rewrite.UseRegex && cookiePath == "" {
						klog.Warningf("session-cookie-path should be set when use-regex is true")
					}

					ups.SessionAffinity.CookieSessionAffinity.Name = anns.SessionAffinity.Cookie.Name
					ups.SessionAffinity.CookieSessionAffinity.Expires = anns.SessionAffinity.Cookie.Expires
					ups.SessionAffinity.CookieSessionAffinity.MaxAge = anns.SessionAffinity.Cookie.MaxAge
					ups.SessionAffinity.CookieSessionAffinity.Secure = anns.SessionAffinity.Cookie.Secure
					ups.SessionAffinity.CookieSessionAffinity.Path = cookiePath
					ups.SessionAffinity.CookieSessionAffinity.Domain = anns.SessionAffinity.Cookie.Domain
					ups.SessionAffinity.CookieSessionAffinity.SameSite = anns.SessionAffinity.Cookie.SameSite
					ups.SessionAffinity.CookieSessionAffinity.ConditionalSameSiteNone = anns.SessionAffinity.Cookie.ConditionalSameSiteNone
					ups.SessionAffinity.CookieSessionAffinity.ChangeOnFailure = anns.SessionAffinity.Cookie.ChangeOnFailure

					locs := ups.SessionAffinity.CookieSessionAffinity.Locations
					if _, ok := locs[host]; !ok {
						locs[host] = []string{}
					}
					locs[host] = append(locs[host], path.Path)

					if len(server.Aliases) > 0 {
						for _, alias := range server.Aliases {
							if _, ok := locs[alias]; !ok {
								locs[alias] = []string{}
							}
							locs[alias] = append(locs[alias], path.Path)
						}
					}
				}
			}
		}

		// set aside canary ingresses to merge later
		if anns.Canary.Enabled {
			canaryIngresses = append(canaryIngresses, ing)
		}
	}

	if nonCanaryIngressExists(ingresses, canaryIngresses) {
		for _, canaryIng := range canaryIngresses {
			mergeAlternativeBackends(canaryIng, upstreams, servers)
		}
	}

	aUpstreams := make([]*ingress.Backend, 0, len(upstreams))

	for _, upstream := range upstreams {
		aUpstreams = append(aUpstreams, upstream)

		if upstream.Name == defUpstreamName {
			continue
		}

		isHTTPSfrom := []*ingress.Server{}
		for _, server := range servers {
			for _, location := range server.Locations {
				// use default backend
				if !shouldCreateUpstreamForLocationDefaultBackend(upstream, location) {
					continue
				}

				if len(location.DefaultBackend.Spec.Ports) == 0 {
					klog.Errorf("Custom default backend service %v/%v has no ports. Ignoring", location.DefaultBackend.Namespace, location.DefaultBackend.Name)
					continue
				}

				sp := location.DefaultBackend.Spec.Ports[0]
				var zone string
				if n.cfg.EnableTopologyAwareRouting {
					zone = getIngressPodZone(location.DefaultBackend)
				} else {
					zone = emptyZone
				}
				endps := getEndpointsFromSlices(location.DefaultBackend, &sp, apiv1.ProtocolTCP, zone, n.store.GetServiceEndpointsSlices)
				// custom backend is valid only if contains at least one endpoint
				if len(endps) > 0 {
					name := fmt.Sprintf("custom-default-backend-%v-%v", location.DefaultBackend.GetNamespace(), location.DefaultBackend.GetName())
					klog.V(3).Infof("Creating \"%v\" upstream based on default backend annotation", name)

					nb := upstream.DeepCopy()
					nb.Name = name
					nb.Endpoints = endps
					aUpstreams = append(aUpstreams, nb)
					location.DefaultBackendUpstreamName = name

					if len(upstream.Endpoints) == 0 {
						klog.V(3).Infof("Upstream %q has no active Endpoint, so using custom default backend for location %q in server %q (Service \"%v/%v\")",
							upstream.Name, location.Path, server.Hostname, location.DefaultBackend.Namespace, location.DefaultBackend.Name)

						location.Backend = name
					}
				}

				if server.SSLPassthrough {
					if location.Path == rootLocation {
						if location.Backend == defUpstreamName {
							klog.Warningf("Server %q has no default backend, ignoring SSL Passthrough.", server.Hostname)
							continue
						}
						isHTTPSfrom = append(isHTTPSfrom, server)
					}
				}
			}
		}

		if len(isHTTPSfrom) > 0 {
			upstream.SSLPassthrough = true
		}
	}

	aServers := make([]*ingress.Server, 0, len(servers))
	for _, value := range servers {
		sort.SliceStable(value.Locations, func(i, j int) bool {
			return value.Locations[i].Path > value.Locations[j].Path
		})

		sort.SliceStable(value.Locations, func(i, j int) bool {
			return len(value.Locations[i].Path) > len(value.Locations[j].Path)
		})
		aServers = append(aServers, value)
	}

	sort.SliceStable(aUpstreams, func(a, b int) bool {
		return aUpstreams[a].Name < aUpstreams[b].Name
	})

	sort.SliceStable(aServers, func(i, j int) bool {
		return aServers[i].Hostname < aServers[j].Hostname
	})

	return aUpstreams, aServers
}

// createUpstreams creates the NGINX upstreams (Endpoints) for each Service
// referenced in Ingress rules.
func (n *NGINXController) createUpstreams(data []*ingress.Ingress, du *ingress.Backend) map[string]*ingress.Backend {
	upstreams := make(map[string]*ingress.Backend)
	upstreams[defUpstreamName] = du

	for _, ing := range data {
		ingKey := k8s.MetaNamespaceKey(ing)
		anns := ing.ParsedAnnotations

		if !n.store.GetBackendConfiguration().AllowSnippetAnnotations {
			dropSnippetDirectives(anns, ingKey)
		}

		var defBackend string
		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
			defBackend = upstreamName(ing.Namespace, ing.Spec.DefaultBackend.Service)

			klog.V(3).Infof("Creating upstream %q", defBackend)
			upstreams[defBackend] = newUpstream(defBackend)

			upstreams[defBackend].UpstreamHashBy.UpstreamHashBy = anns.UpstreamHashBy.UpstreamHashBy
			upstreams[defBackend].UpstreamHashBy.UpstreamHashBySubset = anns.UpstreamHashBy.UpstreamHashBySubset
			upstreams[defBackend].UpstreamHashBy.UpstreamHashBySubsetSize = anns.UpstreamHashBy.UpstreamHashBySubsetSize

			upstreams[defBackend].LoadBalancing = anns.LoadBalancing
			if upstreams[defBackend].LoadBalancing == "" {
				upstreams[defBackend].LoadBalancing = n.store.GetBackendConfiguration().LoadBalancing
			}

			svcKey := fmt.Sprintf("%v/%v", ing.Namespace, ing.Spec.DefaultBackend.Service.Name)

			// add the service ClusterIP as a single Endpoint instead of individual Endpoints
			if anns.ServiceUpstream {
				endpoint, err := n.getServiceClusterEndpoint(svcKey, ing.Spec.DefaultBackend)
				if err != nil {
					klog.Errorf("Failed to determine a suitable ClusterIP Endpoint for Service %q: %v", svcKey, err)
				} else {
					upstreams[defBackend].Endpoints = []ingress.Endpoint{endpoint}
				}
			}

			// configure traffic shaping for canary
			if anns.Canary.Enabled {
				upstreams[defBackend].NoServer = true
				upstreams[defBackend].TrafficShapingPolicy = newTrafficShapingPolicy(&anns.Canary)
			}

			if len(upstreams[defBackend].Endpoints) == 0 {
				_, port := upstreamServiceNameAndPort(ing.Spec.DefaultBackend.Service)
				endps, err := n.serviceEndpoints(svcKey, port.String())
				upstreams[defBackend].Endpoints = append(upstreams[defBackend].Endpoints, endps...)
				if err != nil {
					klog.Warningf("Error creating upstream %q: %v", defBackend, err)
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

			for i, path := range rule.HTTP.Paths {
				if path.Backend.Service == nil {
					// skip non-service backends
					klog.V(3).Infof("Ingress %q and path %q does not contain a service backend, using default backend", ingKey, path.Path)
					continue
				}

				name := upstreamName(ing.Namespace, path.Backend.Service)
				svcName, svcPort := upstreamServiceNameAndPort(path.Backend.Service)
				if _, ok := upstreams[name]; ok {
					continue
				}

				klog.V(3).Infof("Creating upstream %q", name)
				upstreams[name] = newUpstream(name)
				upstreams[name].Port = svcPort

				upstreams[name].UpstreamHashBy.UpstreamHashBy = anns.UpstreamHashBy.UpstreamHashBy
				upstreams[name].UpstreamHashBy.UpstreamHashBySubset = anns.UpstreamHashBy.UpstreamHashBySubset
				upstreams[name].UpstreamHashBy.UpstreamHashBySubsetSize = anns.UpstreamHashBy.UpstreamHashBySubsetSize

				upstreams[name].LoadBalancing = anns.LoadBalancing
				if upstreams[name].LoadBalancing == "" {
					upstreams[name].LoadBalancing = n.store.GetBackendConfiguration().LoadBalancing
				}

				svcKey := fmt.Sprintf("%v/%v", ing.Namespace, svcName)

				// add the service ClusterIP as a single Endpoint instead of individual Endpoints
				if anns.ServiceUpstream {
					endpoint, err := n.getServiceClusterEndpoint(svcKey, &rule.HTTP.Paths[i].Backend)
					if err != nil {
						klog.Errorf("Failed to determine a suitable ClusterIP Endpoint for Service %q: %v", svcKey, err)
					} else {
						upstreams[name].Endpoints = []ingress.Endpoint{endpoint}
					}
				}

				// configure traffic shaping for canary
				if anns.Canary.Enabled {
					upstreams[name].NoServer = true
					upstreams[name].TrafficShapingPolicy = newTrafficShapingPolicy(&anns.Canary)
				}

				if len(upstreams[name].Endpoints) == 0 {
					_, port := upstreamServiceNameAndPort(path.Backend.Service)
					endp, err := n.serviceEndpoints(svcKey, port.String())
					if err != nil {
						klog.Warningf("Error obtaining Endpoints for Service %q: %v", svcKey, err)
						n.metricCollector.IncOrphanIngress(ing.Namespace, ing.Name, orphanMetricLabelNoService)
						continue
					}
					n.metricCollector.DecOrphanIngress(ing.Namespace, ing.Name, orphanMetricLabelNoService)

					if len(endp) == 0 {
						n.metricCollector.IncOrphanIngress(ing.Namespace, ing.Name, orphanMetricLabelNoEndpoint)
					} else {
						n.metricCollector.DecOrphanIngress(ing.Namespace, ing.Name, orphanMetricLabelNoEndpoint)
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
func (n *NGINXController) getServiceClusterEndpoint(svcKey string, backend *networking.IngressBackend) (endpoint ingress.Endpoint, err error) {
	svc, err := n.store.GetService(svcKey)
	if err != nil {
		return endpoint, fmt.Errorf("service %q does not exist", svcKey)
	}

	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		return endpoint, fmt.Errorf("no ClusterIP found for Service %q", svcKey)
	}

	endpoint.Address = svc.Spec.ClusterIP

	// if the Service port is referenced by name in the Ingress, lookup the
	// actual port in the service spec
	if backend.Service != nil {
		_, svcportintorstr := upstreamServiceNameAndPort(backend.Service)
		if svcportintorstr.Type == intstr.String {
			var port int32 = -1
			for _, svcPort := range svc.Spec.Ports {
				if svcPort.Name == svcportintorstr.String() {
					port = svcPort.Port
					break
				}
			}
			if port == -1 {
				return endpoint, fmt.Errorf("service %q does not have a port named %q", svc.Name, svcportintorstr.String())
			}
			endpoint.Port = fmt.Sprintf("%d", port)
		} else {
			endpoint.Port = svcportintorstr.String()
		}
	}

	return endpoint, err
}

// serviceEndpoints returns the upstream servers (Endpoints) associated with a Service.
func (n *NGINXController) serviceEndpoints(svcKey, backendPort string) ([]ingress.Endpoint, error) {
	var upstreams []ingress.Endpoint

	svc, err := n.store.GetService(svcKey)
	if err != nil {
		return upstreams, err
	}
	var zone string
	if n.cfg.EnableTopologyAwareRouting {
		zone = getIngressPodZone(svc)
	} else {
		zone = emptyZone
	}
	klog.V(3).Infof("Obtaining ports information for Service %q", svcKey)
	// Ingress with an ExternalName Service and no port defined for that Service
	if svc.Spec.Type == apiv1.ServiceTypeExternalName {
		if n.cfg.DisableServiceExternalName {
			klog.Warningf("Service %q of type ExternalName not allowed due to Ingress configuration.", svcKey)
			return upstreams, nil
		}
		servicePort := externalNamePorts(backendPort, svc)
		endps := getEndpointsFromSlices(svc, servicePort, apiv1.ProtocolTCP, zone, n.store.GetServiceEndpointsSlices)
		if len(endps) == 0 {
			klog.Warningf("Service %q does not have any active Endpoint.", svcKey)
			return upstreams, nil
		}

		upstreams = append(upstreams, endps...)
		return upstreams, nil
	}

	for i := range svc.Spec.Ports {
		servicePort := svc.Spec.Ports[i]
		// targetPort could be a string, use either the port name or number (int)
		if strconv.Itoa(int(servicePort.Port)) == backendPort ||
			servicePort.TargetPort.String() == backendPort ||
			servicePort.Name == backendPort {
			endps := getEndpointsFromSlices(svc, &servicePort, apiv1.ProtocolTCP, zone, n.store.GetServiceEndpointsSlices)
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

// createServers builds a map of host name to Server structs from a map of
// already computed Upstream structs. Each Server is configured with at least
// one root location, which uses a default backend if left unspecified.
func (n *NGINXController) createServers(data []*ingress.Ingress,
	upstreams map[string]*ingress.Backend,
	du *ingress.Backend,
) map[string]*ingress.Server {
	servers := make(map[string]*ingress.Server, len(data))
	allAliases := make(map[string][]string, len(data))

	bdef := n.store.GetDefaultBackend()
	ngxProxy := proxy.Config{
		BodySize:             bdef.ProxyBodySize,
		ConnectTimeout:       bdef.ProxyConnectTimeout,
		SendTimeout:          bdef.ProxySendTimeout,
		ReadTimeout:          bdef.ProxyReadTimeout,
		BuffersNumber:        bdef.ProxyBuffersNumber,
		BufferSize:           bdef.ProxyBufferSize,
		BusyBuffersSize:      bdef.ProxyBusyBuffersSize,
		CookieDomain:         bdef.ProxyCookieDomain,
		CookiePath:           bdef.ProxyCookiePath,
		NextUpstream:         bdef.ProxyNextUpstream,
		NextUpstreamTimeout:  bdef.ProxyNextUpstreamTimeout,
		NextUpstreamTries:    bdef.ProxyNextUpstreamTries,
		RequestBuffering:     bdef.ProxyRequestBuffering,
		ProxyRedirectFrom:    bdef.ProxyRedirectFrom,
		ProxyBuffering:       bdef.ProxyBuffering,
		ProxyHTTPVersion:     bdef.ProxyHTTPVersion,
		ProxyMaxTempFileSize: bdef.ProxyMaxTempFileSize,
	}

	// initialize default server and root location
	pathTypePrefix := networking.PathTypePrefix
	servers[defServerName] = &ingress.Server{
		Hostname: defServerName,
		SSLCert:  n.getDefaultSSLCertificate(),
		Locations: []*ingress.Location{
			{
				Path:         rootLocation,
				PathType:     &pathTypePrefix,
				IsDefBackend: true,
				Backend:      du.Name,
				Proxy:        ngxProxy,
				Service:      du.Service,
				Logs: log.Config{
					Access:  n.store.GetBackendConfiguration().EnableAccessLogForDefaultBackend,
					Rewrite: false,
				},
			},
		},
	}

	// initialize all other servers
	for _, ing := range data {
		ingKey := k8s.MetaNamespaceKey(ing)
		anns := ing.ParsedAnnotations

		if !n.store.GetBackendConfiguration().AllowSnippetAnnotations {
			dropSnippetDirectives(anns, ingKey)
		}

		// default upstream name
		un := du.Name

		if anns.Canary.Enabled {
			klog.V(2).Infof("Ingress %v is marked as Canary, ignoring", ingKey)
			continue
		}

		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
			defUpstream := upstreamName(ing.Namespace, ing.Spec.DefaultBackend.Service)

			if backendUpstream, ok := upstreams[defUpstream]; ok {
				// use backend specified in Ingress as the default backend for all its rules
				un = backendUpstream.Name

				defLoc := servers[defServerName].Locations[0]
				if defLoc.IsDefBackend && len(ing.Spec.Rules) == 0 {
					klog.V(2).Infof("Ingress %q defines a backend but no rule. Using it to configure the catch-all server %q", ingKey, defServerName)

					defLoc.IsDefBackend = false
					// special "catch all" case, Ingress with a backend but no rule
					defLoc.Backend = backendUpstream.Name
					defLoc.Service = backendUpstream.Service
					defLoc.Ingress = ing
					// TODO: Redirect and rewrite can affect the catch all behavior, skip for now
					originalRedirect := defLoc.Redirect
					originalRewrite := defLoc.Rewrite
					locationApplyAnnotations(defLoc, anns)
					defLoc.Redirect = originalRedirect
					defLoc.Rewrite = originalRewrite
				} else {
					klog.V(3).Infof("Ingress %q defines both a backend and rules. Using its backend as default upstream for all its rules.", ingKey)
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

			loc := &ingress.Location{
				Path:         rootLocation,
				PathType:     &pathTypePrefix,
				IsDefBackend: true,
				Backend:      un,
				Ingress:      ing,
				Service:      &apiv1.Service{},
			}
			locationApplyAnnotations(loc, anns)

			servers[host] = &ingress.Server{
				Hostname: host,
				Locations: []*ingress.Location{
					loc,
				},
				SSLPassthrough:         anns.SSLPassthrough,
				SSLCiphers:             anns.SSLCipher.SSLCiphers,
				SSLPreferServerCiphers: anns.SSLCipher.SSLPreferServerCiphers,
			}
		}
	}

	// configure default location, alias, and SSL
	for _, ing := range data {
		ingKey := k8s.MetaNamespaceKey(ing)
		anns := ing.ParsedAnnotations

		if !n.store.GetBackendConfiguration().AllowSnippetAnnotations {
			dropSnippetDirectives(anns, ingKey)
		}

		if anns.Canary.Enabled {
			klog.V(2).Infof("Ingress %v is marked as Canary, ignoring", ingKey)
			continue
		}

		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = defServerName
			}

			if len(servers[host].Aliases) == 0 {
				servers[host].Aliases = anns.Aliases
				if aliases := allAliases[host]; len(aliases) == 0 {
					allAliases[host] = anns.Aliases
				}
			} else {
				klog.Warningf("Aliases already configured for server %q, skipping (Ingress %q)", host, ingKey)
			}

			if anns.ServerSnippet != "" {
				if servers[host].ServerSnippet == "" {
					servers[host].ServerSnippet = anns.ServerSnippet
				} else {
					klog.Warningf("Server snippet already configured for server %q, skipping (Ingress %q)",
						host, ingKey)
				}
			}

			// only add SSL ciphers if the server does not have them previously configured
			if servers[host].SSLCiphers == "" && anns.SSLCipher.SSLCiphers != "" {
				servers[host].SSLCiphers = anns.SSLCipher.SSLCiphers
			}

			// only add SSLPreferServerCiphers if the server does not have them previously configured
			if servers[host].SSLPreferServerCiphers == "" && anns.SSLCipher.SSLPreferServerCiphers != "" {
				servers[host].SSLPreferServerCiphers = anns.SSLCipher.SSLPreferServerCiphers
			}

			// only add a certificate if the server does not have one previously configured
			if servers[host].SSLCert != nil {
				continue
			}

			if len(ing.Spec.TLS) == 0 {
				klog.V(3).Infof("Ingress %q does not contains a TLS section.", ingKey)
				continue
			}

			tlsSecretName := extractTLSSecretName(host, ing, n.store.GetLocalSSLCert)
			if tlsSecretName == "" {
				klog.V(3).Infof("Host %q is listed in the TLS section but secretName is empty. Using default certificate", host)
				servers[host].SSLCert = n.getDefaultSSLCertificate()
				continue
			}

			secrKey := fmt.Sprintf("%v/%v", ing.Namespace, tlsSecretName)
			cert, err := n.store.GetLocalSSLCert(secrKey)
			if err != nil {
				klog.Warningf("Error getting SSL certificate %q: %v. Using default certificate", secrKey, err)
				servers[host].SSLCert = n.getDefaultSSLCertificate()
				continue
			}

			if cert.Certificate == nil {
				klog.Warningf("SSL certificate %q does not contain a valid SSL certificate for server %q", secrKey, host)
				klog.Warningf("Using default certificate")
				servers[host].SSLCert = n.getDefaultSSLCertificate()
				continue
			}

			err = cert.Certificate.VerifyHostname(host)
			if err != nil {
				klog.Warningf("Unexpected error validating SSL certificate %q for server %q: %v", secrKey, host, err)
				klog.Warning("Validating certificate against DNS names. This will be deprecated in a future version")
				// check the Common Name field
				// https://github.com/golang/go/issues/22922
				err := verifyHostname(host, cert.Certificate)
				if err != nil {
					klog.Warningf("SSL certificate %q does not contain a Common Name or Subject Alternative Name for server %q: %v", secrKey, host, err)
					klog.Warningf("Using default certificate")
					servers[host].SSLCert = n.getDefaultSSLCertificate()
					continue
				}
			}

			servers[host].SSLCert = cert

			now := time.Now()
			if cert.ExpireTime.Before(now) {
				klog.Warningf("SSL certificate for server %q expired (%v)", host, cert.ExpireTime)
			} else if cert.ExpireTime.Before(now.Add(240 * time.Hour)) {
				klog.Warningf("SSL certificate for server %q is about to expire (%v)", host, cert.ExpireTime)
			}
		}
	}

	for host, hostAliases := range allAliases {
		if _, ok := servers[host]; !ok {
			continue
		}

		uniqAliases := sets.NewString()
		for _, alias := range hostAliases {
			if alias == host {
				continue
			}

			if _, ok := servers[alias]; ok {
				continue
			}

			if uniqAliases.Has(alias) {
				continue
			}

			uniqAliases.Insert(alias)
		}

		servers[host].Aliases = uniqAliases.List()
	}

	return servers
}

func locationApplyAnnotations(loc *ingress.Location, anns *annotations.Ingress) {
	loc.BasicDigestAuth = anns.BasicDigestAuth
	loc.ClientBodyBufferSize = anns.ClientBodyBufferSize
	loc.CustomHeaders = anns.CustomHeaders
	loc.ConfigurationSnippet = anns.ConfigurationSnippet
	loc.CorsConfig = anns.CorsConfig
	loc.ExternalAuth = anns.ExternalAuth
	loc.EnableGlobalAuth = anns.EnableGlobalAuth
	loc.HTTP2PushPreload = anns.HTTP2PushPreload
	loc.Opentelemetry = anns.Opentelemetry
	loc.Proxy = anns.Proxy
	loc.ProxySSL = anns.ProxySSL
	loc.RateLimit = anns.RateLimit
	loc.Redirect = anns.Redirect
	loc.Rewrite = anns.Rewrite
	loc.UpstreamVhost = anns.UpstreamVhost
	loc.Denylist = anns.Denylist
	loc.Allowlist = anns.Allowlist
	loc.Denied = anns.Denied
	loc.XForwardedPrefix = anns.XForwardedPrefix
	loc.UsePortInRedirects = anns.UsePortInRedirects
	loc.Connection = anns.Connection
	loc.Logs = anns.Logs
	loc.DefaultBackend = anns.DefaultBackend
	loc.BackendProtocol = anns.BackendProtocol
	loc.FastCGI = anns.FastCGI
	loc.CustomHTTPErrors = anns.CustomHTTPErrors
	loc.DisableProxyInterceptErrors = anns.DisableProxyInterceptErrors
	loc.ModSecurity = anns.ModSecurity
	loc.Satisfy = anns.Satisfy
	loc.Mirror = anns.Mirror

	loc.DefaultBackendUpstreamName = defUpstreamName
}

// OK to merge canary ingresses iff there exists one or more ingresses to potentially merge into
func nonCanaryIngressExists(ingresses, canaryIngresses []*ingress.Ingress) bool {
	return len(ingresses)-len(canaryIngresses) > 0
}

// ensure that the following conditions are met
// 1) names of backends do not match and canary doesn't merge into itself
// 2) primary name is not the default upstream
// 3) the primary has a server
func canMergeBackend(primary, alternative *ingress.Backend) bool {
	return alternative != nil && primary.Name != alternative.Name && primary.Name != defUpstreamName && !primary.NoServer
}

// Performs the merge action and checks to ensure that one two alternative backends do not merge into each other
func mergeAlternativeBackend(ing *ingress.Ingress, priUps, altUps *ingress.Backend) bool {
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

	if ing.ParsedAnnotations != nil && ing.ParsedAnnotations.SessionAffinity.CanaryBehavior != "legacy" {
		priUps.SessionAffinity.DeepCopyInto(&altUps.SessionAffinity)
	}

	priUps.AlternativeBackends = append(priUps.AlternativeBackends, altUps.Name)

	return true
}

// Compares an Ingress of a potential alternative backend's rules with each existing server and finds matching host + path pairs.
// If a match is found, we know that this server should back the alternative backend and add the alternative backend
// to a backend's alternative list.
// If no match is found, then the serverless backend is deleted.
func mergeAlternativeBackends(ing *ingress.Ingress, upstreams map[string]*ingress.Backend,
	servers map[string]*ingress.Server,
) {
	// merge catch-all alternative backends
	if ing.Spec.DefaultBackend != nil {
		upsName := upstreamName(ing.Namespace, ing.Spec.DefaultBackend.Service)

		altUps := upstreams[upsName]

		if altUps == nil {
			klog.Warningf("alternative backend %s has already been removed", upsName)
		} else {
			merged := false
			altEqualsPri := false

			for _, loc := range servers[defServerName].Locations {
				priUps, ok := upstreams[loc.Backend]
				if !ok {
					klog.Warningf("cannot find primary backend %s for location %s%s", loc.Backend, servers[defServerName].Hostname, loc.Path)
					continue
				}
				altEqualsPri = altUps.Name == priUps.Name
				if altEqualsPri {
					klog.Warningf("alternative upstream %s in Ingress %s/%s is primary upstream in Other Ingress for location %s%s!",
						altUps.Name, ing.Namespace, ing.Name, servers[defServerName].Hostname, loc.Path)
					break
				}

				if canMergeBackend(priUps, altUps) {
					klog.V(2).Infof("matching backend %v found for alternative backend %v",
						priUps.Name, altUps.Name)

					merged = mergeAlternativeBackend(ing, priUps, altUps)
				}
			}

			if !altEqualsPri && !merged {
				klog.Warningf("unable to find real backend for alternative backend %v. Deleting.", altUps.Name)
				delete(upstreams, altUps.Name)
			}
		}
	}

	for _, rule := range ing.Spec.Rules {
		host := rule.Host
		if host == "" {
			host = defServerName
		}

		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service == nil {
				// skip non-service backends
				klog.V(3).Infof("Ingress %q and path %q does not contain a service backend, using default backend", k8s.MetaNamespaceKey(ing), path.Path)
				continue
			}

			upsName := upstreamName(ing.Namespace, path.Backend.Service)

			altUps := upstreams[upsName]

			if altUps == nil {
				klog.Warningf("alternative backend %s has already been removed", upsName)
				continue
			}

			merged := false
			altEqualsPri := false

			server, ok := servers[host]
			if !ok {
				klog.Errorf("cannot merge alternative backend %s into hostname %s that does not exist",
					altUps.Name,
					host)

				continue
			}

			// find matching paths
			for _, loc := range server.Locations {
				priUps, ok := upstreams[loc.Backend]
				if !ok {
					klog.Warningf("cannot find primary backend %s for location %s%s", loc.Backend, server.Hostname, loc.Path)
					continue
				}
				altEqualsPri = altUps.Name == priUps.Name
				if altEqualsPri {
					klog.Warningf("alternative upstream %s in Ingress %s/%s is primary upstream in Other Ingress for location %s%s!",
						altUps.Name, ing.Namespace, ing.Name, server.Hostname, loc.Path)
					break
				}

				if canMergeBackend(priUps, altUps) && loc.Path == path.Path && *loc.PathType == *path.PathType {
					klog.V(2).Infof("matching backend %v found for alternative backend %v",
						priUps.Name, altUps.Name)

					merged = mergeAlternativeBackend(ing, priUps, altUps)
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
	getLocalSSLCert func(string) (*ingress.SSLCert, error),
) string {
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

		if cert == nil || cert.Certificate == nil {
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

// checks conditions for whether or not an upstream should be created for a custom default backend
func shouldCreateUpstreamForLocationDefaultBackend(upstream *ingress.Backend, location *ingress.Location) bool {
	return (upstream.Name == location.Backend) &&
		(len(upstream.Endpoints) == 0 || len(location.CustomHTTPErrors) != 0) &&
		location.DefaultBackend != nil
}

func externalNamePorts(name string, svc *apiv1.Service) *apiv1.ServicePort {
	port, err := strconv.Atoi(name) // #nosec
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
		//nolint:gosec // Ignore G109 error
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
		Protocol: "TCP",
		//nolint:gosec // Ignore G109 error
		Port:       int32(port),
		TargetPort: intstr.FromInt(port),
	}
}

func checkOverlap(ing *networking.Ingress, servers []*ingress.Server) error {
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		if rule.Host == "" {
			rule.Host = defServerName
		}

		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service == nil {
				// skip non-service backends
				klog.V(3).Infof("Ingress %q and path %q does not contain a service backend, using default backend", k8s.MetaNamespaceKey(ing), path.Path)
				continue
			}

			if path.Path == "" {
				path.Path = rootLocation
			}

			existingIngresses := ingressForHostPath(rule.Host, path.Path, servers)

			// no previous ingress
			if len(existingIngresses) == 0 {
				continue
			}

			// same ingress
			for _, existing := range existingIngresses {
				if existing.ObjectMeta.Namespace == ing.ObjectMeta.Namespace && existing.ObjectMeta.Name == ing.ObjectMeta.Name {
					return nil
				}
			}

			// path overlap. Check if one of the ingresses has a canary annotation
			isCanaryEnabled, annotationErr := parser.GetBoolAnnotation("canary", ing, canary.CanaryAnnotations.Annotations)
			for _, existing := range existingIngresses {
				isExistingCanaryEnabled, existingAnnotationErr := parser.GetBoolAnnotation("canary", existing, canary.CanaryAnnotations.Annotations)

				if isCanaryEnabled && isExistingCanaryEnabled {
					return fmt.Errorf(`host "%s" and path "%s" is already defined in ingress %s/%s`, rule.Host, path.Path, existing.Namespace, existing.Name)
				}

				if annotationErr == errors.ErrMissingAnnotations && existingAnnotationErr == errors.ErrMissingAnnotations {
					return fmt.Errorf(`host "%s" and path "%s" is already defined in ingress %s/%s`, rule.Host, path.Path, existing.Namespace, existing.Name)
				}
			}

			// no overlap
			return nil
		}
	}

	return nil
}

func ingressForHostPath(hostname, path string, servers []*ingress.Server) []*networking.Ingress {
	ingresses := make([]*networking.Ingress, 0)

	for _, server := range servers {
		if hostname != server.Hostname {
			continue
		}

		for i, location := range server.Locations {
			if location.Path != path {
				continue
			}

			if location.IsDefBackend {
				continue
			}

			ingresses = append(ingresses, &server.Locations[i].Ingress.Ingress)
		}
	}

	return ingresses
}

func (n *NGINXController) getStreamSnippets(ingresses []*ingress.Ingress) []string {
	snippets := make([]string, 0, len(ingresses))
	for _, i := range ingresses {
		if i.ParsedAnnotations.StreamSnippet == "" {
			continue
		}
		snippets = append(snippets, i.ParsedAnnotations.StreamSnippet)
	}
	return snippets
}

// newTrafficShapingPolicy creates new ingress.TrafficShapingPolicy instance using canary configuration
func newTrafficShapingPolicy(cfg *canary.Config) ingress.TrafficShapingPolicy {
	return ingress.TrafficShapingPolicy{
		Weight:        cfg.Weight,
		WeightTotal:   cfg.WeightTotal,
		Header:        cfg.Header,
		HeaderValue:   cfg.HeaderValue,
		HeaderPattern: cfg.HeaderPattern,
		Cookie:        cfg.Cookie,
	}
}
