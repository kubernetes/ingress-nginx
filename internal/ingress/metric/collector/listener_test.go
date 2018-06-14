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
	"sync/atomic"
	"testing"
	"time"
)

func TestNewUDPLogListener(t *testing.T) {
	port := freeUDPPort()

	var count uint64

	fn := func(message []byte) {
		t.Logf("message: %v", string(message))
		atomic.AddUint64(&count, 1)
	}

	t.Logf("UDP Port: %v", port)

	l, err := newUDPListener(port)
	if err != nil {
		t.Errorf("unexpected error creating UDP listener: %v", err)
	}
	if l == nil {
		t.Errorf("expected a listener but none returned")
	}

	go handleMessages(l, fn)

	conn, _ := net.Dial("udp", fmt.Sprintf(":%v", port))
	conn.Write([]byte("message"))
	conn.Close()

	time.Sleep(1 * time.Millisecond)
	if count != 1 {
		t.Errorf("expected only one message from the UDP listern but %v returned", count)
	}
}

func freeUDPPort() int {
	l, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return 0
	}

	if err := l.Close(); err != nil {
		return 0
	}

	return l.LocalAddr().(*net.UDPAddr).Port
}
