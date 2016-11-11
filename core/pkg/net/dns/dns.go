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

package dns

import (
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
)

// GetSystemNameServers returns the list of nameservers located in the file /etc/resolv.conf
func GetSystemNameServers() ([]string, error) {
	var nameservers []string
	file, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nameservers, err
	}

	// Lines of the form "nameserver 1.2.3.4" accumulate.
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
	return nameservers, nil
}
