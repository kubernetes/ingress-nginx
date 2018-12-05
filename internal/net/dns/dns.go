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
	"net"
	"strings"

	"k8s.io/klog"
)

var defResolvConf = "/etc/resolv.conf"

// GetSystemNameServers returns the list of nameservers located in the file /etc/resolv.conf
func GetSystemNameServers() ([]net.IP, error) {
	var nameservers []net.IP
	file, err := ioutil.ReadFile(defResolvConf)
	if err != nil {
		return nameservers, err
	}

	// Lines of the form "nameserver 1.2.3.4" accumulate.
	lines := strings.Split(string(file), "\n")
	for l := range lines {
		trimmed := strings.TrimSpace(lines[l])
		if len(trimmed) == 0 || trimmed[0] == '#' || trimmed[0] == ';' {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}
		if fields[0] == "nameserver" {
			ip := net.ParseIP(fields[1])
			if ip != nil {
				nameservers = append(nameservers, ip)
			}
		}
	}

	klog.V(3).Infof("nameservers IP address/es to use: %v", nameservers)
	return nameservers, nil
}
