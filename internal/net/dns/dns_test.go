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

package dns

import (
	"net"
	"os"
	"testing"

	"k8s.io/ingress-nginx/internal/file"
)

func TestGetDNSServers(t *testing.T) {
	s, err := GetSystemNameServers()
	if err != nil {
		t.Fatalf("unexpected error reading /etc/resolv.conf file: %v", err)
	}
	if len(s) < 1 {
		t.Error("expected at least 1 nameserver in /etc/resolv.conf")
	}

	f, err := os.CreateTemp("", "fw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	os.WriteFile(f.Name(), []byte(`
	# comment
	; comment
	nameserver 2001:4860:4860::8844
	nameserver 2001:4860:4860::8888
	nameserver 8.8.8.8
	`), file.ReadWriteByUser)

	defResolvConf = f.Name()
	s, err = GetSystemNameServers()
	if err != nil {
		t.Fatalf("unexpected error reading /etc/resolv.conf file: %v", err)
	}
	if len(s) < 3 {
		t.Errorf("expected at 3 nameservers but %v returned", len(s))
	}

	eip := net.ParseIP("2001:4860:4860::8844")
	if !s[0].Equal(eip) {
		t.Errorf("expected %v as nameservers but %v returned", eip, s[0])
	}
}
