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

package net

import (
	"net"
	"testing"
)

func TestIsIPV6(t *testing.T) {
	tests := []struct {
		in     net.IP
		isIPV6 bool
	}{
		{net.ParseIP("2001:4860:4860::8844"), true},
		{net.ParseIP("2001:4860:4860::8888"), true},
		{net.ParseIP("0:0:0:0:0:ffff:c868:8165"), true},
		{net.ParseIP("2001:db8:85a3::8a2e:370:7334"), true},
		{net.ParseIP("::1"), true},
		{net.ParseIP("8.8.8.8"), false},
	}

	for _, test := range tests {
		isIPV6 := IsIPV6(test.in)
		if isIPV6 && !test.isIPV6 {
			t.Fatalf("%v expected %v but returned %v", test.in, test.isIPV6, isIPV6)
		}
	}
}

func TestIsPortAvailable(t *testing.T) {
	if !IsPortAvailable(0) {
		t.Fatal("expected port 0 to be available (random port) but returned false")
	}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer ln.Close()

	p := ln.Addr().(*net.TCPAddr).Port
	if IsPortAvailable(p) {
		t.Fatalf("expected port %v to not be available", p)
	}
}

/*
// TODO: this test should be optional or running behind a flag
func TestIsIPv6Enabled(t *testing.T) {
	isEnabled := IsIPv6Enabled()
	if !isEnabled {
		t.Fatalf("expected IPV6 be enabled")
	}
}
*/
