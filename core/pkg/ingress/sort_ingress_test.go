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

package ingress

import (
	"testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildBackendByNameServers() BackendByNameServers {
	return []*Backend{
		{
			Name:      "foo1",
			Secure:    true,
			Endpoints: []Endpoint{},
		},
		{
			Name:      "foo2",
			Secure:    false,
			Endpoints: []Endpoint{},
		},
		{
			Name:      "foo3",
			Secure:    true,
			Endpoints: []Endpoint{},
		},
	}
}

func TestBackendByNameServersLen(t *testing.T) {
	fooTests := []struct {
		backends BackendByNameServers
		el       int
	}{
		{[]*Backend{}, 0},
		{buildBackendByNameServers(), 3},
		{nil, 0},
	}

	for _, fooTest := range fooTests {
		r := fooTest.backends.Len()
		if r != fooTest.el {
			t.Errorf("returned %v but expected %v for the len of BackendByNameServers: %v", r, fooTest.el, fooTest.backends)
		}
	}
}

func TestBackendByNameServersSwap(t *testing.T) {
	fooTests := []struct {
		backends BackendByNameServers
		i        int
		j        int
	}{
		{buildBackendByNameServers(), 0, 1},
		{buildBackendByNameServers(), 2, 1},
	}

	for _, fooTest := range fooTests {
		fooi := fooTest.backends[fooTest.i]
		fooj := fooTest.backends[fooTest.j]
		fooTest.backends.Swap(fooTest.i, fooTest.j)
		if fooi.Name != fooTest.backends[fooTest.j].Name || fooj.Name != fooTest.backends[fooTest.i].Name {
			t.Errorf("failed to swap for ByNameServers, foo: %v", fooTest)
		}
	}
}

func TestBackendByNameServersLess(t *testing.T) {
	fooTests := []struct {
		backends BackendByNameServers
		i        int
		j        int
		er       bool
	}{
		// order by name
		{buildBackendByNameServers(), 0, 2, true},
		{buildBackendByNameServers(), 1, 0, false},
	}

	for _, fooTest := range fooTests {
		r := fooTest.backends.Less(fooTest.i, fooTest.j)
		if r != fooTest.er {
			t.Errorf("returned %v but expected %v for the foo: %v", r, fooTest.er, fooTest)
		}
	}
}

func buildEndpointByAddrPort() EndpointByAddrPort {
	return []Endpoint{
		{
			Address:     "127.0.0.1",
			Port:        "8080",
			MaxFails:    3,
			FailTimeout: 10,
		},
		{
			Address:     "127.0.0.1",
			Port:        "8081",
			MaxFails:    3,
			FailTimeout: 10,
		},
		{
			Address:     "127.0.0.1",
			Port:        "8082",
			MaxFails:    3,
			FailTimeout: 10,
		},
		{
			Address:     "127.0.0.2",
			Port:        "8082",
			MaxFails:    3,
			FailTimeout: 10,
		},
	}
}

func TestEndpointByAddrPortLen(t *testing.T) {
	fooTests := []struct {
		endpoints EndpointByAddrPort
		el        int
	}{
		{[]Endpoint{}, 0},
		{buildEndpointByAddrPort(), 4},
		{nil, 0},
	}

	for _, fooTest := range fooTests {
		r := fooTest.endpoints.Len()
		if r != fooTest.el {
			t.Errorf("returned %v but expected %v for the len of EndpointByAddrPort: %v", r, fooTest.el, fooTest.endpoints)
		}
	}
}

func TestEndpointByAddrPortSwap(t *testing.T) {
	fooTests := []struct {
		endpoints EndpointByAddrPort
		i         int
		j         int
	}{
		{buildEndpointByAddrPort(), 0, 1},
		{buildEndpointByAddrPort(), 2, 1},
	}

	for _, fooTest := range fooTests {
		fooi := fooTest.endpoints[fooTest.i]
		fooj := fooTest.endpoints[fooTest.j]
		fooTest.endpoints.Swap(fooTest.i, fooTest.j)
		if fooi.Port != fooTest.endpoints[fooTest.j].Port ||
			fooi.Address != fooTest.endpoints[fooTest.j].Address ||
			fooj.Port != fooTest.endpoints[fooTest.i].Port ||
			fooj.Address != fooTest.endpoints[fooTest.i].Address {
			t.Errorf("failed to swap for EndpointByAddrPort, foo: %v", fooTest)
		}
	}
}

func TestEndpointByAddrPortLess(t *testing.T) {
	fooTests := []struct {
		endpoints EndpointByAddrPort
		i         int
		j         int
		er        bool
	}{
		// 1) order by name
		// 2) order by port(if the name is the same one)
		{buildEndpointByAddrPort(), 0, 1, true},
		{buildEndpointByAddrPort(), 2, 1, false},
		{buildEndpointByAddrPort(), 2, 3, true},
	}

	for _, fooTest := range fooTests {
		r := fooTest.endpoints.Less(fooTest.i, fooTest.j)
		if r != fooTest.er {
			t.Errorf("returned %v but expected %v for the foo: %v", r, fooTest.er, fooTest)
		}
	}
}

func buildServerByName() ServerByName {
	return []*Server{
		{
			Hostname:       "foo1",
			SSLPassthrough: true,
			SSLCertificate: "foo1_cert",
			SSLPemChecksum: "foo1_pem",
			Locations:      []*Location{},
		},
		{
			Hostname:       "foo2",
			SSLPassthrough: true,
			SSLCertificate: "foo2_cert",
			SSLPemChecksum: "foo2_pem",
			Locations:      []*Location{},
		},
		{
			Hostname:       "foo3",
			SSLPassthrough: false,
			SSLCertificate: "foo3_cert",
			SSLPemChecksum: "foo3_pem",
			Locations:      []*Location{},
		},
		{
			Hostname:       "_",
			SSLPassthrough: false,
			SSLCertificate: "foo4_cert",
			SSLPemChecksum: "foo4_pem",
			Locations:      []*Location{},
		},
	}
}

func TestServerByNameLen(t *testing.T) {
	fooTests := []struct {
		servers ServerByName
		el      int
	}{
		{[]*Server{}, 0},
		{buildServerByName(), 4},
		{nil, 0},
	}

	for _, fooTest := range fooTests {
		r := fooTest.servers.Len()
		if r != fooTest.el {
			t.Errorf("returned %v but expected %v for the len of ServerByName: %v", r, fooTest.el, fooTest.servers)
		}
	}
}

func TestServerByNameSwap(t *testing.T) {
	fooTests := []struct {
		servers ServerByName
		i       int
		j       int
	}{
		{buildServerByName(), 0, 1},
		{buildServerByName(), 2, 1},
	}

	for _, fooTest := range fooTests {
		fooi := fooTest.servers[fooTest.i]
		fooj := fooTest.servers[fooTest.j]
		fooTest.servers.Swap(fooTest.i, fooTest.j)
		if fooi.Hostname != fooTest.servers[fooTest.j].Hostname ||
			fooj.Hostname != fooTest.servers[fooTest.i].Hostname {
			t.Errorf("failed to swap for ServerByName, foo: %v", fooTest)
		}
	}
}

func TestServerByNameLess(t *testing.T) {
	fooTests := []struct {
		servers ServerByName
		i       int
		j       int
		er      bool
	}{
		{buildServerByName(), 0, 1, true},
		{buildServerByName(), 2, 1, false},
		{buildServerByName(), 2, 3, false},
	}

	for _, fooTest := range fooTests {
		r := fooTest.servers.Less(fooTest.i, fooTest.j)
		if r != fooTest.er {
			t.Errorf("returned %v but expected %v for the foo: %v", r, fooTest.er, fooTest)
		}
	}
}

func buildLocationByPath() LocationByPath {
	return []*Location{
		{
			Path:         "a",
			IsDefBackend: true,
			Backend:      "a_back",
		},
		{
			Path:         "b",
			IsDefBackend: true,
			Backend:      "b_back",
		},
		{
			Path:         "c",
			IsDefBackend: true,
			Backend:      "c_back",
		},
	}
}

func TestLocationByPath(t *testing.T) {
	fooTests := []struct {
		locations LocationByPath
		el        int
	}{
		{[]*Location{}, 0},
		{buildLocationByPath(), 3},
		{nil, 0},
	}

	for _, fooTest := range fooTests {
		r := fooTest.locations.Len()
		if r != fooTest.el {
			t.Errorf("returned %v but expected %v for the len of LocationByPath: %v", r, fooTest.el, fooTest.locations)
		}
	}
}

func TestLocationByPathSwap(t *testing.T) {
	fooTests := []struct {
		locations LocationByPath
		i         int
		j         int
	}{
		{buildLocationByPath(), 0, 1},
		{buildLocationByPath(), 2, 1},
	}

	for _, fooTest := range fooTests {
		fooi := fooTest.locations[fooTest.i]
		fooj := fooTest.locations[fooTest.j]
		fooTest.locations.Swap(fooTest.i, fooTest.j)
		if fooi.Path != fooTest.locations[fooTest.j].Path ||
			fooj.Path != fooTest.locations[fooTest.i].Path {
			t.Errorf("failed to swap for LocationByPath, foo: %v", fooTest)
		}
	}
}

func TestLocationByPathLess(t *testing.T) {
	fooTests := []struct {
		locations LocationByPath
		i         int
		j         int
		er        bool
	}{
		// sorts location by path in descending order
		{buildLocationByPath(), 0, 1, false},
		{buildLocationByPath(), 2, 1, true},
	}

	for _, fooTest := range fooTests {
		r := fooTest.locations.Less(fooTest.i, fooTest.j)
		if r != fooTest.er {
			t.Errorf("returned %v but expected %v for the foo: %v", r, fooTest.er, fooTest)
		}
	}
}

func TestGetObjectKindForSSLCert(t *testing.T) {
	fk := &SSLCert{
		ObjectMeta:  meta_v1.ObjectMeta{},
		CAFileName:  "ca_file",
		PemFileName: "pemfile",
		PemSHA:      "pem_sha",
		CN:          []string{},
	}

	r := fk.GetObjectKind()
	if r == nil {
		t.Errorf("Returned nil but expected a valid ObjectKind")
	}
}
