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

package loadbalancers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/glog"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress/controllers/gce/backends"
	"k8s.io/ingress/controllers/gce/storage"
	"k8s.io/ingress/controllers/gce/utils"
)

const (

	// The gce api uses the name of a path rule to match a host rule.
	hostRulePrefix = "host"

	// DefaultHost is the host used if none is specified. It is a valid value
	// for the "Host" field recognized by GCE.
	DefaultHost = "*"

	// DefaultPath is the path used if none is specified. It is a valid path
	// recognized by GCE.
	DefaultPath = "/*"

	// A single target proxy/urlmap/forwarding rule is created per loadbalancer.
	// Tagged with the namespace/name of the Ingress.
	// TODO: Move the namer to its own package out of utils and move the prefix
	// with it. Currently the construction of the loadbalancer resources names
	// are split between the namer and the loadbalancers package.
	targetProxyPrefix         = "k8s-tp"
	targetHTTPSProxyPrefix    = "k8s-tps"
	sslCertPrefix             = "k8s-ssl"
	forwardingRulePrefix      = "k8s-fw"
	httpsForwardingRulePrefix = "k8s-fws"
	urlMapPrefix              = "k8s-um"
	httpDefaultPortRange      = "80-80"
	httpsDefaultPortRange     = "443-443"
)

// L7s implements LoadBalancerPool.
type L7s struct {
	cloud       LoadBalancers
	snapshotter storage.Snapshotter
	// TODO: Remove this field and always ask the BackendPool using the NodePort.
	glbcDefaultBackend     *compute.BackendService
	defaultBackendPool     backends.BackendPool
	defaultBackendNodePort backends.ServicePort
	namer                  *utils.Namer
}

// NewLoadBalancerPool returns a new loadbalancer pool.
// - cloud: implements LoadBalancers. Used to sync L7 loadbalancer resources
//	 with the cloud.
// - defaultBackendPool: a BackendPool used to manage the GCE BackendService for
//   the default backend.
// - defaultBackendNodePort: The nodePort of the Kubernetes service representing
//   the default backend.
func NewLoadBalancerPool(
	cloud LoadBalancers,
	defaultBackendPool backends.BackendPool,
	defaultBackendNodePort backends.ServicePort, namer *utils.Namer) LoadBalancerPool {
	return &L7s{cloud, storage.NewInMemoryPool(), nil, defaultBackendPool, defaultBackendNodePort, namer}
}

func (l *L7s) create(ri *L7RuntimeInfo) (*L7, error) {
	if l.glbcDefaultBackend == nil {
		glog.Warningf("Creating l7 without a default backend")
	}
	return &L7{
		runtimeInfo:        ri,
		Name:               l.namer.LBName(ri.Name),
		cloud:              l.cloud,
		glbcDefaultBackend: l.glbcDefaultBackend,
		namer:              l.namer,
		sslCert:            nil,
	}, nil
}

// Get returns the loadbalancer by name.
func (l *L7s) Get(name string) (*L7, error) {
	name = l.namer.LBName(name)
	lb, exists := l.snapshotter.Get(name)
	if !exists {
		return nil, fmt.Errorf("loadbalancer %v not in pool", name)
	}
	return lb.(*L7), nil
}

// Add gets or creates a loadbalancer.
// If the loadbalancer already exists, it checks that its edges are valid.
func (l *L7s) Add(ri *L7RuntimeInfo) (err error) {
	name := l.namer.LBName(ri.Name)

	lb, _ := l.Get(name)
	if lb == nil {
		glog.Infof("Creating l7 %v", name)
		lb, err = l.create(ri)
		if err != nil {
			return err
		}
	} else {
		if !reflect.DeepEqual(lb.runtimeInfo, ri) {
			glog.Infof("LB %v runtime info changed, old %+v new %+v", lb.Name, lb.runtimeInfo, ri)
			lb.runtimeInfo = ri
		}
	}
	// Add the lb to the pool, in case we create an UrlMap but run out
	// of quota in creating the ForwardingRule we still need to cleanup
	// the UrlMap during GC.
	defer l.snapshotter.Add(name, lb)

	// Why edge hop for the create?
	// The loadbalancer is a fictitious resource, it doesn't exist in gce. To
	// make it exist we need to create a collection of gce resources, done
	// through the edge hop.
	if err := lb.edgeHop(); err != nil {
		return err
	}

	return nil
}

// Delete deletes a loadbalancer by name.
func (l *L7s) Delete(name string) error {
	name = l.namer.LBName(name)
	lb, err := l.Get(name)
	if err != nil {
		return err
	}
	glog.Infof("Deleting lb %v", name)
	if err := lb.Cleanup(); err != nil {
		return err
	}
	l.snapshotter.Delete(name)
	return nil
}

// Sync loadbalancers with the given runtime info from the controller.
func (l *L7s) Sync(lbs []*L7RuntimeInfo) error {
	glog.V(3).Infof("Syncing loadbalancers %v", lbs)

	if len(lbs) != 0 {
		// Lazily create a default backend so we don't tax users who don't care
		// about Ingress by consuming 1 of their 3 GCE BackendServices. This
		// BackendService is GC'd when there are no more Ingresses.
		if err := l.defaultBackendPool.Add(l.defaultBackendNodePort); err != nil {
			return err
		}
		defaultBackend, err := l.defaultBackendPool.Get(l.defaultBackendNodePort.Port)
		if err != nil {
			return err
		}
		l.glbcDefaultBackend = defaultBackend
	}
	// create new loadbalancers, validate existing
	for _, ri := range lbs {
		if err := l.Add(ri); err != nil {
			return err
		}
	}
	return nil
}

// GC garbage collects loadbalancers not in the input list.
func (l *L7s) GC(names []string) error {
	knownLoadBalancers := sets.NewString()
	for _, n := range names {
		knownLoadBalancers.Insert(l.namer.LBName(n))
	}
	pool := l.snapshotter.Snapshot()

	// Delete unknown loadbalancers
	for name := range pool {
		if knownLoadBalancers.Has(name) {
			continue
		}
		glog.V(3).Infof("GCing loadbalancer %v", name)
		if err := l.Delete(name); err != nil {
			return err
		}
	}
	// Tear down the default backend when there are no more loadbalancers.
	// This needs to happen after we've deleted all url-maps that might be
	// using it.
	if len(names) == 0 {
		if err := l.defaultBackendPool.Delete(l.defaultBackendNodePort.Port); err != nil {
			return err
		}
		l.glbcDefaultBackend = nil
	}
	return nil
}

// Shutdown logs whether or not the pool is empty.
func (l *L7s) Shutdown() error {
	if err := l.GC([]string{}); err != nil {
		return err
	}
	if err := l.defaultBackendPool.Shutdown(); err != nil {
		return err
	}
	glog.Infof("Loadbalancer pool shutdown.")
	return nil
}

// TLSCerts encapsulates .pem encoded TLS information.
type TLSCerts struct {
	// Key is private key.
	Key string
	// Cert is a public key.
	Cert string
	// Chain is a certificate chain.
	Chain string
}

// L7RuntimeInfo is info passed to this module from the controller runtime.
type L7RuntimeInfo struct {
	// Name is the name of a loadbalancer.
	Name string
	// IP is the desired ip of the loadbalancer, eg from a staticIP.
	IP string
	// TLS are the tls certs to use in termination.
	TLS *TLSCerts
	// TLSName is the name of/for the tls cert to use.
	TLSName string
	// AllowHTTP will not setup :80, if TLS is nil and AllowHTTP is set,
	// no loadbalancer is created.
	AllowHTTP bool
	// The name of a Global Static IP. If specified, the IP associated with
	// this name is used in the Forwarding Rules for this loadbalancer.
	StaticIPName string
}

// String returns the load balancer name
func (l *L7RuntimeInfo) String() string {
	return l.Name
}

// L7 represents a single L7 loadbalancer.
type L7 struct {
	Name string
	// runtimeInfo is non-cloudprovider information passed from the controller.
	runtimeInfo *L7RuntimeInfo
	// cloud is an interface to manage loadbalancers in the GCE cloud.
	cloud LoadBalancers
	// um is the UrlMap associated with this L7.
	um *compute.UrlMap
	// tp is the TargetHTTPProxy associated with this L7.
	tp *compute.TargetHttpProxy
	// tps is the TargetHTTPSProxy associated with this L7.
	tps *compute.TargetHttpsProxy
	// fw is the GlobalForwardingRule that points to the TargetHTTPProxy.
	fw *compute.ForwardingRule
	// fws is the GlobalForwardingRule that points to the TargetHTTPSProxy.
	fws *compute.ForwardingRule
	// ip is the static-ip associated with both GlobalForwardingRules.
	ip *compute.Address
	// sslCert is the ssl cert associated with the targetHTTPSProxy.
	// TODO: Make this a custom type that contains crt+key
	sslCert *compute.SslCertificate
	// oldSSLCert is the certificate that used to be hooked up to the
	// targetHTTPSProxy. We can't update a cert in place, so we need
	// to create - update - delete and storing the old cert in a field
	// prevents leakage if there's a failure along the way.
	oldSSLCert *compute.SslCertificate
	// glbcDefaultBacked is the backend to use if no path rules match.
	// TODO: Expose this to users.
	glbcDefaultBackend *compute.BackendService
	// namer is used to compute names of the various sub-components of an L7.
	namer *utils.Namer
}

func (l *L7) checkUrlMap(backend *compute.BackendService) (err error) {
	if l.glbcDefaultBackend == nil {
		return fmt.Errorf("cannot create urlmap without default backend")
	}
	urlMapName := l.namer.Truncate(fmt.Sprintf("%v-%v", urlMapPrefix, l.Name))
	urlMap, _ := l.cloud.GetUrlMap(urlMapName)
	if urlMap != nil {
		glog.V(3).Infof("Url map %v already exists", urlMap.Name)
		l.um = urlMap
		return nil
	}

	glog.Infof("Creating url map %v for backend %v", urlMapName, l.glbcDefaultBackend.Name)
	newUrlMap := &compute.UrlMap{
		Name:           urlMapName,
		DefaultService: l.glbcDefaultBackend.SelfLink,
	}
	if err = l.cloud.CreateUrlMap(newUrlMap); err != nil {
		return err
	}
	urlMap, err = l.cloud.GetUrlMap(urlMapName)
	if err != nil {
		return err
	}
	l.um = urlMap
	return nil
}

func (l *L7) checkProxy() (err error) {
	if l.um == nil {
		return fmt.Errorf("cannot create proxy without urlmap")
	}
	proxyName := l.namer.Truncate(fmt.Sprintf("%v-%v", targetProxyPrefix, l.Name))
	proxy, _ := l.cloud.GetTargetHttpProxy(proxyName)
	if proxy == nil {
		glog.Infof("Creating new http proxy for urlmap %v", l.um.Name)
		newProxy := &compute.TargetHttpProxy{
			Name:   proxyName,
			UrlMap: l.um.SelfLink,
		}
		if err = l.cloud.CreateTargetHttpProxy(newProxy); err != nil {
			return err
		}
		proxy, err = l.cloud.GetTargetHttpProxy(proxyName)
		if err != nil {
			return err
		}
		l.tp = proxy
		return nil
	}
	if !utils.CompareLinks(proxy.UrlMap, l.um.SelfLink) {
		glog.Infof("Proxy %v has the wrong url map, setting %v overwriting %v",
			proxy.Name, l.um, proxy.UrlMap)
		if err := l.cloud.SetUrlMapForTargetHttpProxy(proxy, l.um); err != nil {
			return err
		}
	}
	l.tp = proxy
	return nil
}

func (l *L7) deleteOldSSLCert() (err error) {
	if l.oldSSLCert == nil || l.sslCert == nil ||
		l.oldSSLCert.Name == l.sslCert.Name || !strings.HasPrefix(l.oldSSLCert.Name, sslCertPrefix) {
		return nil
	}
	glog.Infof("Cleaning up old SSL Certificate %v, current name %v", l.oldSSLCert.Name, l.sslCert.Name)
	if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteSslCertificate(l.oldSSLCert.Name)); err != nil {
		return err
	}
	l.oldSSLCert = nil
	return nil
}

// Returns the name portion of a link - which is the last section
func getResourceNameFromLink(link string) string {
	s := strings.Split(link, "/")
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}

func (l *L7) usePreSharedCert() (bool, error) {
	// Use the named GCE cert when it is specified by the annotation.
	preSharedCertName := l.runtimeInfo.TLSName
	if preSharedCertName == "" {
		return false, nil
	}

	// Ask GCE for the cert, checking for problems and existence.
	cert, err := l.cloud.GetSslCertificate(preSharedCertName)
	if err != nil {
		return true, err
	}
	if cert == nil {
		return true, fmt.Errorf("cannot find existing sslCertificate %v for %v", preSharedCertName, l.Name)
	}

	glog.V(2).Infof("Using existing sslCertificate %v for %v", preSharedCertName, l.Name)
	l.sslCert = cert
	return true, nil
}

func (l *L7) populateSSLCert() error {
	// Determine what certificate name is being used
	var expectedCertName string
	if l.sslCert != nil {
		expectedCertName = l.sslCert.Name
	} else {
		// Retrieve the ssl certificate in use by the expected target proxy (if exists)
		expectedCertName = getResourceNameFromLink(l.getSslCertLinkInUse())
	}

	var err error
	if expectedCertName != "" {
		// Retrieve the certificate and ignore error if certificate wasn't found
		l.sslCert, err = l.cloud.GetSslCertificate(expectedCertName)
		if err != nil {
			return utils.IgnoreHTTPNotFound(err)
		}
	}
	return nil
}

func (l *L7) nextCertificateName() string {
	// The name of the cert for this lb flip-flops between these 2 on
	// every certificate update. We don't append the index at the end so we're
	// sure it isn't truncated.
	// TODO: Clean this code up into a ring buffer.
	primaryCertName := l.namer.Truncate(fmt.Sprintf("%v-%v", sslCertPrefix, l.Name))
	secondaryCertName := l.namer.Truncate(fmt.Sprintf("%v-%d-%v", sslCertPrefix, 1, l.Name))

	if l.sslCert != nil && l.sslCert.Name == primaryCertName {
		return secondaryCertName
	}
	return primaryCertName
}

func (l *L7) checkSSLCert() error {
	// Handle Pre-Shared cert and early return if used
	if used, err := l.usePreSharedCert(); used {
		return err
	}

	// Get updated value of certificate for comparison
	if err := l.populateSSLCert(); err != nil {
		return err
	}

	// TODO: Currently, GCE only supports a single certificate per static IP
	// so we don't need to bother with disambiguation. Naming the cert after
	// the loadbalancer is a simplification.
	ingCert := l.runtimeInfo.TLS.Cert
	ingKey := l.runtimeInfo.TLS.Key

	// PrivateKey is write only, so compare certs alone. We're assuming that
	// no one will change just the key. We can remember the key and compare,
	// but a bug could end up leaking it, which feels worse.
	if l.sslCert != nil && ingCert == l.sslCert.Certificate {
		return nil
	}

	// Controller needs to create or update the certificate.
	// Generate the next certificate name to use.
	newCertName := l.nextCertificateName()

	// Perform a delete in case a certificate exists with the exact name
	// This certificate should be unused since we check the target proxy's certificate prior
	// to this point. Although, it's possible an actor pointed a target proxy to this certificate.
	if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteSslCertificate(newCertName)); err != nil {
		return fmt.Errorf("unable to delete ssl certificate with name %q, expected it to be unused. err: %v", newCertName, err)
	}

	glog.V(2).Infof("Creating new sslCertificate %v for %v", newCertName, l.Name)
	cert, err := l.cloud.CreateSslCertificate(&compute.SslCertificate{
		Name:        newCertName,
		Certificate: ingCert,
		PrivateKey:  ingKey,
	})
	if err != nil {
		return err
	}
	// Save the current cert for cleanup after we update the target proxy.
	l.oldSSLCert = l.sslCert
	l.sslCert = cert

	return nil
}

func (l *L7) getSslCertLinkInUse() string {
	proxyName := l.namer.Truncate(fmt.Sprintf("%v-%v", targetHTTPSProxyPrefix, l.Name))
	proxy, _ := l.cloud.GetTargetHttpsProxy(proxyName)
	if proxy != nil && len(proxy.SslCertificates) > 0 {
		return proxy.SslCertificates[0]
	}
	return ""
}

func (l *L7) checkHttpsProxy() (err error) {
	if l.sslCert == nil {
		glog.V(3).Infof("No SSL certificates for %v, will not create HTTPS proxy.", l.Name)
		return nil
	}
	if l.um == nil {
		return fmt.Errorf("no UrlMap for %v, will not create HTTPS proxy", l.Name)
	}
	proxyName := l.namer.Truncate(fmt.Sprintf("%v-%v", targetHTTPSProxyPrefix, l.Name))
	proxy, _ := l.cloud.GetTargetHttpsProxy(proxyName)
	if proxy == nil {
		glog.Infof("Creating new https proxy for urlmap %v", l.um.Name)
		newProxy := &compute.TargetHttpsProxy{
			Name:            proxyName,
			UrlMap:          l.um.SelfLink,
			SslCertificates: []string{l.sslCert.SelfLink},
		}
		if err = l.cloud.CreateTargetHttpsProxy(newProxy); err != nil {
			return err
		}

		proxy, err = l.cloud.GetTargetHttpsProxy(proxyName)
		if err != nil {
			return err
		}

		l.tps = proxy
		return nil
	}
	if !utils.CompareLinks(proxy.UrlMap, l.um.SelfLink) {
		glog.Infof("Https proxy %v has the wrong url map, setting %v overwriting %v",
			proxy.Name, l.um, proxy.UrlMap)
		if err := l.cloud.SetUrlMapForTargetHttpsProxy(proxy, l.um); err != nil {
			return err
		}
	}
	cert := proxy.SslCertificates[0]
	if !utils.CompareLinks(cert, l.sslCert.SelfLink) {
		glog.Infof("Https proxy %v has the wrong ssl certs, setting %v overwriting %v",
			proxy.Name, l.sslCert.SelfLink, cert)
		if err := l.cloud.SetSslCertificateForTargetHttpsProxy(proxy, l.sslCert); err != nil {
			return err
		}
	}
	glog.V(3).Infof("Created target https proxy %v", proxy.Name)
	l.tps = proxy
	return nil
}

func (l *L7) checkForwardingRule(name, proxyLink, ip, portRange string) (fw *compute.ForwardingRule, err error) {
	fw, _ = l.cloud.GetGlobalForwardingRule(name)
	if fw != nil && (ip != "" && fw.IPAddress != ip || fw.PortRange != portRange) {
		glog.Warningf("Recreating forwarding rule %v(%v), so it has %v(%v)",
			fw.IPAddress, fw.PortRange, ip, portRange)
		if err = utils.IgnoreHTTPNotFound(l.cloud.DeleteGlobalForwardingRule(name)); err != nil {
			return nil, err
		}
		fw = nil
	}
	if fw == nil {
		parts := strings.Split(proxyLink, "/")
		glog.Infof("Creating forwarding rule for proxy %v and ip %v:%v", parts[len(parts)-1:], ip, portRange)
		rule := &compute.ForwardingRule{
			Name:       name,
			IPAddress:  ip,
			Target:     proxyLink,
			PortRange:  portRange,
			IPProtocol: "TCP",
		}
		if err = l.cloud.CreateGlobalForwardingRule(rule); err != nil {
			return nil, err
		}
		fw, err = l.cloud.GetGlobalForwardingRule(name)
		if err != nil {
			return nil, err
		}
	}
	// TODO: If the port range and protocol don't match, recreate the rule
	if utils.CompareLinks(fw.Target, proxyLink) {
		glog.V(3).Infof("Forwarding rule %v already exists", fw.Name)
	} else {
		glog.Infof("Forwarding rule %v has the wrong proxy, setting %v overwriting %v",
			fw.Name, fw.Target, proxyLink)
		if err := l.cloud.SetProxyForGlobalForwardingRule(fw.Name, proxyLink); err != nil {
			return nil, err
		}
	}
	return fw, nil
}

// getEffectiveIP returns a string with the IP to use in the HTTP and HTTPS
// forwarding rules, and a boolean indicating if this is an IP the controller
// should manage or not.
func (l *L7) getEffectiveIP() (string, bool) {

	// A note on IP management:
	// User specifies a different IP on startup:
	//	- We create a forwarding rule with the given IP.
	//		- If this ip doesn't exist in GCE, we create another one in the hope
	//		  that they will rectify it later on.
	//	- In the happy case, no static ip is created or deleted by this controller.
	// Controller allocates a staticIP/ephemeralIP, but user changes it:
	//  - We still delete the old static IP, but only when we tear down the
	//	  Ingress in Cleanup(). Till then the static IP stays around, but
	//    the forwarding rules get deleted/created with the new IP.
	//  - There will be a period of downtime as we flip IPs.
	// User specifies the same static IP to 2 Ingresses:
	//  - GCE will throw a 400, and the controller will keep trying to use
	//    the IP in the hope that the user manually resolves the conflict
	//    or deletes/modifies the Ingress.
	// TODO: Handle the last case better.

	if l.runtimeInfo.StaticIPName != "" {
		// Existing static IPs allocated to forwarding rules will get orphaned
		// till the Ingress is torn down.
		if ip, err := l.cloud.GetGlobalAddress(l.runtimeInfo.StaticIPName); err != nil || ip == nil {
			glog.Warningf("The given static IP name %v doesn't translate to an existing global static IP, ignoring it and allocating a new IP: %v",
				l.runtimeInfo.StaticIPName, err)
		} else {
			return ip.Address, false
		}
	}
	if l.ip != nil {
		return l.ip.Address, true
	}
	return "", true
}

func (l *L7) checkHttpForwardingRule() (err error) {
	if l.tp == nil {
		return fmt.Errorf("cannot create forwarding rule without proxy")
	}
	name := l.namer.Truncate(fmt.Sprintf("%v-%v", forwardingRulePrefix, l.Name))
	address, _ := l.getEffectiveIP()
	fw, err := l.checkForwardingRule(name, l.tp.SelfLink, address, httpDefaultPortRange)
	if err != nil {
		return err
	}
	l.fw = fw
	return nil
}

func (l *L7) checkHttpsForwardingRule() (err error) {
	if l.tps == nil {
		glog.V(3).Infof("No https target proxy for %v, not created https forwarding rule", l.Name)
		return nil
	}
	name := l.namer.Truncate(fmt.Sprintf("%v-%v", httpsForwardingRulePrefix, l.Name))
	address, _ := l.getEffectiveIP()
	fws, err := l.checkForwardingRule(name, l.tps.SelfLink, address, httpsDefaultPortRange)
	if err != nil {
		return err
	}
	l.fws = fws
	return nil
}

// checkStaticIP reserves a static IP allocated to the Forwarding Rule.
func (l *L7) checkStaticIP() (err error) {
	if l.fw == nil || l.fw.IPAddress == "" {
		return fmt.Errorf("will not create static IP without a forwarding rule")
	}
	// Don't manage staticIPs if the user has specified an IP.
	if address, manageStaticIP := l.getEffectiveIP(); !manageStaticIP {
		glog.V(3).Infof("Not managing user specified static IP %v", address)
		return nil
	}
	staticIPName := l.namer.Truncate(fmt.Sprintf("%v-%v", forwardingRulePrefix, l.Name))
	ip, _ := l.cloud.GetGlobalAddress(staticIPName)
	if ip == nil {
		glog.Infof("Creating static ip %v", staticIPName)
		err = l.cloud.ReserveGlobalAddress(&compute.Address{Name: staticIPName, Address: l.fw.IPAddress})
		if err != nil {
			if utils.IsHTTPErrorCode(err, http.StatusConflict) ||
				utils.IsHTTPErrorCode(err, http.StatusBadRequest) {
				glog.V(3).Infof("IP %v(%v) is already reserved, assuming it is OK to use.",
					l.fw.IPAddress, staticIPName)
				return nil
			}
			return err
		}
		ip, err = l.cloud.GetGlobalAddress(staticIPName)
		if err != nil {
			return err
		}
	}
	l.ip = ip
	return nil
}

func (l *L7) edgeHop() error {
	if err := l.checkUrlMap(l.glbcDefaultBackend); err != nil {
		return err
	}
	if l.runtimeInfo.AllowHTTP {
		if err := l.edgeHopHttp(); err != nil {
			return err
		}
	}
	// Defer promoting an ephemeral to a static IP until it's really needed.
	if l.runtimeInfo.AllowHTTP && (l.runtimeInfo.TLS != nil || l.runtimeInfo.TLSName != "") {
		glog.V(3).Infof("checking static ip for %v", l.Name)
		if err := l.checkStaticIP(); err != nil {
			return err
		}
	}
	if l.runtimeInfo.TLS != nil || l.runtimeInfo.TLSName != "" {
		glog.V(3).Infof("validating https for %v", l.Name)
		if err := l.edgeHopHttps(); err != nil {
			return err
		}
	}
	return nil
}

func (l *L7) edgeHopHttp() error {
	if err := l.checkProxy(); err != nil {
		return err
	}
	if err := l.checkHttpForwardingRule(); err != nil {
		return err
	}
	return nil
}

func (l *L7) edgeHopHttps() error {
	if err := l.checkSSLCert(); err != nil {
		return err
	}
	if err := l.checkHttpsProxy(); err != nil {
		return err
	}
	if err := l.checkHttpsForwardingRule(); err != nil {
		return err
	}
	if err := l.deleteOldSSLCert(); err != nil {
		return err
	}
	return nil
}

// GetIP returns the ip associated with the forwarding rule for this l7.
func (l *L7) GetIP() string {
	if l.fw != nil {
		return l.fw.IPAddress
	}
	if l.fws != nil {
		return l.fws.IPAddress
	}
	return ""
}

// getNameForPathMatcher returns a name for a pathMatcher based on the given host rule.
// The host rule can be a regex, the path matcher name used to associate the 2 cannot.
func getNameForPathMatcher(hostRule string) string {
	hasher := md5.New()
	hasher.Write([]byte(hostRule))
	return fmt.Sprintf("%v%v", hostRulePrefix, hex.EncodeToString(hasher.Sum(nil)))
}

// UpdateUrlMap translates the given hostname: endpoint->port mapping into a gce url map.
//
// HostRule: Conceptually contains all PathRules for a given host.
// PathMatcher: Associates a path rule with a host rule. Mostly an optimization.
// PathRule: Maps a single path regex to a backend.
//
// The GCE url map allows multiple hosts to share url->backend mappings without duplication, eg:
//   Host: foo(PathMatcher1), bar(PathMatcher1,2)
//   PathMatcher1:
//     /a -> b1
//     /b -> b2
//   PathMatcher2:
//     /c -> b1
// This leads to a lot of complexity in the common case, where all we want is a mapping of
// host->{/path: backend}.
//
// Consider some alternatives:
// 1. Using a single backend per PathMatcher:
//   Host: foo(PathMatcher1,3) bar(PathMatcher1,2,3)
//   PathMatcher1:
//     /a -> b1
//   PathMatcher2:
//     /c -> b1
//   PathMatcher3:
//     /b -> b2
// 2. Using a single host per PathMatcher:
//   Host: foo(PathMatcher1)
//   PathMatcher1:
//     /a -> b1
//     /b -> b2
//   Host: bar(PathMatcher2)
//   PathMatcher2:
//     /a -> b1
//     /b -> b2
//     /c -> b1
// In the context of kubernetes services, 2 makes more sense, because we
// rarely want to lookup backends (service:nodeport). When a service is
// deleted, we need to find all host PathMatchers that have the backend
// and remove the mapping. When a new path is added to a host (happens
// more frequently than service deletion) we just need to lookup the 1
// pathmatcher of the host.
func (l *L7) UpdateUrlMap(ingressRules utils.GCEURLMap) error {
	if l.um == nil {
		return fmt.Errorf("cannot add url without an urlmap")
	}

	// All UrlMaps must have a default backend. If the Ingress has a default
	// backend, it applies to all host rules as well as to the urlmap itself.
	// If it doesn't the urlmap might have a stale default, so replace it with
	// glbc's default backend.
	defaultBackend := ingressRules.GetDefaultBackend()
	if defaultBackend != nil {
		l.um.DefaultService = defaultBackend.SelfLink
	} else {
		l.um.DefaultService = l.glbcDefaultBackend.SelfLink
	}

	// Every update replaces the entire urlmap.
	// TODO:  when we have multiple loadbalancers point to a single gce url map
	// this needs modification. For now, there is a 1:1 mapping of urlmaps to
	// Ingresses, so if the given Ingress doesn't have a host rule we should
	// delete the path to that backend.
	l.um.HostRules = []*compute.HostRule{}
	l.um.PathMatchers = []*compute.PathMatcher{}

	for hostname, urlToBackend := range ingressRules {
		// Create a host rule
		// Create a path matcher
		// Add all given endpoint:backends to pathRules in path matcher
		pmName := getNameForPathMatcher(hostname)
		l.um.HostRules = append(l.um.HostRules, &compute.HostRule{
			Hosts:       []string{hostname},
			PathMatcher: pmName,
		})

		pathMatcher := &compute.PathMatcher{
			Name:           pmName,
			DefaultService: l.um.DefaultService,
			PathRules:      []*compute.PathRule{},
		}

		// Longest prefix wins. For equal rules, first hit wins, i.e the second
		// /foo rule when the first is deleted.
		for expr, be := range urlToBackend {
			pathMatcher.PathRules = append(
				pathMatcher.PathRules, &compute.PathRule{Paths: []string{expr}, Service: be.SelfLink})
		}
		l.um.PathMatchers = append(l.um.PathMatchers, pathMatcher)
	}
	oldMap, _ := l.cloud.GetUrlMap(l.um.Name)
	if oldMap != nil && mapsEqual(oldMap, l.um) {
		glog.Infof("UrlMap for l7 %v is unchanged", l.Name)
		return nil
	}

	glog.V(3).Infof("Updating URLMap: %q", l.Name)
	if err := l.cloud.UpdateUrlMap(l.um); err != nil {
		return err
	}

	um, err := l.cloud.GetUrlMap(l.um.Name)
	if err != nil {
		return err
	}

	l.um = um
	return nil
}

func mapsEqual(a, b *compute.UrlMap) bool {
	if a.DefaultService != b.DefaultService {
		return false
	}
	if len(a.HostRules) != len(b.HostRules) {
		return false
	}
	for i := range a.HostRules {
		a := a.HostRules[i]
		b := b.HostRules[i]
		if a.Description != b.Description {
			return false
		}
		if len(a.Hosts) != len(b.Hosts) {
			return false
		}
		for i := range a.Hosts {
			if a.Hosts[i] != b.Hosts[i] {
				return false
			}
		}
		if a.PathMatcher != b.PathMatcher {
			return false
		}
	}
	if len(a.PathMatchers) != len(b.PathMatchers) {
		return false
	}
	for i := range a.PathMatchers {
		a := a.PathMatchers[i]
		b := b.PathMatchers[i]
		if a.DefaultService != b.DefaultService {
			return false
		}
		if a.Description != b.Description {
			return false
		}
		if a.Name != b.Name {
			return false
		}
		if len(a.PathRules) != len(b.PathRules) {
			return false
		}
		for i := range a.PathRules {
			a := a.PathRules[i]
			b := b.PathRules[i]
			if len(a.Paths) != len(b.Paths) {
				return false
			}
			for i := range a.Paths {
				if a.Paths[i] != b.Paths[i] {
					return false
				}
			}
			if a.Service != b.Service {
				return false
			}
		}
	}
	return true
}

// Cleanup deletes resources specific to this l7 in the right order.
// forwarding rule -> target proxy -> url map
// This leaves backends and health checks, which are shared across loadbalancers.
func (l *L7) Cleanup() error {
	if l.fw != nil {
		glog.V(2).Infof("Deleting global forwarding rule %v", l.fw.Name)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteGlobalForwardingRule(l.fw.Name)); err != nil {
			return err
		}
		l.fw = nil
	}
	if l.fws != nil {
		glog.V(2).Infof("Deleting global forwarding rule %v", l.fws.Name)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteGlobalForwardingRule(l.fws.Name)); err != nil {
			return err
		}
		l.fws = nil
	}
	if l.ip != nil {
		glog.V(2).Infof("Deleting static IP %v(%v)", l.ip.Name, l.ip.Address)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteGlobalAddress(l.ip.Name)); err != nil {
			return err
		}
		l.ip = nil
	}
	if l.tps != nil {
		glog.V(2).Infof("Deleting target https proxy %v", l.tps.Name)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteTargetHttpsProxy(l.tps.Name)); err != nil {
			return err
		}
		l.tps = nil
	}
	// Delete the SSL cert if it is from a secret, not referencing a pre-created GCE cert.
	if l.sslCert != nil && l.runtimeInfo.TLSName == "" {
		glog.V(2).Infof("Deleting sslcert %v", l.sslCert.Name)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteSslCertificate(l.sslCert.Name)); err != nil {
			return err
		}
		l.sslCert = nil
	}
	if l.tp != nil {
		glog.V(2).Infof("Deleting target http proxy %v", l.tp.Name)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteTargetHttpProxy(l.tp.Name)); err != nil {
			return err
		}
		l.tp = nil
	}
	if l.um != nil {
		glog.V(2).Infof("Deleting url map %v", l.um.Name)
		if err := utils.IgnoreHTTPNotFound(l.cloud.DeleteUrlMap(l.um.Name)); err != nil {
			return err
		}
		l.um = nil
	}
	return nil
}

// getBackendNames returns the names of backends in this L7 urlmap.
func (l *L7) getBackendNames() []string {
	if l.um == nil {
		return []string{}
	}
	beNames := sets.NewString()
	for _, pathMatcher := range l.um.PathMatchers {
		for _, pathRule := range pathMatcher.PathRules {
			// This is gross, but the urlmap only has links to backend services.
			parts := strings.Split(pathRule.Service, "/")
			name := parts[len(parts)-1]
			if name != "" {
				beNames.Insert(name)
			}
		}
	}
	// The default Service recorded in the urlMap is a link to the backend.
	// Note that this can either be user specified, or the L7 controller's
	// global default.
	parts := strings.Split(l.um.DefaultService, "/")
	defaultBackendName := parts[len(parts)-1]
	if defaultBackendName != "" {
		beNames.Insert(defaultBackendName)
	}
	return beNames.List()
}

// GetLBAnnotations returns the annotations of an l7. This includes it's current status.
func GetLBAnnotations(l7 *L7, existing map[string]string, backendPool backends.BackendPool) map[string]string {
	if existing == nil {
		existing = map[string]string{}
	}
	backends := l7.getBackendNames()
	backendState := map[string]string{}
	for _, beName := range backends {
		backendState[beName] = backendPool.Status(beName)
	}
	jsonBackendState := "Unknown"
	b, err := json.Marshal(backendState)
	if err == nil {
		jsonBackendState = string(b)
	}
	existing[fmt.Sprintf("%v/url-map", utils.K8sAnnotationPrefix)] = l7.um.Name
	// Forwarding rule and target proxy might not exist if allowHTTP == false
	if l7.fw != nil {
		existing[fmt.Sprintf("%v/forwarding-rule", utils.K8sAnnotationPrefix)] = l7.fw.Name
	}
	if l7.tp != nil {
		existing[fmt.Sprintf("%v/target-proxy", utils.K8sAnnotationPrefix)] = l7.tp.Name
	}
	// HTTPs resources might not exist if TLS == nil
	if l7.fws != nil {
		existing[fmt.Sprintf("%v/https-forwarding-rule", utils.K8sAnnotationPrefix)] = l7.fws.Name
	}
	if l7.tps != nil {
		existing[fmt.Sprintf("%v/https-target-proxy", utils.K8sAnnotationPrefix)] = l7.tps.Name
	}
	if l7.ip != nil {
		existing[fmt.Sprintf("%v/static-ip", utils.K8sAnnotationPrefix)] = l7.ip.Name
	}
	if l7.sslCert != nil {
		existing[fmt.Sprintf("%v/ssl-cert", utils.K8sAnnotationPrefix)] = l7.sslCert.Name
	}
	// TODO: We really want to know *when* a backend flipped states.
	existing[fmt.Sprintf("%v/backends", utils.K8sAnnotationPrefix)] = jsonBackendState
	return existing
}

// GCEResourceName retrieves the name of the gce resource created for this
// Ingress, of the given resource type, by inspecting the map of ingress
// annotations.
func GCEResourceName(ingAnnotations map[string]string, resourceName string) string {
	// Even though this function is trivial, it exists to keep the annotation
	// parsing logic in a single location.
	resourceName, _ = ingAnnotations[fmt.Sprintf("%v/%v", utils.K8sAnnotationPrefix, resourceName)]
	return resourceName
}
