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

package net

import (
	"net"
	"strings"
)

// IPNet maps string to net.IPNet.
type IPNet map[string]*net.IPNet

// IP maps string to net.IP.
type IP map[string]net.IP

// ParseIPNets parses string slice to IPNet.
func ParseIPNets(specs ...string) (IPNet, IP, error) {
	ipnetset := make(IPNet)
	ipset := make(IP)

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		_, ipnet, err := net.ParseCIDR(spec)
		if err != nil {
			ip := net.ParseIP(spec)
			if ip == nil {
				return nil, nil, err
			}
			i := ip.String()
			ipset[i] = ip
			continue
		}

		k := ipnet.String()
		ipnetset[k] = ipnet
	}

	return ipnetset, ipset, nil
}
