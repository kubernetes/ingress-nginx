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
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/golang/glog"

	"github.com/mitchellh/mapstructure"
	"k8s.io/kubernetes/pkg/api"
)

// getDNSServers returns the list of nameservers located in the file /etc/resolv.conf
func getDNSServers() []string {
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

// getConfigKeyToStructKeyMap returns a map with the ConfigMapKey as key and the StructName as value.
func getConfigKeyToStructKeyMap() map[string]string {
	keyMap := map[string]string{}
	n := &nginxConfiguration{}
	val := reflect.Indirect(reflect.ValueOf(n))
	for i := 0; i < val.Type().NumField(); i++ {
		fieldSt := val.Type().Field(i)
		configMapKey := strings.Split(fieldSt.Tag.Get("structs"), ",")[0]
		structKey := fieldSt.Name
		keyMap[configMapKey] = structKey
	}
	return keyMap
}

// ReadConfig obtains the configuration defined by the user merged with the defaults.
func (ngx *Manager) ReadConfig(config *api.ConfigMap) nginxConfiguration {
	if len(config.Data) == 0 {
		return newDefaultNginxCfg()
	}

	cfgCM := nginxConfiguration{}
	cfgDefault := newDefaultNginxCfg()

	metadata := &mapstructure.Metadata{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "structs",
		Result:           &cfgCM,
		WeaklyTypedInput: true,
		Metadata:         metadata,
	})

	err = decoder.Decode(config.Data)
	if err != nil {
		glog.Infof("%v", err)
	}

	keyMap := getConfigKeyToStructKeyMap()

	valCM := reflect.Indirect(reflect.ValueOf(cfgCM))

	for _, key := range metadata.Keys {
		fieldName, ok := keyMap[key]
		if !ok {
			continue
		}

		valDefault := reflect.ValueOf(&cfgDefault).Elem().FieldByName(fieldName)

		fieldCM := valCM.FieldByName(fieldName)

		if valDefault.IsValid() {
			valDefault.Set(fieldCM)
		}
	}

	return cfgDefault
}

func (ngx *Manager) needsReload(data *bytes.Buffer) (bool, error) {
	filename := ngx.ConfigFile
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
			glog.Errorf("error computing diff: %s", err)
			return true, nil
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
