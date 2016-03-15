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
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

// IsHealthy checks if nginx is running
func (ngx *NginxManager) IsHealthy() error {
	res, err := http.Get("http://127.0.0.1:8080/healthz")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("NGINX is unhealthy")
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

	glog.V(3).Infof("nameservers to use: %v", nameservers)
	return nameservers
}

// ReadConfig obtains the configuration defined by the user or returns the default if it does not
// exists or if is not a well formed json object
func (ngx *NginxManager) ReadConfig(data string) (*nginxConfiguration, error) {
	if data == "" {
		return newDefaultNginxCfg(), nil
	}

	cfg := nginxConfiguration{}
	err := json.Unmarshal([]byte(data), &cfg)
	if err != nil {
		glog.Errorf("invalid json: %v", err)
		return newDefaultNginxCfg(), fmt.Errorf("invalid custom nginx configuration: %v", err)
	}

	return &cfg, nil
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

func checkChanges(filename string, data *bytes.Buffer) (bool, error) {
	in, err := os.Open(filename)
	if err != nil {
		return false, err
	}

	src, err := ioutil.ReadAll(in)
	in.Close()
	if err != nil {
		return false, err
	}

	res := data.Bytes()
	if !bytes.Equal(src, res) {
		err = ioutil.WriteFile(filename, res, 0644)
		if err != nil {
			return false, err
		}

		dData, err := diff(src, res)
		if err != nil {
			return false, fmt.Errorf("computing diff: %s", err)
		}

		if glog.V(2) {
			glog.Infof("NGINX configuration diff a/%s b/%s\n", filename, filename)
			glog.Infof("%v", string(dData))
		}

		return len(dData) > 0, nil
	}

	return false, nil
}

func diff(b1, b2 []byte) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return
}
