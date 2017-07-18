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
	compute "google.golang.org/api/compute/v1"
)

// SingleFirewallPool syncs the firewall rule for L7 traffic.
type SingleFirewallPool interface {
	// TODO: Take a list of node ports for the firewall.
	Sync(nodePorts []int64, nodeNames []string) error
	Shutdown() error
}

// Firewall interfaces with the GCE firewall api.
// This interface is a little different from the rest because it dovetails into
// the same firewall methods used by the TCPLoadBalancer.
type Firewall interface {
	CreateFirewall(f *compute.Firewall) error
	GetFirewall(name string) (*compute.Firewall, error)
	DeleteFirewall(name string) error
	UpdateFirewall(f *compute.Firewall) error
	GetNodeTags(nodeNames []string) ([]string, error)
	NetworkURL() string
}
