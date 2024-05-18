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

func TestSetupSSLProxy(_ *testing.T) {
	// TODO TestSetupSSLProxy
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
