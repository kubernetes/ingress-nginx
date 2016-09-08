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

package nginx

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
)

func getConfigNginxBool(data map[string]string) config.Configuration {
	manager := &Manager{}
	configMap := &api.ConfigMap{
		Data: data,
	}
	return manager.ReadConfig(configMap)
}

func TestManagerReadConfigBoolFalse(t *testing.T) {
	configNginx := getConfigNginxBool(map[string]string{
		"hsts-include-subdomains": "false",
		"use-proxy-protocol":      "false",
	})
	if configNginx.HSTSIncludeSubdomains {
		t.Error("Failed to config boolean value (default true) to false")
	}
	if configNginx.UseProxyProtocol {
		t.Error("Failed to config boolean value (default false) to false")
	}
}

func TestManagerReadConfigBoolTrue(t *testing.T) {
	configNginx := getConfigNginxBool(map[string]string{
		"hsts-include-subdomains": "true",
		"use-proxy-protocol":      "true",
	})
	if !configNginx.HSTSIncludeSubdomains {
		t.Error("Failed to config boolean value (default true) to true")
	}
	if !configNginx.UseProxyProtocol {
		t.Error("Failed to config boolean value (default false) to true")
	}
}

func TestManagerReadConfigBoolNothing(t *testing.T) {
	configNginx := getConfigNginxBool(map[string]string{
		"invaild-key": "true",
	})
	if !configNginx.HSTSIncludeSubdomains {
		t.Error("Failed to get default boolean value true")
	}
	if configNginx.UseProxyProtocol {
		t.Error("Failed to get default boolean value false")
	}
}

func TestManagerReadConfigStringSet(t *testing.T) {
	configNginx := getConfigNginxBool(map[string]string{
		"ssl-protocols": "TLSv1.2",
	})
	exp := "TLSv1.2"
	if configNginx.SSLProtocols != exp {
		t.Errorf("Failed to set string value true actual='%s' expected='%s'", configNginx.SSLProtocols, exp)
	}
}

func TestManagerReadConfigStringNothing(t *testing.T) {
	configNginx := getConfigNginxBool(map[string]string{
		"not-existing": "TLSv1.2",
	})
	exp := "10m"
	if configNginx.SSLSessionTimeout != exp {
		t.Errorf("Failed to set string value true actual='%s' expected='%s'", configNginx.SSLSessionTimeout, exp)
	}
}
