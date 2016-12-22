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
	_net "net"
	"strings"
)

// IsIPV6 checks if the input contains a valid IPV6 address
func IsIPV6(ip _net.IP) bool {
	if dp := strings.Index(ip.String(), ":"); dp != -1 {
		return true
	}
	return false
}
