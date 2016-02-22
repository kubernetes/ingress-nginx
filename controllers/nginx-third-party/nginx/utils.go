/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

// SyncIngress creates a GET request to nginx to indicate that is required to refresh the Ingress rules.
func (ngx *NginxManager) SyncIngress(ingList []interface{}) error {
	encData, _ := json.Marshal(ingList)
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/update-ingress", bytes.NewBuffer(encData))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		glog.Errorf("Error: %v", string(body))
		return fmt.Errorf("nginx status is unhealthy")
	}

	return nil

}

// IsHealthy checks if nginx is running
func (ngx *NginxManager) IsHealthy() error {
	res, err := http.Get("http://127.0.0.1:8080/healthz")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("nginx status is unhealthy")
	}

	return nil
}

// getDnsServers returns the list of nameservers located in the file /etc/resolv.conf
func getDnsServers() []string {
	file, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return []string{}
	}

	// Lines of the form "nameserver 1.2.3.4" accumulate.
	nameservers := []string{}

	lines := strings.Split(string(file), "\n")
	for l := range lines {
		trimmed := strings.TrimSpace(lines[l])
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "nameserver" {
			nameservers = append(nameservers, fields[1:]...)
		}
	}

	glog.V(2).Infof("Nameservers to use: %v", nameservers)
	return nameservers
}

// ReadConfig obtains the configuration defined by the user or returns the default if it does not
// exists or if is not a well formed json object
func (ngx *NginxManager) ReadConfig(data string) (cfg *nginxConfiguration, err error) {
	err = json.Unmarshal([]byte(data), &cfg)
	if err != nil {
		glog.Errorf("Invalid json: %v", err)
		cfg = &nginxConfiguration{}
		err = fmt.Errorf("Invalid custom nginx configuration: %v", err)
		return
	}

	cfg = newDefaultNginxCfg()
	err = fmt.Errorf("No custom nginx configuration. Using defaults")
	return
}

func merge(dst, src map[string]interface{}) map[string]interface{} {
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := toMap(srcVal)
			dstMap, dstMapOk := toMap(dstVal)
			if srcMapOk && dstMapOk {
				srcVal = merge(dstMap, srcMap)
			}
		}
		dst[key] = srcVal
	}

	return dst
}

func toMap(iface interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(iface)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}

		return m, true
	}

	return map[string]interface{}{}, false
}
