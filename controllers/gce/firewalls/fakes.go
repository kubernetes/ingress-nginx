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

	compute "google.golang.org/api/compute/v1"

	"k8s.io/ingress/controllers/gce/utils"
)

type fakeFirewallsProvider struct {
	fw         map[string]*compute.Firewall
	networkUrl string
}

// NewFakeFirewallsProvider creates a fake for firewall rules.
func NewFakeFirewallsProvider() *fakeFirewallsProvider {
	return &fakeFirewallsProvider{
		fw: make(map[string]*compute.Firewall),
	}
}

func (ff *fakeFirewallsProvider) GetFirewall(name string) (*compute.Firewall, error) {
	rule, exists := ff.fw[name]
	if exists {
		return rule, nil
	}
	return nil, utils.FakeGoogleAPINotFoundErr()
}

func (ff *fakeFirewallsProvider) CreateFirewall(f *compute.Firewall) error {
	if _, exists := ff.fw[f.Name]; exists {
		return fmt.Errorf("firewall rule %v already exists", f.Name)
	}
	ff.fw[f.Name] = f
	return nil
}

func (ff *fakeFirewallsProvider) DeleteFirewall(name string) error {
	// We need the full name for the same reason as CreateFirewall.
	_, exists := ff.fw[name]
	if !exists {
		return utils.FakeGoogleAPINotFoundErr()
	}

	delete(ff.fw, name)
	return nil
}

func (ff *fakeFirewallsProvider) UpdateFirewall(f *compute.Firewall) error {
	// We need the full name for the same reason as CreateFirewall.
	_, exists := ff.fw[f.Name]
	if !exists {
		return fmt.Errorf("update failed for rule %v, srcRange %v ports %+v, rule not found", f.Name, f.SourceRanges, f.Allowed)
	}

	ff.fw[f.Name] = f
	return nil
}

func (ff *fakeFirewallsProvider) NetworkURL() string {
	return ff.networkUrl
}

func (ff *fakeFirewallsProvider) GetNodeTags(nodeNames []string) ([]string, error) {
	return nodeNames, nil
}
