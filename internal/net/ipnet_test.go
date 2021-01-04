/*
Copyright 2017 The Kubernetes Authors.

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

package net

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewIPSet(t *testing.T) {
	ipsets, ips, err := ParseIPNets("1.0.0.0", "2.0.0.0/8", "3.0.0.0/8")
	if err != nil {
		t.Errorf("error parsing IPNets: %v", err)
	}
	if len(ipsets) != 2 {
		t.Errorf("Expected len=2: %d", len(ipsets))
	}
	if len(ips) != 1 {
		t.Errorf("Expected len=1: %d", len(ips))
	}
}

func TestParseCIDRs(t *testing.T) {
	cidr, _ := ParseCIDRs("invalid.com")
	if cidr != nil {
		t.Errorf("expected %v but got %v", nil, cidr)
	}

	expected := []string{"192.0.0.1", "192.0.1.0/24"}
	cidr, err := ParseCIDRs("192.0.0.1, 192.0.1.0/24")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	sort.Strings(cidr)
	if !reflect.DeepEqual(expected, cidr) {
		t.Errorf("expected %v but got %v", expected, cidr)
	}
}
