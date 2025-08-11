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

package flags

import (
	"os"
	"testing"
	"time"

	"k8s.io/ingress-nginx/internal/ingress/controller"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

func TestNoMandatoryFlag(t *testing.T) {
	_, _, err := ParseFlags()
	if err != nil {
		t.Fatalf("Expected no error but got: %s", err)
	}
}

func TestDefaults(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{
		"cmd",
		"--default-backend-service", "namespace/test",
		"--http-port", "0",
		"--https-port", "0",
	}

	showVersion, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if showVersion {
		t.Fatal("Expected flag \"show-version\" to be false")
	}

	if conf == nil {
		t.Fatal("Expected a controller Configuration")
	}
}

func TestSetupSSLProxy(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		description    string
		validateConfig func(t *testing.T, _ bool, cfg *controller.Configuration)
	}{
		{
			name:        "valid SSL proxy configuration with passthrough enabled",
			args:        []string{"cmd", "--enable-ssl-passthrough", "--ssl-passthrough-proxy-port", "9999"},
			expectError: false,
			description: "Should accept valid SSL proxy port with passthrough enabled",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if !cfg.EnableSSLPassthrough {
					t.Error("Expected EnableSSLPassthrough to be true")
				}
				if cfg.ListenPorts.SSLProxy != 9999 {
					t.Errorf("Expected SSLProxy port to be 9999, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
		{
			name:        "SSL proxy port without explicit passthrough enabling",
			args:        []string{"cmd", "--ssl-passthrough-proxy-port", "8443"},
			expectError: false,
			description: "Should accept SSL proxy port configuration without explicit passthrough enable",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if cfg.ListenPorts.SSLProxy != 8443 {
					t.Errorf("Expected SSLProxy port to be 8443, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
		{
			name:        "SSL proxy with default backend service",
			args:        []string{"cmd", "--enable-ssl-passthrough", "--default-backend-service", "default/backend", "--ssl-passthrough-proxy-port", "9000"},
			expectError: false,
			description: "Should work with default backend service and SSL passthrough",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if !cfg.EnableSSLPassthrough {
					t.Error("Expected EnableSSLPassthrough to be true")
				}
				if cfg.DefaultService != "default/backend" {
					t.Errorf("Expected DefaultService to be 'default/backend', got %s", cfg.DefaultService)
				}
				if cfg.ListenPorts.SSLProxy != 9000 {
					t.Errorf("Expected SSLProxy port to be 9000, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
		{
			name:        "SSL proxy with default SSL certificate",
			args:        []string{"cmd", "--enable-ssl-passthrough", "--default-ssl-certificate", "default/tls-cert", "--ssl-passthrough-proxy-port", "8080"},
			expectError: false,
			description: "Should work with default SSL certificate and passthrough",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if !cfg.EnableSSLPassthrough {
					t.Error("Expected EnableSSLPassthrough to be true")
				}
				if cfg.DefaultSSLCertificate != "default/tls-cert" {
					t.Errorf("Expected DefaultSSLCertificate to be 'default/tls-cert', got %s", cfg.DefaultSSLCertificate)
				}
				if cfg.ListenPorts.SSLProxy != 8080 {
					t.Errorf("Expected SSLProxy port to be 8080, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
		{
			name:        "SSL proxy with chain completion enabled",
			args:        []string{"cmd", "--enable-ssl-passthrough", "--enable-ssl-chain-completion", "--ssl-passthrough-proxy-port", "7443"},
			expectError: false,
			description: "Should work with SSL chain completion and passthrough",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if !cfg.EnableSSLPassthrough {
					t.Error("Expected EnableSSLPassthrough to be true")
				}
				if !config.EnableSSLChainCompletion {
					t.Error("Expected EnableSSLChainCompletion to be true")
				}
				if cfg.ListenPorts.SSLProxy != 7443 {
					t.Errorf("Expected SSLProxy port to be 7443, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
		{
			name:        "SSL proxy with minimal configuration",
			args:        []string{"cmd", "--enable-ssl-passthrough"},
			expectError: false,
			description: "Should work with minimal SSL passthrough configuration using default port",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if !cfg.EnableSSLPassthrough {
					t.Error("Expected EnableSSLPassthrough to be true")
				}
				// Default port should be 442
				if cfg.ListenPorts.SSLProxy != 442 {
					t.Errorf("Expected default SSLProxy port to be 442, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
		{
			name:        "SSL proxy with comprehensive configuration",
			args:        []string{"cmd", "--enable-ssl-passthrough", "--enable-ssl-chain-completion", "--default-ssl-certificate", "kube-system/default-cert", "--default-backend-service", "kube-system/default-backend", "--ssl-passthrough-proxy-port", "10443"},
			expectError: false,
			description: "Should work with comprehensive SSL proxy configuration",
			validateConfig: func(t *testing.T, _ bool, cfg *controller.Configuration) {
				if !cfg.EnableSSLPassthrough {
					t.Error("Expected EnableSSLPassthrough to be true")
				}
				if !config.EnableSSLChainCompletion {
					t.Error("Expected EnableSSLChainCompletion to be true")
				}
				if cfg.DefaultSSLCertificate != "kube-system/default-cert" {
					t.Errorf("Expected DefaultSSLCertificate to be 'kube-system/default-cert', got %s", cfg.DefaultSSLCertificate)
				}
				if cfg.DefaultService != "kube-system/default-backend" {
					t.Errorf("Expected DefaultService to be 'kube-system/default-backend', got %s", cfg.DefaultService)
				}
				if cfg.ListenPorts.SSLProxy != 10443 {
					t.Errorf("Expected SSLProxy port to be 10443, got %d", cfg.ListenPorts.SSLProxy)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetForTesting(func() { t.Fatal("Parsing failed") })

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = tt.args

			showVersion, cfg, err := ParseFlags()
			if tt.expectError && err == nil {
				t.Fatalf("Expected error for %s, but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Expected no error for %s, got: %v", tt.description, err)
			}

			// Run additional validation if provided and no error occurred
			if !tt.expectError && tt.validateConfig != nil {
				tt.validateConfig(t, showVersion, cfg)
			}
		})
	}
}

func TestFlagConflict(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--publish-service", "namespace/test", "--http-port", "0", "--https-port", "0", "--publish-status-address", "1.1.1.1"}

	_, _, err := ParseFlags()
	if err == nil {
		t.Fatalf("Expected an error parsing flags but none returned")
	}
}

func TestMaxmindEdition(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--publish-service", "namespace/test", "--http-port", "0", "--https-port", "0", "--maxmind-license-key", "0000000", "--maxmind-edition-ids", "GeoLite2-City, TestCheck"}

	_, _, err := ParseFlags()
	if err == nil {
		t.Fatalf("Expected an error parsing flags but none returned")
	}
}

func TestMaxmindMirror(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--publish-service", "namespace/test", "--http-port", "0", "--https-port", "0", "--maxmind-mirror", "http://geoip.local", "--maxmind-license-key", "0000000", "--maxmind-edition-ids", "GeoLite2-City, TestCheck"}

	_, _, err := ParseFlags()
	if err == nil {
		t.Fatalf("Expected an error parsing flags but none returned")
	}
}

func TestMaxmindRetryDownload(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--publish-service", "namespace/test", "--http-port", "0", "--https-port", "0", "--maxmind-mirror", "http://127.0.0.1", "--maxmind-license-key", "0000000", "--maxmind-edition-ids", "GeoLite2-City", "--maxmind-retries-timeout", "1s", "--maxmind-retries-count", "3"}

	_, _, err := ParseFlags()
	if err == nil {
		t.Fatalf("Expected an error parsing flags but none returned")
	}
}

func TestDisableLeaderElectionFlag(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--disable-leader-election", "--http-port", "80", "--https-port", "443"}

	_, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if !conf.DisableLeaderElection {
		t.Fatalf("Expected --disable-leader-election and conf.DisableLeaderElection as true, but found: %v", conf.DisableLeaderElection)
	}
}

func TestIfLeaderElectionDisabledFlagIsFalse(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--http-port", "80", "--https-port", "443"}

	_, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if conf.DisableLeaderElection {
		t.Fatalf("Expected --disable-leader-election and conf.DisableLeaderElection as false, but found: %v", conf.DisableLeaderElection)
	}
}

func TestLeaderElectionTTLDefaultValue(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--http-port", "80", "--https-port", "443"}

	_, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if conf.ElectionTTL != 30*time.Second {
		t.Fatalf("Expected --election-ttl and conf.ElectionTTL as 30s, but found: %v", conf.ElectionTTL)
	}
}

func TestLeaderElectionTTLParseValueInSeconds(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--http-port", "80", "--https-port", "443", "--election-ttl", "10s"}

	_, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if conf.ElectionTTL != 10*time.Second {
		t.Fatalf("Expected --election-ttl and conf.ElectionTTL as 10s, but found: %v", conf.ElectionTTL)
	}
}

func TestLeaderElectionTTLParseValueInMinutes(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--http-port", "80", "--https-port", "443", "--election-ttl", "10m"}

	_, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if conf.ElectionTTL != 10*time.Minute {
		t.Fatalf("Expected --election-ttl and conf.ElectionTTL as 10m, but found: %v", conf.ElectionTTL)
	}
}

func TestLeaderElectionTTLParseValueInHours(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--http-port", "80", "--https-port", "443", "--election-ttl", "1h"}

	_, conf, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error parsing default flags: %v", err)
	}

	if conf.ElectionTTL != 1*time.Hour {
		t.Fatalf("Expected --election-ttl and conf.ElectionTTL as 1h, but found: %v", conf.ElectionTTL)
	}
}

func TestMetricsPerUndefinedHost(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--metrics-per-undefined-host=true"}

	_, _, err := ParseFlags()
	if err != nil {
		t.Fatalf("Expected no error but got: %s", err)
	}
}

func TestMetricsPerUndefinedHostWithMetricsPerHostFalse(t *testing.T) {
	ResetForTesting(func() { t.Fatal("Parsing failed") })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--metrics-per-host=false", "--metrics-per-undefined-host=true"}

	_, _, err := ParseFlags()
	if err == nil {
		t.Fatalf("Expected an error parsing flags but none returned")
	}
}
