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
	"strconv"

	"github.com/golang/glog"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	netset "k8s.io/kubernetes/pkg/util/net/sets"

	"k8s.io/ingress/controllers/gce/utils"
)

// Src ranges from which the GCE L7 performs health checks.
var l7SrcRanges = []string{"130.211.0.0/22", "35.191.0.0/16"}

// FirewallRules manages firewall rules.
type FirewallRules struct {
	cloud     Firewall
	namer     *utils.Namer
	srcRanges []string
}

// NewFirewallPool creates a new firewall rule manager.
// cloud: the cloud object implementing Firewall.
// namer: cluster namer.
func NewFirewallPool(cloud Firewall, namer *utils.Namer) SingleFirewallPool {
	_, err := netset.ParseIPNets(l7SrcRanges...)
	if err != nil {
		glog.Fatalf("Could not parse L7 src ranges %v for firewall rule: %v", l7SrcRanges, err)
	}
	return &FirewallRules{cloud: cloud, namer: namer, srcRanges: l7SrcRanges}
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

	firewall, err := fr.createFirewallObject(name, "GCE L7 firewall rule", nodePorts, nodeNames)
	if err != nil {
		return err
	}

	if rule == nil {
		glog.Infof("Creating global l7 firewall rule %v", name)
		return fr.cloud.CreateFirewall(firewall)
	}

	requiredPorts := sets.NewString()
	for _, p := range nodePorts {
		requiredPorts.Insert(strconv.Itoa(int(p)))
	}
	existingPorts := sets.NewString()
	for _, allowed := range rule.Allowed {
		for _, p := range allowed.Ports {
			existingPorts.Insert(p)
		}
	}

	requiredCIDRs := sets.NewString(l7SrcRanges...)
	existingCIDRs := sets.NewString(rule.SourceRanges...)

	// Do not update if ports and source cidrs are not outdated.
	// NOTE: We are not checking if nodeNames matches the firewall targetTags
	if requiredPorts.Equal(existingPorts) && requiredCIDRs.Equal(existingCIDRs) {
		glog.V(4).Info("Firewall does not need update of ports or source ranges")
		return nil
	}
	glog.V(3).Infof("Firewall %v already exists, updating nodeports %v", name, nodePorts)
	return fr.cloud.UpdateFirewall(firewall)
}

// Shutdown shuts down this firewall rules manager.
func (fr *FirewallRules) Shutdown() error {
	name := fr.namer.FrName(fr.namer.FrSuffix())
	glog.Infof("Deleting firewall %v", name)
	err := fr.cloud.DeleteFirewall(name)
	if err != nil && utils.IsHTTPErrorCode(err, 404) {
		glog.Infof("Firewall with name %v didn't exist at Shutdown", name)
		return nil
	}
	return err
}

// GetFirewall just returns the firewall object corresponding to the given name.
// TODO: Currently only used in testing. Modify so we don't leak compute
// objects out of this interface by returning just the (src, ports, error).
func (fr *FirewallRules) GetFirewall(name string) (*compute.Firewall, error) {
	return fr.cloud.GetFirewall(name)
}

func (fr *FirewallRules) createFirewallObject(firewallName, description string, nodePorts []int64, nodeNames []string) (*compute.Firewall, error) {
	ports := make([]string, len(nodePorts))
	for ix := range nodePorts {
		ports[ix] = strconv.Itoa(int(nodePorts[ix]))
	}

	// If the node tags to be used for this cluster have been predefined in the
	// provider config, just use them. Otherwise, invoke computeHostTags method to get the tags.
	targetTags, err := fr.cloud.GetNodeTags(nodeNames)
	if err != nil {
		return nil, err
	}

	return &compute.Firewall{
		Name:         firewallName,
		Description:  description,
		SourceRanges: fr.srcRanges,
		Network:      fr.cloud.NetworkURL(),
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      ports,
			},
		},
		TargetTags: targetTags,
	}, nil
}
