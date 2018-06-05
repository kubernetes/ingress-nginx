/*
Copyright 2018 The Kubernetes Authors.

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

package collector

import (
	"fmt"
	"net"
)

const packetSize = 1024 * 65

// newUDPListener creates a new UDP listener used to process messages
//  from the NGINX log phase containing information about a request
func newUDPListener(port int) (*net.UDPConn, error) {
	service := fmt.Sprintf("127.0.0.1:%v", port)

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		return nil, err
	}

	return net.ListenUDP("udp", udpAddr)
}

// handleMessages process packets received in an UDP connection
func handleMessages(conn *net.UDPConn, fn func([]byte)) {
	msg := make([]byte, packetSize)

	for {
		s, _, err := conn.ReadFrom(msg[0:])
		if err != nil {
			continue
		}

		fn(msg[0:s])
	}
}
