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
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress/controllers/gce/utils"
	netset "k8s.io/kubernetes/pkg/util/net/sets"
)

const allCIDR = "0.0.0.0/0"

func TestSyncFirewallPool(t *testing.T) {
	namer := utils.NewNamer("ABC", "XYZ")
	fwp := NewFakeFirewallsProvider(namer)
	fp := NewFirewallPool(fwp, namer)
	ruleName := namer.FrName(namer.FrSuffix())

	// Test creating a firewall rule via Sync
	nodePorts := []int64{80, 443, 3000}
	nodes := []string{"node-a", "node-b", "node-c"}
	err := fp.Sync(nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Sync to fewer ports
	nodePorts = []int64{80, 443}
	err = fp.Sync(nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	srcRanges, _ := netset.ParseIPNets(allCIDR)
	err = fwp.UpdateFirewall(namer.FrSuffix(), "", srcRanges, nodePorts, nodes)
	if err != nil {
		t.Errorf("failed to update firewall rule, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, []string{allCIDR}, t)

	// Run Sync and expect l7 src ranges to be returned
	err = fp.Sync(nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Add node and expect firewall to remain the same
	// NOTE: See computeHostTag(..) in gce cloudprovider
	nodes = []string{"node-a", "node-b", "node-c", "node-d"}
	err = fp.Sync(nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Remove all ports and expect firewall rule to disappear
	nodePorts = []int64{}
	err = fp.Sync(nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}

	err = fp.Shutdown()
	if err != nil {
		t.Errorf("unexpected err when deleting firewall, err: %v", err)
	}
}

func verifyFirewallRule(fwp *fakeFirewallsProvider, ruleName string, expectedPorts []int64, expectedNodes, expectedCIDRs []string, t *testing.T) {
	var strPorts []string
	for _, v := range expectedPorts {
		strPorts = append(strPorts, strconv.FormatInt(v, 10))
	}

	// Verify firewall rule was created
	f, err := fwp.GetFirewall(ruleName)
	if err != nil {
		t.Errorf("could not retrieve firewall via cloud api, err %v", err)
	}

	// Verify firewall rule has correct ports
	if !sets.NewString(f.Allowed[0].Ports...).Equal(sets.NewString(strPorts...)) {
		t.Errorf("allowed ports doesn't equal expected ports, Actual: %v, Expected: %v", f.Allowed[0].Ports, strPorts)
	}

	// Verify firewall rule has correct CIDRs
	if !sets.NewString(f.SourceRanges...).Equal(sets.NewString(expectedCIDRs...)) {
		t.Errorf("source CIDRs doesn't equal expected CIDRs. Actual: %v, Expected: %v", f.SourceRanges, expectedCIDRs)
	}
}
