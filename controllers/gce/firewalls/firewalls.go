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
	"github.com/golang/glog"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	netset "k8s.io/kubernetes/pkg/util/net/sets"
)

// Src range from which the GCE L7 performs health checks.
const l7SrcRange = "130.211.0.0/22"

// FirewallRules manages firewall rules.
type FirewallRules struct {
	cloud    Firewall
	namer    utils.Namer
	srcRange netset.IPNet
}

// NewFirewallPool creates a new firewall rule manager.
// cloud: the cloud object implementing Firewall.
// namer: cluster namer.
func NewFirewallPool(cloud Firewall, namer utils.Namer) SingleFirewallPool {
	srcNetSet, err := netset.ParseIPNets(l7SrcRange)
	if err != nil {
		glog.Fatalf("Could not parse L7 src range %v for firewall rule: %v", l7SrcRange, err)
	}
	return &FirewallRules{cloud: cloud, namer: namer, srcRange: srcNetSet}
}

// Sync sync firewall rules with the cloud.
func (fr *FirewallRules) Sync(nodePorts []int64, nodeNames []string) error {
	if len(nodePorts) == 0 {
		return fr.Shutdown()
	}
	// Firewall rule prefix must match that inserted by the gce library.
	suffix := fr.namer.FrSuffix()
	// TODO: Fix upstream gce cloudprovider lib so GET also takes the suffix
	// instead of the whole name.
	name := fr.namer.FrName(suffix)
	rule, _ := fr.cloud.GetFirewall(name)
	if rule == nil {
		glog.Infof("Creating global l7 firewall rule %v", name)
		return fr.cloud.CreateFirewall(suffix, "GCE L7 firewall rule", fr.srcRange, nodePorts, nodeNames)
	}
	glog.V(3).Infof("Firewall rule already %v exists, verifying for nodeports %v", name, nodePorts)
	return fr.cloud.UpdateFirewall(suffix, "GCE L7 firewall rule", fr.srcRange, nodePorts, nodeNames)
}

// Shutdown shuts down this firewall rules manager.
func (fr *FirewallRules) Shutdown() error {
	glog.Infof("Deleting fireawll rule with suffix %v", fr.namer.FrSuffix())
	return fr.cloud.DeleteFirewall(fr.namer.FrSuffix())
}

// GetFirewall just returns the firewall object corresponding to the given name.
// TODO: Currently only used in testing. Modify so we don't leak compute
// objects out of this interface by returning just the (src, ports, error).
func (fr *FirewallRules) GetFirewall(name string) (*compute.Firewall, error) {
	return fr.cloud.GetFirewall(name)
}
