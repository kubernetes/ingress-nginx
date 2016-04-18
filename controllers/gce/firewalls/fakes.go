/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	netset "k8s.io/kubernetes/pkg/util/net/sets"
)

type fakeFirewallRules struct {
	fw    []*compute.Firewall
	namer utils.Namer
}

func (f *fakeFirewallRules) GetFirewall(name string) (*compute.Firewall, error) {
	for _, rule := range f.fw {
		if rule.Name == name {
			return rule, nil
		}
	}
	return nil, fmt.Errorf("Firewall rule %v not found.", name)
}

func (f *fakeFirewallRules) CreateFirewall(name, msgTag string, srcRange netset.IPNet, ports []int64, hosts []string) error {
	strPorts := []string{}
	for _, p := range ports {
		strPorts = append(strPorts, fmt.Sprintf("%v", p))
	}
	f.fw = append(f.fw, &compute.Firewall{
		// To accurately mimic the cloudprovider we need to add the k8s-fw
		// prefix to the given rule name.
		Name:         f.namer.FrName(name),
		SourceRanges: srcRange.StringSlice(),
		Allowed:      []*compute.FirewallAllowed{{Ports: strPorts}},
	})
	return nil
}

func (f *fakeFirewallRules) DeleteFirewall(name string) error {
	firewalls := []*compute.Firewall{}
	exists := false
	// We need the full name for the same reason as CreateFirewall.
	name = f.namer.FrName(name)
	for _, rule := range f.fw {
		if rule.Name == name {
			exists = true
			continue
		}
		firewalls = append(firewalls, rule)
	}
	if !exists {
		return fmt.Errorf("Failed to find health check %v", name)
	}
	f.fw = firewalls
	return nil
}

func (f *fakeFirewallRules) UpdateFirewall(name, msgTag string, srcRange netset.IPNet, ports []int64, hosts []string) error {
	var exists bool
	strPorts := []string{}
	for _, p := range ports {
		strPorts = append(strPorts, fmt.Sprintf("%v", p))
	}

	// To accurately mimic the cloudprovider we need to add the k8s-fw
	// prefix to the given rule name.
	name = f.namer.FrName(name)
	for i := range f.fw {
		if f.fw[i].Name == name {
			exists = true
			f.fw[i] = &compute.Firewall{
				Name:         name,
				SourceRanges: srcRange.StringSlice(),
				Allowed:      []*compute.FirewallAllowed{{Ports: strPorts}},
			}
		}
	}
	if exists {
		return nil
	}
	return fmt.Errorf("Update failed for rule %v, srcRange %v ports %v, rule not found", name, srcRange, ports)
}

// NewFakeFirewallRules creates a fake for firewall rules.
func NewFakeFirewallRules() *fakeFirewallRules {
	return &fakeFirewallRules{fw: []*compute.Firewall{}, namer: utils.Namer{}}
}
