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

package firewalls

import (
	"fmt"
	"strconv"

	compute "google.golang.org/api/compute/v1"
	netset "k8s.io/kubernetes/pkg/util/net/sets"

	"k8s.io/ingress/controllers/gce/utils"
)

type fakeFirewallsProvider struct {
	fw    map[string]*compute.Firewall
	namer *utils.Namer
}

// NewFakeFirewallsProvider creates a fake for firewall rules.
func NewFakeFirewallsProvider(namer *utils.Namer) *fakeFirewallsProvider {
	return &fakeFirewallsProvider{
		fw:    make(map[string]*compute.Firewall),
		namer: namer,
	}
}

func (f *fakeFirewallsProvider) GetFirewall(prefixedName string) (*compute.Firewall, error) {
	rule, exists := f.fw[prefixedName]
	if exists {
		return rule, nil
	}
	return nil, utils.FakeGoogleAPINotFoundErr()
}

func (f *fakeFirewallsProvider) CreateFirewall(name, msgTag string, srcRange netset.IPNet, ports []int64, hosts []string) error {
	prefixedName := f.namer.FrName(name)
	strPorts := []string{}
	for _, p := range ports {
		strPorts = append(strPorts, strconv.FormatInt(p, 10))
	}
	if _, exists := f.fw[prefixedName]; exists {
		return fmt.Errorf("firewall rule %v already exists", prefixedName)
	}

	f.fw[prefixedName] = &compute.Firewall{
		// To accurately mimic the cloudprovider we need to add the k8s-fw
		// prefix to the given rule name.
		Name:         prefixedName,
		SourceRanges: srcRange.StringSlice(),
		Allowed:      []*compute.FirewallAllowed{{Ports: strPorts}},
		TargetTags:   hosts, // WARNING: This is actually not correct, but good enough for testing this package
	}
	return nil
}

func (f *fakeFirewallsProvider) DeleteFirewall(name string) error {
	// We need the full name for the same reason as CreateFirewall.
	prefixedName := f.namer.FrName(name)
	_, exists := f.fw[prefixedName]
	if !exists {
		return utils.FakeGoogleAPINotFoundErr()
	}

	delete(f.fw, prefixedName)
	return nil
}

func (f *fakeFirewallsProvider) UpdateFirewall(name, msgTag string, srcRange netset.IPNet, ports []int64, hosts []string) error {
	strPorts := []string{}
	for _, p := range ports {
		strPorts = append(strPorts, strconv.FormatInt(p, 10))
	}

	// We need the full name for the same reason as CreateFirewall.
	prefixedName := f.namer.FrName(name)
	_, exists := f.fw[prefixedName]
	if !exists {
		return fmt.Errorf("update failed for rule %v, srcRange %v ports %v, rule not found", prefixedName, srcRange, ports)
	}

	f.fw[prefixedName] = &compute.Firewall{
		Name:         name,
		SourceRanges: srcRange.StringSlice(),
		Allowed:      []*compute.FirewallAllowed{{Ports: strPorts}},
		TargetTags:   hosts, // WARNING: This is actually not correct, but good enough for testing this package
	}
	return nil
}
