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

package controller

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/nginx"
)

func TestIsDynamicConfigurationEnough(t *testing.T) {
	backends := []*ingress.Backend{{
		Name: "fakenamespace-myapp-80",
		Endpoints: []ingress.Endpoint{
			{
				Address: "10.0.0.1",
				Port:    "8080",
			},
			{
				Address: "10.0.0.2",
				Port:    "8080",
			},
		},
	}}

	servers := []*ingress.Server{{
		Hostname: "myapp.fake",
		Locations: []*ingress.Location{
			{
				Path:    "/",
				Backend: "fakenamespace-myapp-80",
			},
		},
		SSLCert: &ingress.SSLCert{
			PemCertKey: "fake-certificate",
		},
	}}

	commonConfig := &ingress.Configuration{
		Backends: backends,
		Servers:  servers,
	}

	n := &NGINXController{
		runningConfig: &ingress.Configuration{
			Backends: backends,
			Servers:  servers,
		},
		cfg: &Configuration{},
	}

	newConfig := commonConfig
	if !n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("When new config is same as the running config it should be deemed as dynamically configurable")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "another-backend-8081"}},
		Servers:  []*ingress.Server{{Hostname: "myapp1.fake"}},
	}
	if n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to not be dynamically configurable when there's more than just backends change")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "a-backend-8080"}},
		Servers:  servers,
	}

	if !n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to be dynamically configurable when only backends change")
	}

	newServers := []*ingress.Server{{
		Hostname: "myapp1.fake",
		Locations: []*ingress.Location{
			{
				Path:    "/",
				Backend: "fakenamespace-myapp-80",
			},
		},
		SSLCert: &ingress.SSLCert{
			PemCertKey: "fake-certificate",
		},
	}}

	newConfig = &ingress.Configuration{
		Backends: backends,
		Servers:  newServers,
	}
	if n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to not be dynamically configurable when dynamic certificates is enabled and a non-certificate field in servers is updated")
	}

	newServers[0].Hostname = "myapp.fake"
	newServers[0].SSLCert.PemCertKey = "new-fake-certificate"

	newConfig = &ingress.Configuration{
		Backends: backends,
		Servers:  newServers,
	}
	if !n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to be dynamically configurable when only SSLCert changes")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "a-backend-8080"}},
		Servers:  newServers,
	}
	if !n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to be dynamically configurable when backend and SSLCert changes")
	}

	if !n.runningConfig.Equal(commonConfig) {
		t.Errorf("Expected running config to not change")
	}

	if !newConfig.Equal(&ingress.Configuration{Backends: []*ingress.Backend{{Name: "a-backend-8080"}}, Servers: newServers}) {
		t.Errorf("Expected new config to not change")
	}
}

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

	n := &NGINXController{
		runningConfig: &ingress.Configuration{},
		cfg:           &Configuration{},
	}

	err = n.configureDynamically(commonConfig)
	if err != nil {
		t.Errorf("unexpected error posting dynamic configuration: %v", err)
	}
	if commonConfig.Backends[0].Endpoints[0].Target != target {
		t.Errorf("unexpected change in the configuration object after configureDynamically invocation")
	}

	resetEndpointStats()
	n.runningConfig.Backends = backends
	err = n.configureDynamically(commonConfig)
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
	n.runningConfig.Servers = servers
	err = n.configureDynamically(commonConfig)
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
	err = n.configureDynamically(commonConfig)
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

func TestNginxHashBucketSize(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 32},
		{1, 32},
		{2, 32},
		{3, 32},
		// ...
		{13, 32},
		{14, 32},
		{15, 64},
		{16, 64},
		// ...
		{45, 64},
		{46, 64},
		{47, 128},
		{48, 128},
		// ...
		// ...
		{109, 128},
		{110, 128},
		{111, 256},
		{112, 256},
		// ...
		{237, 256},
		{238, 256},
		{239, 512},
		{240, 512},
	}

	for _, test := range tests {
		actual := nginxHashBucketSize(test.n)
		if actual != test.expected {
			t.Errorf("Test nginxHashBucketSize(%d): expected %d but returned %d", test.n, test.expected, actual)
		}
	}
}

func TestNextPowerOf2(t *testing.T) {
	// Powers of 2
	actual := nextPowerOf2(2)
	if actual != 2 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 2, actual)
	}
	actual = nextPowerOf2(4)
	if actual != 4 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 4, actual)
	}
	actual = nextPowerOf2(32)
	if actual != 32 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 32, actual)
	}
	actual = nextPowerOf2(256)
	if actual != 256 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 256, actual)
	}

	// Not Powers of 2
	actual = nextPowerOf2(7)
	if actual != 8 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 8, actual)
	}
	actual = nextPowerOf2(9)
	if actual != 16 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 16, actual)
	}
	actual = nextPowerOf2(15)
	if actual != 16 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 16, actual)
	}
	actual = nextPowerOf2(17)
	if actual != 32 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 32, actual)
	}
	actual = nextPowerOf2(250)
	if actual != 256 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 256, actual)
	}

	// Other
	actual = nextPowerOf2(0)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
	actual = nextPowerOf2(-1)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
	actual = nextPowerOf2(-2)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
}

func TestCleanTempNginxCfg(t *testing.T) {
	err := cleanTempNginxCfg()
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err := os.CreateTemp("", tempNginxPattern)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	dur, err := time.ParseDuration("-10m")
	if err != nil {
		t.Fatal(err)
	}

	oldTime := time.Now().Add(dur)
	err = os.Chtimes(tmpfile.Name(), oldTime, oldTime)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err = os.CreateTemp("", tempNginxPattern)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	err = cleanTempNginxCfg()
	if err != nil {
		t.Fatal(err)
	}

	var files []string

	err = filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && os.TempDir() != path {
			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), tempNginxPattern) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Errorf("expected one file but %d were found", len(files))
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
