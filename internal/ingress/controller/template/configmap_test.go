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

package template

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mitchellh/hashstructure"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

func TestFilterErrors(t *testing.T) {
	e := filterErrors([]int{200, 300, 345, 500, 555, 999})
	if len(e) != 4 {
		t.Errorf("expected 4 elements but %v returned", len(e))
	}
}

func TestProxyTimeoutParsing(t *testing.T) {
	testCases := map[string]struct {
		input  string
		expect time.Duration // duration in seconds
	}{
		"valid duration":   {"35s", time.Duration(35) * time.Second},
		"invalid duration": {"3zxs", time.Duration(5) * time.Second},
	}
	for n, tc := range testCases {
		cfg := ReadConfig(map[string]string{"proxy-protocol-header-timeout": tc.input})
		if cfg.ProxyProtocolHeaderTimeout.Seconds() != tc.expect.Seconds() {
			t.Errorf("Testing %v. Expected %v seconds but got %v seconds", n, tc.expect, cfg.ProxyProtocolHeaderTimeout)
		}
	}
}

func TestMergeConfigMapToStruct(t *testing.T) {
	conf := map[string]string{
		"custom-http-errors":            "300,400,demo",
		"proxy-read-timeout":            "1",
		"proxy-send-timeout":            "2",
		"skip-access-log-urls":          "/log,/demo,/test",
		"use-proxy-protocol":            "true",
		"disable-access-log":            "true",
		"access-log-params":             "buffer=4k gzip",
		"access-log-path":               "/var/log/test/access.log",
		"error-log-path":                "/var/log/test/error.log",
		"use-gzip":                      "true",
		"enable-dynamic-tls-records":    "false",
		"gzip-level":                    "9",
		"gzip-types":                    "text/html",
		"proxy-real-ip-cidr":            "1.1.1.1/8,2.2.2.2/24",
		"bind-address":                  "1.1.1.1,2.2.2.2,3.3.3,2001:db8:a0b:12f0::1,3731:54:65fe:2::a7,33:33:33::33::33",
		"worker-shutdown-timeout":       "99s",
		"nginx-status-ipv4-whitelist":   "127.0.0.1,10.0.0.0/24",
		"nginx-status-ipv6-whitelist":   "::1,2001::/16",
		"proxy-add-original-uri-header": "false",
	}
	def := config.NewDefault()
	def.CustomHTTPErrors = []int{300, 400}
	def.DisableAccessLog = true
	def.AccessLogParams = "buffer=4k gzip"
	def.AccessLogPath = "/var/log/test/access.log"
	def.ErrorLogPath = "/var/log/test/error.log"
	def.SkipAccessLogURLs = []string{"/log", "/demo", "/test"}
	def.ProxyReadTimeout = 1
	def.ProxySendTimeout = 2
	def.EnableDynamicTLSRecords = false
	def.UseProxyProtocol = true
	def.GzipLevel = 9
	def.GzipTypes = "text/html"
	def.ProxyRealIPCIDR = []string{"1.1.1.1/8", "2.2.2.2/24"}
	def.BindAddressIpv4 = []string{"1.1.1.1", "2.2.2.2"}
	def.BindAddressIpv6 = []string{"[2001:db8:a0b:12f0::1]", "[3731:54:65fe:2::a7]"}
	def.WorkerShutdownTimeout = "99s"
	def.NginxStatusIpv4Whitelist = []string{"127.0.0.1", "10.0.0.0/24"}
	def.NginxStatusIpv6Whitelist = []string{"::1", "2001::/16"}
	def.ProxyAddOriginalURIHeader = false

	hash, err := hashstructure.Hash(def, &hashstructure.HashOptions{
		TagName: "json",
	})
	if err != nil {
		t.Fatalf("unexpected error obtaining hash: %v", err)
	}
	def.Checksum = fmt.Sprintf("%v", hash)

	to := ReadConfig(conf)
	if diff := pretty.Compare(to, def); diff != "" {
		t.Errorf("unexpected diff: (-got +want)\n%s", diff)
	}

	to = ReadConfig(conf)
	def.BindAddressIpv4 = []string{}
	def.BindAddressIpv6 = []string{}

	if !reflect.DeepEqual(to.BindAddressIpv4, []string{"1.1.1.1", "2.2.2.2"}) {
		t.Errorf("unexpected bindAddressIpv4")
	}

	if !reflect.DeepEqual(to.BindAddressIpv6, []string{"[2001:db8:a0b:12f0::1]", "[3731:54:65fe:2::a7]"}) {
		t.Logf("%v", to.BindAddressIpv6)
		t.Errorf("unexpected bindAddressIpv6")
	}

	def = config.NewDefault()
	hash, err = hashstructure.Hash(def, &hashstructure.HashOptions{
		TagName: "json",
	})
	if err != nil {
		t.Fatalf("unexpected error obtaining hash: %v", err)
	}
	def.Checksum = fmt.Sprintf("%v", hash)

	to = ReadConfig(map[string]string{})
	if diff := pretty.Compare(to, def); diff != "" {
		t.Errorf("unexpected diff: (-got +want)\n%s", diff)
	}

	def = config.NewDefault()
	def.WhitelistSourceRange = []string{"1.1.1.1/32"}

	hash, err = hashstructure.Hash(def, &hashstructure.HashOptions{
		TagName: "json",
	})
	if err != nil {
		t.Fatalf("unexpected error obtaining hash: %v", err)
	}
	def.Checksum = fmt.Sprintf("%v", hash)

	to = ReadConfig(map[string]string{
		"whitelist-source-range": "1.1.1.1/32",
	})

	if diff := pretty.Compare(to, def); diff != "" {
		t.Errorf("unexpected diff: (-got +want)\n%s", diff)
	}
}
