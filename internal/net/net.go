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

package net

import (
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	_net "net"
	"os"
	"os/exec"
)

// IsIPV6 checks if the input contains a valid IPV6 address
func IsIPV6(ip _net.IP) bool {
	return ip != nil && ip.To4() == nil
}

// IsPortAvailable checks if a TCP port is available or not
func IsPortAvailable(p int) bool {
	ln, err := _net.Listen("tcp", fmt.Sprintf(":%v", p))
	if err != nil {
		return false
	}
	defer ln.Close()
	return true
}

// IsIPv6Enabled checks if IPV6 is enabled or not and we have
// at least one configured in the pod
func IsIPv6Enabled() bool {
	cmd := exec.Command("test", "-f", "/proc/net/if_inet6")
	if cmd.Run() != nil {
		return false
	}

	addrs, err := _net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, addr := range addrs {
		ip, _, _ := _net.ParseCIDR(addr.String())
		if IsIPV6(ip) {
			return true
		}
	}

	return false
}

// CheckCapNetBind checks if cap_net_bind_service is set for ingress
func CheckCapNetBind() error {
	processID := os.Getpid()
	set, err := cap.GetPID(processID)
	if err != nil {
		return err
	}
	klog.InfoS("ingress-nginx capability set %v", set.String())

	//check effective
	// Value 10 = NET_BIND_SERVICE
	effective, err := set.GetFlag(0, 10)
	if err != nil {
		return err
	}

	//check permitted
	permitted, err := set.GetFlag(1, 10)
	if err != nil {
		return err
	}
	klog.InfoS("ingress-nginx capabilities: permitted %v effective %v", permitted, effective)
	if !permitted && !effective {
		return errors.New(fmt.Sprintf("ingress-nginx capabilities: permitted %v effective %v", permitted, effective))
	}

	return nil
}
