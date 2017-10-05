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

	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress/controllers/gce/utils"
)

func TestSyncFirewallPool(t *testing.T) {
	namer := utils.NewNamer("ABC", "XYZ")
	fwp := NewFakeFirewallsProvider(false, false)
	fp := NewFirewallPool(fwp, namer, nil)
	ruleName := namer.FrName(namer.FrSuffix())

	// Test creating a firewall rule via Sync
	nodePorts := []int64{80, 443, 3000}
	nodes := []string{"node-a", "node-b", "node-c"}
	err := fp.Sync(nodePorts, nodes, nil)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Sync to fewer ports
	nodePorts = []int64{80, 443}
	err = fp.Sync(nodePorts, nodes, nil)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	firewall, err := fp.(*FirewallRules).createFirewallObject(ruleName, "", nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when creating firewall object, err: %v", err)
	}

	err = fwp.doUpdateFirewall(firewall)
	if err != nil {
		t.Errorf("failed to update firewall rule, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Run Sync and expect l7 src ranges to be returned
	err = fp.Sync(nodePorts, nodes, nil)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Add node and expect firewall to remain the same
	// NOTE: See computeHostTag(..) in gce cloudprovider
	nodes = []string{"node-a", "node-b", "node-c", "node-d"}
	err = fp.Sync(nodePorts, nodes, nil)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)

	// Remove all ports and expect firewall rule to disappear
	nodePorts = []int64{}
	err = fp.Sync(nodePorts, nodes, nil)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}

	err = fp.Shutdown()
	if err != nil {
		t.Errorf("unexpected err when deleting firewall, err: %v", err)
	}
}

// TestSyncOnXPNWithPermission tests that firwall sync continues to work when OnXPN=true
func TestSyncOnXPNWithPermission(t *testing.T) {
	namer := utils.NewNamer("ABC", "XYZ")
	fwp := NewFakeFirewallsProvider(true, false)
	fp := NewFirewallPool(fwp, namer, nil)
	ruleName := namer.FrName(namer.FrSuffix())

	// Test creating a firewall rule via Sync
	nodePorts := []int64{80, 443, 3000}
	nodes := []string{"node-a", "node-b", "node-c"}
	err := fp.Sync(nodePorts, nodes, nil)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	verifyFirewallRule(fwp, ruleName, nodePorts, nodes, l7SrcRanges, t)
}

// TestSyncOnXPNReadOnly tests that controller behavior is accurate when the controller
// does not have permission to create/update/delete firewall rules.
// Sync should NOT return an error. An event should be raised.
func TestSyncOnXPNReadOnly(t *testing.T) {
	ing := &extensions.Ingress{ObjectMeta: meta_v1.ObjectMeta{Name: "xpn-ingress"}}
	var events []string
	eventer := func(ing *extensions.Ingress, reason, msg string) {
		events = append(events, msg)
	}

	namer := utils.NewNamer("ABC", "XYZ")
	fwp := NewFakeFirewallsProvider(true, true)
	fp := NewFirewallPool(fwp, namer, eventer)
	ruleName := namer.FrName(namer.FrSuffix())

	// Test creating a firewall rule via Sync
	nodePorts := []int64{80, 443, 3000}
	nodes := []string{"node-a", "node-b", "node-c"}
	err := fp.Sync(nodePorts, nodes, ing)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}

	// Expect an event saying a firewall needs to be created
	if len(events) != 1 {
		t.Errorf("expected %v events but received %v: %+v", 1, len(events), events)
	}

	// Clear events
	events = events[:0]

	// Manually create the firewall
	firewall, err := fp.(*FirewallRules).createFirewallObject(ruleName, "", nodePorts, nodes)
	if err != nil {
		t.Errorf("unexpected err when creating firewall object, err: %v", err)
	}
	err = fwp.doCreateFirewall(firewall)
	if err != nil {
		t.Errorf("unexpected err when creating firewall, err: %v", err)
	}

	// Run sync again with same state - expect no event
	err = fp.Sync(nodePorts, nodes, ing)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	if len(events) > 0 {
		t.Errorf("received unexpected event(s): %+v", events)
	}

	// Modify nodePorts to cause an event
	nodePorts = append(nodePorts, 3001)

	// Run sync again with same state - expect no event
	err = fp.Sync(nodePorts, nodes, ing)
	if err != nil {
		t.Errorf("unexpected err when syncing firewall, err: %v", err)
	}
	// Expect an event saying a firewall needs to be created
	if len(events) != 1 {
		t.Errorf("expected %v events but received %v: %+v", 1, len(events), events)
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
