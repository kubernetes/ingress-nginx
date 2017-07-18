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

	"k8s.io/ingress/controllers/gce/backends"
	"k8s.io/ingress/controllers/gce/healthchecks"
	"k8s.io/ingress/controllers/gce/instances"
	"k8s.io/ingress/controllers/gce/utils"
)

const (
	defaultZone = "zone-a"
)

var (
	testDefaultBeNodePort = backends.ServicePort{Port: 3000, Protocol: utils.ProtocolHTTP}
)

func newFakeLoadBalancerPool(f LoadBalancers, t *testing.T) LoadBalancerPool {
	fakeBackends := backends.NewFakeBackendServices(func(op int, be *compute.BackendService) error { return nil })
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	fakeHCP := healthchecks.NewFakeHealthCheckProvider()
	namer := &utils.Namer{}
	healthChecker := healthchecks.NewHealthChecker(fakeHCP, "/", namer)
	nodePool := instances.NewNodePool(fakeIGs)
	nodePool.Init(&instances.FakeZoneLister{Zones: []string{defaultZone}})
	backendPool := backends.NewBackendPool(
		fakeBackends, healthChecker, nodePool, namer, []int64{}, false)
	return NewLoadBalancerPool(f, backendPool, testDefaultBeNodePort, namer)
}

func TestCreateHTTPLoadBalancer(t *testing.T) {
	// This should NOT create the forwarding rule and target proxy
	// associated with the HTTPS branch of this loadbalancer.
	lbInfo := &L7RuntimeInfo{Name: "test", AllowHTTP: true}
	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil || l7 == nil {
		t.Fatalf("Expected l7 not created")
	}
	um, err := f.GetUrlMap(f.umName())
	if err != nil ||
		um.DefaultService != pool.(*L7s).glbcDefaultBackend.SelfLink {
		t.Fatalf("%v", err)
	}
	tp, err := f.GetTargetHttpProxy(f.tpName(false))
	if err != nil || tp.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	fw, err := f.GetGlobalForwardingRule(f.fwName(false))
	if err != nil || fw.Target != tp.SelfLink {
		t.Fatalf("%v", err)
	}
}

func TestCreateHTTPSLoadBalancer(t *testing.T) {
	// This should NOT create the forwarding rule and target proxy
	// associated with the HTTP branch of this loadbalancer.
	lbInfo := &L7RuntimeInfo{
		Name:      "test",
		AllowHTTP: false,
		TLS:       &TLSCerts{Key: "key", Cert: "cert"},
	}
	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil || l7 == nil {
		t.Fatalf("Expected l7 not created")
	}
	um, err := f.GetUrlMap(f.umName())
	if err != nil ||
		um.DefaultService != pool.(*L7s).glbcDefaultBackend.SelfLink {
		t.Fatalf("%v", err)
	}
	tps, err := f.GetTargetHttpsProxy(f.tpName(true))
	if err != nil || tps.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	fws, err := f.GetGlobalForwardingRule(f.fwName(true))
	if err != nil || fws.Target != tps.SelfLink {
		t.Fatalf("%v", err)
	}
}

// Tests that a certificate is created from the provided Key/Cert combo
// and the proxy is updated to another cert when the provided cert changes
func TestCertUpdate(t *testing.T) {
	primaryCertName := "k8s-ssl-test"
	secondaryCertName := "k8s-ssl-1-test"
	lbInfo := &L7RuntimeInfo{
		Name:      "test",
		AllowHTTP: false,
		TLS:       &TLSCerts{Key: "key", Cert: "cert"},
	}

	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)

	// Sync first cert
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(primaryCertName, lbInfo.TLS.Cert, f, t)

	// Sync with different cert
	lbInfo.TLS = &TLSCerts{Key: "key2", Cert: "cert2"}
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(secondaryCertName, lbInfo.TLS.Cert, f, t)
}

// Tests that controller can overwrite existing, unused certificates
func TestCertCreationWithCollision(t *testing.T) {
	primaryCertName := "k8s-ssl-test"
	secondaryCertName := "k8s-ssl-1-test"
	lbInfo := &L7RuntimeInfo{
		Name:      "test",
		AllowHTTP: false,
		TLS:       &TLSCerts{Key: "key", Cert: "cert"},
	}

	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)

	// Have both names already used by orphaned certs
	f.CreateSslCertificate(&compute.SslCertificate{
		Name:        primaryCertName,
		Certificate: "abc",
		SelfLink:    "existing",
	})
	f.CreateSslCertificate(&compute.SslCertificate{
		Name:        secondaryCertName,
		Certificate: "xyz",
		SelfLink:    "existing",
	})

	// Sync first cert
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(primaryCertName, lbInfo.TLS.Cert, f, t)

	// Sync with different cert
	lbInfo.TLS = &TLSCerts{Key: "key2", Cert: "cert2"}
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(secondaryCertName, lbInfo.TLS.Cert, f, t)
}

func TestCertRetentionAfterRestart(t *testing.T) {
	primaryCertName := "k8s-ssl-test"
	secondaryCertName := "k8s-ssl-1-test"
	lbInfo := &L7RuntimeInfo{
		Name:      "test",
		AllowHTTP: false,
		TLS:       &TLSCerts{Key: "key", Cert: "cert"},
	}

	f := NewFakeLoadBalancers(lbInfo.Name)
	firstPool := newFakeLoadBalancerPool(f, t)

	// Sync twice so the expected certificate uses the secondary name
	firstPool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(primaryCertName, lbInfo.TLS.Cert, f, t)
	lbInfo.TLS = &TLSCerts{Key: "key2", Cert: "cert2"}
	firstPool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(secondaryCertName, lbInfo.TLS.Cert, f, t)

	// Restart of controller represented by a new pool
	secondPool := newFakeLoadBalancerPool(f, t)
	secondPool.Sync([]*L7RuntimeInfo{lbInfo})

	// Verify second name is still used
	verifyCertAndProxyLink(secondaryCertName, lbInfo.TLS.Cert, f, t)

	// Update cert one more time to verify loop
	lbInfo.TLS = &TLSCerts{Key: "key3", Cert: "cert3"}
	secondPool.Sync([]*L7RuntimeInfo{lbInfo})
	verifyCertAndProxyLink(primaryCertName, lbInfo.TLS.Cert, f, t)

}

func verifyCertAndProxyLink(certName, certValue string, f *FakeLoadBalancers, t *testing.T) {
	cert, err := f.GetSslCertificate(certName)
	if err != nil {
		t.Fatalf("expected ssl certificate to exist: %v, err: %v", certName, err)
	}

	if cert.Certificate != certValue {
		t.Fatalf("unexpected certificate value; expected %v, actual %v", certValue, cert.Certificate)
	}

	tps, err := f.GetTargetHttpsProxy(f.tpName(true))
	if err != nil {
		t.Fatalf("expected https proxy to exist: %v, err: %v", certName, err)
	}

	if len(tps.SslCertificates) == 0 || tps.SslCertificates[0] != cert.SelfLink {
		t.Fatalf("expected ssl certificate to be linked in target proxy; Cert Link: %q; Target Proxy Certs: %v", cert.SelfLink, tps.SslCertificates)
	}
}

func TestCreateHTTPSLoadBalancerAnnotationCert(t *testing.T) {
	// This should NOT create the forwarding rule and target proxy
	// associated with the HTTP branch of this loadbalancer.
	tlsName := "external-cert-name"
	lbInfo := &L7RuntimeInfo{
		Name:      "test",
		AllowHTTP: false,
		TLSName:   tlsName,
	}
	f := NewFakeLoadBalancers(lbInfo.Name)
	f.CreateSslCertificate(&compute.SslCertificate{
		Name: tlsName,
	})
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil || l7 == nil {
		t.Fatalf("Expected l7 not created")
	}
	um, err := f.GetUrlMap(f.umName())
	if err != nil ||
		um.DefaultService != pool.(*L7s).glbcDefaultBackend.SelfLink {
		t.Fatalf("%v", err)
	}
	tps, err := f.GetTargetHttpsProxy(f.tpName(true))
	if err != nil || tps.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	fws, err := f.GetGlobalForwardingRule(f.fwName(true))
	if err != nil || fws.Target != tps.SelfLink {
		t.Fatalf("%v", err)
	}
}

func TestCreateBothLoadBalancers(t *testing.T) {
	// This should create 2 forwarding rules and target proxies
	// but they should use the same urlmap, and have the same
	// static ip.
	lbInfo := &L7RuntimeInfo{
		Name:      "test",
		AllowHTTP: true,
		TLS:       &TLSCerts{Key: "key", Cert: "cert"},
	}
	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil || l7 == nil {
		t.Fatalf("Expected l7 not created")
	}
	um, err := f.GetUrlMap(f.umName())
	if err != nil ||
		um.DefaultService != pool.(*L7s).glbcDefaultBackend.SelfLink {
		t.Fatalf("%v", err)
	}
	tps, err := f.GetTargetHttpsProxy(f.tpName(true))
	if err != nil || tps.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	tp, err := f.GetTargetHttpProxy(f.tpName(false))
	if err != nil || tp.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	fws, err := f.GetGlobalForwardingRule(f.fwName(true))
	if err != nil || fws.Target != tps.SelfLink {
		t.Fatalf("%v", err)
	}
	fw, err := f.GetGlobalForwardingRule(f.fwName(false))
	if err != nil || fw.Target != tp.SelfLink {
		t.Fatalf("%v", err)
	}
	ip, err := f.GetGlobalAddress(f.fwName(false))
	if err != nil || ip.Address != fw.IPAddress || ip.Address != fws.IPAddress {
		t.Fatalf("%v", err)
	}
}

func TestUpdateUrlMap(t *testing.T) {
	um1 := utils.GCEURLMap{
		"bar.example.com": {
			"/bar2": &compute.BackendService{SelfLink: "bar2svc"},
		},
	}
	um2 := utils.GCEURLMap{
		"foo.example.com": {
			"/foo1": &compute.BackendService{SelfLink: "foo1svc"},
			"/foo2": &compute.BackendService{SelfLink: "foo2svc"},
		},
		"bar.example.com": {
			"/bar1": &compute.BackendService{SelfLink: "bar1svc"},
		},
	}
	um2.PutDefaultBackend(&compute.BackendService{SelfLink: "default"})

	lbInfo := &L7RuntimeInfo{Name: "test", AllowHTTP: true}
	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, ir := range []utils.GCEURLMap{um1, um2} {
		if err := l7.UpdateUrlMap(ir); err != nil {
			t.Fatalf("%v", err)
		}
	}
	// The final map doesn't contain /bar2
	expectedMap := map[string]utils.FakeIngressRuleValueMap{
		utils.DefaultBackendKey: {
			utils.DefaultBackendKey: "default",
		},
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
		"bar.example.com": {
			"/bar1": "bar1svc",
		},
	}
	f.CheckURLMap(t, l7, expectedMap)
}

func TestUpdateUrlMapNoChanges(t *testing.T) {
	um1 := utils.GCEURLMap{
		"foo.example.com": {
			"/foo1": &compute.BackendService{SelfLink: "foo1svc"},
			"/foo2": &compute.BackendService{SelfLink: "foo2svc"},
		},
		"bar.example.com": {
			"/bar1": &compute.BackendService{SelfLink: "bar1svc"},
		},
	}
	um1.PutDefaultBackend(&compute.BackendService{SelfLink: "default"})
	um2 := utils.GCEURLMap{
		"foo.example.com": {
			"/foo1": &compute.BackendService{SelfLink: "foo1svc"},
			"/foo2": &compute.BackendService{SelfLink: "foo2svc"},
		},
		"bar.example.com": {
			"/bar1": &compute.BackendService{SelfLink: "bar1svc"},
		},
	}
	um2.PutDefaultBackend(&compute.BackendService{SelfLink: "default"})

	lbInfo := &L7RuntimeInfo{Name: "test", AllowHTTP: true}
	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, ir := range []utils.GCEURLMap{um1, um2} {
		if err := l7.UpdateUrlMap(ir); err != nil {
			t.Fatalf("%v", err)
		}
	}
	for _, call := range f.calls {
		if call == "UpdateUrlMap" {
			t.Errorf("UpdateUrlMap() should not have been called")
		}
	}
}

func TestNameParsing(t *testing.T) {
	clusterName := "123"
	firewallName := clusterName
	namer := utils.NewNamer(clusterName, firewallName)
	fullName := namer.Truncate(fmt.Sprintf("%v-%v", forwardingRulePrefix, namer.LBName("testlb")))
	annotationsMap := map[string]string{
		fmt.Sprintf("%v/forwarding-rule", utils.K8sAnnotationPrefix): fullName,
	}
	components := namer.ParseName(GCEResourceName(annotationsMap, "forwarding-rule"))
	t.Logf("%+v", components)
	if components.ClusterName != clusterName {
		t.Errorf("Failed to parse cluster name from %v, expected %v got %v", fullName, clusterName, components.ClusterName)
	}
	resourceName := "fw"
	if components.Resource != resourceName {
		t.Errorf("Failed to parse resource from %v, expected %v got %v", fullName, resourceName, components.Resource)
	}
}

func TestClusterNameChange(t *testing.T) {
	lbInfo := &L7RuntimeInfo{
		Name: "test",
		TLS:  &TLSCerts{Key: "key", Cert: "cert"},
	}
	f := NewFakeLoadBalancers(lbInfo.Name)
	pool := newFakeLoadBalancerPool(f, t)
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err := pool.Get(lbInfo.Name)
	if err != nil || l7 == nil {
		t.Fatalf("Expected l7 not created")
	}
	um, err := f.GetUrlMap(f.umName())
	if err != nil ||
		um.DefaultService != pool.(*L7s).glbcDefaultBackend.SelfLink {
		t.Fatalf("%v", err)
	}
	tps, err := f.GetTargetHttpsProxy(f.tpName(true))
	if err != nil || tps.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	fws, err := f.GetGlobalForwardingRule(f.fwName(true))
	if err != nil || fws.Target != tps.SelfLink {
		t.Fatalf("%v", err)
	}
	newName := "newName"
	namer := pool.(*L7s).namer
	namer.SetClusterName(newName)
	f.name = fmt.Sprintf("%v--%v", lbInfo.Name, newName)

	// Now the components should get renamed with the next suffix.
	pool.Sync([]*L7RuntimeInfo{lbInfo})
	l7, err = pool.Get(lbInfo.Name)
	if err != nil || namer.ParseName(l7.Name).ClusterName != newName {
		t.Fatalf("Expected L7 name to change.")
	}
	um, err = f.GetUrlMap(f.umName())
	if err != nil || namer.ParseName(um.Name).ClusterName != newName {
		t.Fatalf("Expected urlmap name to change.")
	}
	if err != nil ||
		um.DefaultService != pool.(*L7s).glbcDefaultBackend.SelfLink {
		t.Fatalf("%v", err)
	}

	tps, err = f.GetTargetHttpsProxy(f.tpName(true))
	if err != nil || tps.UrlMap != um.SelfLink {
		t.Fatalf("%v", err)
	}
	fws, err = f.GetGlobalForwardingRule(f.fwName(true))
	if err != nil || fws.Target != tps.SelfLink {
		t.Fatalf("%v", err)
	}
}

func TestInvalidClusterNameChange(t *testing.T) {
	namer := utils.NewNamer("test--123", "test--123")
	if got := namer.GetClusterName(); got != "123" {
		t.Fatalf("Expected name 123, got %v", got)
	}
	// A name with `--` should take the last token
	for _, testCase := range []struct{ newName, expected string }{
		{"foo--bar", "bar"},
		{"--", ""},
		{"", ""},
		{"foo--bar--com", "com"},
	} {
		namer.SetClusterName(testCase.newName)
		if got := namer.GetClusterName(); got != testCase.expected {
			t.Fatalf("Expected %q got %q", testCase.expected, got)
		}
	}

}
