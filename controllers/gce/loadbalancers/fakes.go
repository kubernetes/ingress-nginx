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
	"fmt"
	"testing"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress/controllers/gce/utils"
)

var testIPManager = testIP{}

type testIP struct {
	start int
}

func (t *testIP) ip() string {
	t.start++
	return fmt.Sprintf("0.0.0.%v", t.start)
}

// Loadbalancer fakes

// FakeLoadBalancers is a type that fakes out the loadbalancer interface.
type FakeLoadBalancers struct {
	Fw    []*compute.ForwardingRule
	Um    []*compute.UrlMap
	Tp    []*compute.TargetHttpProxy
	Tps   []*compute.TargetHttpsProxy
	IP    []*compute.Address
	Certs []*compute.SslCertificate
	name  string
	calls []string // list of calls that were made
}

// TODO: There is some duplication between these functions and the name mungers in
// loadbalancer file.
func (f *FakeLoadBalancers) fwName(https bool) string {
	if https {
		return fmt.Sprintf("%v-%v", httpsForwardingRulePrefix, f.name)
	}
	return fmt.Sprintf("%v-%v", forwardingRulePrefix, f.name)
}

func (f *FakeLoadBalancers) umName() string {
	return fmt.Sprintf("%v-%v", urlMapPrefix, f.name)
}

func (f *FakeLoadBalancers) tpName(https bool) string {
	if https {
		return fmt.Sprintf("%v-%v", targetHTTPSProxyPrefix, f.name)
	}
	return fmt.Sprintf("%v-%v", targetProxyPrefix, f.name)
}

// String is the string method for FakeLoadBalancers.
func (f *FakeLoadBalancers) String() string {
	msg := fmt.Sprintf(
		"Loadbalancer %v,\nforwarding rules:\n", f.name)
	for _, fw := range f.Fw {
		msg += fmt.Sprintf("\t%v\n", fw.Name)
	}
	msg += fmt.Sprintf("Target proxies\n")
	for _, tp := range f.Tp {
		msg += fmt.Sprintf("\t%v\n", tp.Name)
	}
	msg += fmt.Sprintf("UrlMaps\n")
	for _, um := range f.Um {
		msg += fmt.Sprintf("%v\n", um.Name)
		msg += fmt.Sprintf("\tHost Rules:\n")
		for _, hostRule := range um.HostRules {
			msg += fmt.Sprintf("\t\t%v\n", hostRule)
		}
		msg += fmt.Sprintf("\tPath Matcher:\n")
		for _, pathMatcher := range um.PathMatchers {
			msg += fmt.Sprintf("\t\t%v\n", pathMatcher.Name)
			for _, pathRule := range pathMatcher.PathRules {
				msg += fmt.Sprintf("\t\t\t%+v\n", pathRule)
			}
		}
	}
	return msg
}

// Forwarding Rule fakes

// GetGlobalForwardingRule returns a fake forwarding rule.
func (f *FakeLoadBalancers) GetGlobalForwardingRule(name string) (*compute.ForwardingRule, error) {
	f.calls = append(f.calls, "GetGlobalForwardingRule")
	for i := range f.Fw {
		if f.Fw[i].Name == name {
			return f.Fw[i], nil
		}
	}
	return nil, fmt.Errorf("forwarding rule %v not found", name)
}

// CreateGlobalForwardingRule fakes forwarding rule creation.
func (f *FakeLoadBalancers) CreateGlobalForwardingRule(rule *compute.ForwardingRule) error {
	f.calls = append(f.calls, "CreateGlobalForwardingRule")
	if rule.IPAddress == "" {
		rule.IPAddress = fmt.Sprintf(testIPManager.ip())
	}
	rule.SelfLink = rule.Name
	f.Fw = append(f.Fw, rule)
	return nil
}

// SetProxyForGlobalForwardingRule fakes setting a global forwarding rule.
func (f *FakeLoadBalancers) SetProxyForGlobalForwardingRule(forwardingRuleName, proxyLink string) error {
	f.calls = append(f.calls, "SetProxyForGlobalForwardingRule")
	for i := range f.Fw {
		if f.Fw[i].Name == forwardingRuleName {
			f.Fw[i].Target = proxyLink
		}
	}
	return nil
}

// DeleteGlobalForwardingRule fakes deleting a global forwarding rule.
func (f *FakeLoadBalancers) DeleteGlobalForwardingRule(name string) error {
	f.calls = append(f.calls, "DeleteGlobalForwardingRule")
	fw := []*compute.ForwardingRule{}
	for i := range f.Fw {
		if f.Fw[i].Name != name {
			fw = append(fw, f.Fw[i])
		}
	}
	f.Fw = fw
	return nil
}

// GetForwardingRulesWithIPs returns all forwarding rules that match the given ips.
func (f *FakeLoadBalancers) GetForwardingRulesWithIPs(ip []string) (fwRules []*compute.ForwardingRule) {
	f.calls = append(f.calls, "GetForwardingRulesWithIPs")
	ipSet := sets.NewString(ip...)
	for i := range f.Fw {
		if ipSet.Has(f.Fw[i].IPAddress) {
			fwRules = append(fwRules, f.Fw[i])
		}
	}
	return fwRules
}

// UrlMaps fakes

// GetUrlMap fakes getting url maps from the cloud.
func (f *FakeLoadBalancers) GetUrlMap(name string) (*compute.UrlMap, error) {
	f.calls = append(f.calls, "GetUrlMap")
	for i := range f.Um {
		if f.Um[i].Name == name {
			return f.Um[i], nil
		}
	}
	return nil, fmt.Errorf("url map %v not found", name)
}

// CreateUrlMap fakes url-map creation.
func (f *FakeLoadBalancers) CreateUrlMap(urlMap *compute.UrlMap) error {
	f.calls = append(f.calls, "CreateUrlMap")
	urlMap.SelfLink = f.umName()
	f.Um = append(f.Um, urlMap)
	return nil
}

// UpdateUrlMap fakes updating url-maps.
func (f *FakeLoadBalancers) UpdateUrlMap(urlMap *compute.UrlMap) error {
	f.calls = append(f.calls, "UpdateUrlMap")
	for i := range f.Um {
		if f.Um[i].Name == urlMap.Name {
			f.Um[i] = urlMap
			return nil
		}
	}
	return fmt.Errorf("url map %v not found", urlMap.Name)
}

// DeleteUrlMap fakes url-map deletion.
func (f *FakeLoadBalancers) DeleteUrlMap(name string) error {
	f.calls = append(f.calls, "DeleteUrlMap")
	um := []*compute.UrlMap{}
	for i := range f.Um {
		if f.Um[i].Name != name {
			um = append(um, f.Um[i])
		}
	}
	f.Um = um
	return nil
}

// TargetProxies fakes

// GetTargetHttpProxy fakes getting target http proxies from the cloud.
func (f *FakeLoadBalancers) GetTargetHttpProxy(name string) (*compute.TargetHttpProxy, error) {
	f.calls = append(f.calls, "GetTargetHttpProxy")
	for i := range f.Tp {
		if f.Tp[i].Name == name {
			return f.Tp[i], nil
		}
	}
	return nil, fmt.Errorf("target http proxy %v not found", name)
}

// CreateTargetHttpProxy fakes creating a target http proxy.
func (f *FakeLoadBalancers) CreateTargetHttpProxy(proxy *compute.TargetHttpProxy) error {
	f.calls = append(f.calls, "CreateTargetHttpProxy")
	proxy.SelfLink = proxy.Name
	f.Tp = append(f.Tp, proxy)
	return nil
}

// DeleteTargetHttpProxy fakes deleting a target http proxy.
func (f *FakeLoadBalancers) DeleteTargetHttpProxy(name string) error {
	f.calls = append(f.calls, "DeleteTargetHttpProxy")
	tp := []*compute.TargetHttpProxy{}
	for i := range f.Tp {
		if f.Tp[i].Name != name {
			tp = append(tp, f.Tp[i])
		}
	}
	f.Tp = tp
	return nil
}

// SetUrlMapForTargetHttpProxy fakes setting an url-map for a target http proxy.
func (f *FakeLoadBalancers) SetUrlMapForTargetHttpProxy(proxy *compute.TargetHttpProxy, urlMap *compute.UrlMap) error {
	f.calls = append(f.calls, "SetUrlMapForTargetHttpProxy")
	for i := range f.Tp {
		if f.Tp[i].Name == proxy.Name {
			f.Tp[i].UrlMap = urlMap.SelfLink
		}
	}
	return nil
}

// TargetHttpsProxy fakes

// GetTargetHttpsProxy fakes getting target http proxies from the cloud.
func (f *FakeLoadBalancers) GetTargetHttpsProxy(name string) (*compute.TargetHttpsProxy, error) {
	f.calls = append(f.calls, "GetTargetHttpsProxy")
	for i := range f.Tps {
		if f.Tps[i].Name == name {
			return f.Tps[i], nil
		}
	}
	return nil, fmt.Errorf("target https proxy %v not found", name)
}

// CreateTargetHttpsProxy fakes creating a target http proxy.
func (f *FakeLoadBalancers) CreateTargetHttpsProxy(proxy *compute.TargetHttpsProxy) error {
	f.calls = append(f.calls, "CreateTargetHttpsProxy")
	proxy.SelfLink = proxy.Name
	f.Tps = append(f.Tps, proxy)
	return nil
}

// DeleteTargetHttpsProxy fakes deleting a target http proxy.
func (f *FakeLoadBalancers) DeleteTargetHttpsProxy(name string) error {
	f.calls = append(f.calls, "DeleteTargetHttpsProxy")
	tp := []*compute.TargetHttpsProxy{}
	for i := range f.Tps {
		if f.Tps[i].Name != name {
			tp = append(tp, f.Tps[i])
		}
	}
	f.Tps = tp
	return nil
}

// SetUrlMapForTargetHttpsProxy fakes setting an url-map for a target http proxy.
func (f *FakeLoadBalancers) SetUrlMapForTargetHttpsProxy(proxy *compute.TargetHttpsProxy, urlMap *compute.UrlMap) error {
	f.calls = append(f.calls, "SetUrlMapForTargetHttpsProxy")
	for i := range f.Tps {
		if f.Tps[i].Name == proxy.Name {
			f.Tps[i].UrlMap = urlMap.SelfLink
		}
	}
	return nil
}

// SetSslCertificateForTargetHttpsProxy fakes out setting certificates.
func (f *FakeLoadBalancers) SetSslCertificateForTargetHttpsProxy(proxy *compute.TargetHttpsProxy, SSLCert *compute.SslCertificate) error {
	f.calls = append(f.calls, "SetSslCertificateForTargetHttpsProxy")
	found := false
	for i := range f.Tps {
		if f.Tps[i].Name == proxy.Name {
			f.Tps[i].SslCertificates = []string{SSLCert.SelfLink}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("failed to find proxy %v", proxy.Name)
	}
	return nil
}

// UrlMap fakes

// CheckURLMap checks the URL map.
func (f *FakeLoadBalancers) CheckURLMap(t *testing.T, l7 *L7, expectedMap map[string]utils.FakeIngressRuleValueMap) {
	f.calls = append(f.calls, "CheckURLMap")
	um, err := f.GetUrlMap(l7.um.Name)
	if err != nil || um == nil {
		t.Fatalf("%v", err)
	}
	// Check the default backend
	var d string
	if h, ok := expectedMap[utils.DefaultBackendKey]; ok {
		if d, ok = h[utils.DefaultBackendKey]; ok {
			delete(h, utils.DefaultBackendKey)
		}
		delete(expectedMap, utils.DefaultBackendKey)
	}
	// The urlmap should have a default backend, and each path matcher.
	if d != "" && l7.um.DefaultService != d {
		t.Fatalf("Expected default backend %v found %v",
			d, l7.um.DefaultService)
	}

	for _, matcher := range l7.um.PathMatchers {
		var hostname string
		// There's a 1:1 mapping between pathmatchers and hosts
		for _, hostRule := range l7.um.HostRules {
			if matcher.Name == hostRule.PathMatcher {
				if len(hostRule.Hosts) != 1 {
					t.Fatalf("Unexpected hosts in hostrules %+v", hostRule)
				}
				if d != "" && matcher.DefaultService != d {
					t.Fatalf("Expected default backend %v found %v",
						d, matcher.DefaultService)
				}
				hostname = hostRule.Hosts[0]
				break
			}
		}
		// These are all pathrules for a single host, found above
		for _, rule := range matcher.PathRules {
			if len(rule.Paths) != 1 {
				t.Fatalf("Unexpected rule in pathrules %+v", rule)
			}
			pathRule := rule.Paths[0]
			if hostMap, ok := expectedMap[hostname]; !ok {
				t.Fatalf("Expected map for host %v: %v", hostname, hostMap)
			} else if svc, ok := expectedMap[hostname][pathRule]; !ok {
				t.Fatalf("Expected rule %v in host map", pathRule)
			} else if svc != rule.Service {
				t.Fatalf("Expected service %v found %v", svc, rule.Service)
			}
			delete(expectedMap[hostname], pathRule)
			if len(expectedMap[hostname]) == 0 {
				delete(expectedMap, hostname)
			}
		}
	}
	if len(expectedMap) != 0 {
		t.Fatalf("Untranslated entries %+v", expectedMap)
	}
}

// Static IP fakes

// ReserveGlobalAddress fakes out static IP reservation.
func (f *FakeLoadBalancers) ReserveGlobalAddress(addr *compute.Address) error {
	f.calls = append(f.calls, "ReserveGlobalAddress")
	f.IP = append(f.IP, addr)
	return nil
}

// GetGlobalAddress fakes out static IP retrieval.
func (f *FakeLoadBalancers) GetGlobalAddress(name string) (*compute.Address, error) {
	f.calls = append(f.calls, "GetGlobalAddress")
	for i := range f.IP {
		if f.IP[i].Name == name {
			return f.IP[i], nil
		}
	}
	return nil, fmt.Errorf("static IP %v not found", name)
}

// DeleteGlobalAddress fakes out static IP deletion.
func (f *FakeLoadBalancers) DeleteGlobalAddress(name string) error {
	f.calls = append(f.calls, "DeleteGlobalAddress")
	ip := []*compute.Address{}
	for i := range f.IP {
		if f.IP[i].Name != name {
			ip = append(ip, f.IP[i])
		}
	}
	f.IP = ip
	return nil
}

// SslCertificate fakes

// GetSslCertificate fakes out getting ssl certs.
func (f *FakeLoadBalancers) GetSslCertificate(name string) (*compute.SslCertificate, error) {
	f.calls = append(f.calls, "GetSslCertificate")
	for i := range f.Certs {
		if f.Certs[i].Name == name {
			return f.Certs[i], nil
		}
	}
	return nil, fmt.Errorf("cert %v not found", name)
}

// CreateSslCertificate fakes out certificate creation.
func (f *FakeLoadBalancers) CreateSslCertificate(cert *compute.SslCertificate) (*compute.SslCertificate, error) {
	f.calls = append(f.calls, "CreateSslCertificate")
	cert.SelfLink = cert.Name
	f.Certs = append(f.Certs, cert)
	return cert, nil
}

// DeleteSslCertificate fakes out certificate deletion.
func (f *FakeLoadBalancers) DeleteSslCertificate(name string) error {
	f.calls = append(f.calls, "DeleteSslCertificate")
	certs := []*compute.SslCertificate{}
	for i := range f.Certs {
		if f.Certs[i].Name != name {
			certs = append(certs, f.Certs[i])
		}
	}
	f.Certs = certs
	return nil
}

// NewFakeLoadBalancers creates a fake cloud client. Name is the name
// inserted into the selfLink of the associated resources for testing.
// eg: forwardingRule.SelfLink == k8-fw-name.
func NewFakeLoadBalancers(name string) *FakeLoadBalancers {
	return &FakeLoadBalancers{
		Fw:   []*compute.ForwardingRule{},
		name: name,
	}
}
