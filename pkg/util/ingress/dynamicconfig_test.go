/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

func TestConfigureDynamically(t *testing.T) {
	listener, err := tryListen("tcp", fmt.Sprintf(":%v", nginx.StatusPort))
	if err != nil {
		t.Fatalf("creating tcp listener: %s", err)
	}
	defer listener.Close()

	streamListener, err := tryListen("tcp", fmt.Sprintf(":%v", nginx.StreamPort))
	if err != nil {
		t.Fatalf("creating tcp listener: %s", err)
	}
	defer streamListener.Close()

	endpointStats := map[string]int{"/configuration/backends": 0, "/configuration/general": 0, "/configuration/servers": 0}
	resetEndpointStats := func() {
		for k := range endpointStats {
			endpointStats[k] = 0
		}
	}

	server := &httptest.Server{
		Listener: listener,
		Config: &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)

				if r.Method != "POST" {
					t.Errorf("expected a 'POST' request, got '%s'", r.Method)
				}

				b, err := io.ReadAll(r.Body)
				if err != nil && err != io.EOF {
					t.Fatal(err)
				}
				body := string(b)

				endpointStats[r.URL.Path]++

				switch r.URL.Path {
				case "/configuration/backends":
					{
						if strings.Contains(body, "target") {
							t.Errorf("unexpected target reference in JSON content: %v", body)
						}

						if !strings.Contains(body, "service") {
							t.Errorf("service reference should be present in JSON content: %v", body)
						}
					}
				case "/configuration/general":
					{
					}
				case "/configuration/servers":
					{
						if !strings.Contains(body, `{"certificates":{},"servers":{"myapp.fake":"-1"}}`) {
							t.Errorf("should be present in JSON content: %v", body)
						}
					}
				default:
					t.Errorf("unknown request to %s", r.URL.Path)
				}
			}),
		},
	}
	defer server.Close()
	server.Start()

	target := &apiv1.ObjectReference{}

	backends := []*ingress.Backend{{
		Name:    "fakenamespace-myapp-80",
		Service: &apiv1.Service{},
		Endpoints: []ingress.Endpoint{
			{
				Address: "10.0.0.1",
				Port:    "8080",
				Target:  target,
			},
			{
				Address: "10.0.0.2",
				Port:    "8080",
				Target:  target,
			},
		},
	}}

	servers := []*ingress.Server{{
		Hostname: "myapp.fake",
		Locations: []*ingress.Location{
			{
				Path:    "/",
				Backend: "fakenamespace-myapp-80",
				Service: &apiv1.Service{},
			},
		},
	}}

	commonConfig := &ingress.Configuration{
		Backends: backends,
		Servers:  servers,
	}

	runningConfig := &ingress.Configuration{}

	err = ConfigureDynamically(commonConfig, runningConfig)
	if err != nil {
		t.Errorf("unexpected error posting dynamic configuration: %v", err)
	}
	if commonConfig.Backends[0].Endpoints[0].Target != target {
		t.Errorf("unexpected change in the configuration object after configureDynamically invocation")
	}

	resetEndpointStats()
	runningConfig.Backends = backends
	err = ConfigureDynamically(commonConfig, runningConfig)
	if err != nil {
		t.Errorf("unexpected error posting dynamic configuration: %v", err)
	}
	for endpoint, count := range endpointStats {
		if endpoint == "/configuration/backends" {
			if count != 0 {
				t.Errorf("Expected %v to receive %d requests but received %d.", endpoint, 0, count)
			}
		}
	}

	resetEndpointStats()
	runningConfig.Servers = servers
	err = ConfigureDynamically(commonConfig, runningConfig)
	if err != nil {
		t.Errorf("unexpected error posting dynamic configuration: %v", err)
	}
	if count := endpointStats["/configuration/backends"]; count != 0 {
		t.Errorf("Expected %v to receive %d requests but received %d.", "/configuration/backends", 0, count)
	}
	if count := endpointStats["/configuration/servers"]; count != 0 {
		t.Errorf("Expected %v to receive %d requests but received %d.", "/configuration/servers", 0, count)
	}

	resetEndpointStats()
	err = ConfigureDynamically(commonConfig, runningConfig)
	if err != nil {
		t.Errorf("unexpected error posting dynamic configuration: %v", err)
	}
	for endpoint, count := range endpointStats {
		if count != 0 {
			t.Errorf("Expected %v to receive %d requests but received %d.", endpoint, 0, count)
		}
	}
}

func TestConfigureCertificates(t *testing.T) {
	listener, err := tryListen("tcp", fmt.Sprintf(":%v", nginx.StatusPort))
	if err != nil {
		t.Fatalf("creating tcp listener: %s", err)
	}
	defer listener.Close()

	streamListener, err := tryListen("tcp", fmt.Sprintf(":%v", nginx.StreamPort))
	if err != nil {
		t.Fatalf("creating tcp listener: %s", err)
	}
	defer streamListener.Close()

	servers := []*ingress.Server{
		{
			Hostname: "myapp.fake",
			SSLCert: &ingress.SSLCert{
				PemCertKey: "fake-cert",
				UID:        "c89a5111-b2e9-4af8-be19-c2a4a924c256",
			},
		},
		{
			Hostname: "myapp.nossl",
		},
	}

	server := &httptest.Server{
		Listener: listener,
		Config: &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)

				if r.Method != "POST" {
					t.Errorf("expected a 'POST' request, got '%s'", r.Method)
				}

				b, err := io.ReadAll(r.Body)
				if err != nil && err != io.EOF {
					t.Fatal(err)
				}
				var conf sslConfiguration
				err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(b, &conf)
				if err != nil {
					t.Fatal(err)
				}

				if len(servers) != len(conf.Servers) {
					t.Errorf("Expected servers to be the same length as the posted servers")
				}

				for _, server := range servers {
					if server.SSLCert == nil {
						if conf.Servers[server.Hostname] != emptyUID {
							t.Errorf("Expected server %s to have UID of %s but got %s", server.Hostname, emptyUID, conf.Servers[server.Hostname])
						}
					} else {
						if server.SSLCert.UID != conf.Servers[server.Hostname] {
							t.Errorf("Expected server %s to have UID of %s but got %s", server.Hostname, server.SSLCert.UID, conf.Servers[server.Hostname])
						}
					}
				}
			}),
		},
	}
	defer server.Close()
	server.Start()

	err = configureCertificates(servers)
	if err != nil {
		t.Errorf("unexpected error posting dynamic certificate configuration: %v", err)
	}
}

func tryListen(network, address string) (l net.Listener, err error) {
	condFunc := func() (bool, error) {
		l, err = net.Listen(network, address)
		if err == nil {
			return true, nil
		}
		if strings.Contains(err.Error(), "bind: address already in use") {
			return false, nil
		}
		return false, err
	}

	backoff := wait.Backoff{
		Duration: 500 * time.Millisecond,
		Factor:   2,
		Steps:    6,
		Cap:      128 * time.Second,
	}
	err = wait.ExponentialBackoff(backoff, condFunc)
	return
}
