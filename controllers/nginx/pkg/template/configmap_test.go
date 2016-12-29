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
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"

	"k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/net/dns"
)

func TestFilterErrors(t *testing.T) {
	e := filterErrors([]int{200, 300, 345, 500, 555, 999})
	if len(e) != 4 {
		t.Errorf("expected 4 elements but %v returned", len(e))
	}
}

func TestMergeConfigMapToStruct(t *testing.T) {
	conf := &api.ConfigMap{
		Data: map[string]string{
			"custom-http-errors":         "300,400",
			"proxy-read-timeout":         "1",
			"proxy-send-timeout":         "2",
			"skip-access-log-urls":       "/log,/demo,/test",
			"use-proxy-protocol":         "true",
			"use-gzip":                   "true",
			"enable-dynamic-tls-records": "false",
			"gzip-types":                 "text/html",
		},
	}
	def := config.NewDefault()
	def.CustomHTTPErrors = []int{300, 400}
	def.SkipAccessLogURLs = []string{"/log", "/demo", "/test"}
	def.ProxyReadTimeout = 1
	def.ProxySendTimeout = 2
	def.EnableDynamicTLSRecords = false
	def.UseProxyProtocol = true
	def.GzipTypes = "text/html"

	h, _ := dns.GetSystemNameServers()
	def.Resolver = h

	to := ReadConfig(conf)
	if !reflect.DeepEqual(def, to) {
		t.Errorf("expected %v but retuned %v", def, to)
	}
}
