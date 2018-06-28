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
	"os/exec"
	"path"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewUDPLogListener(t *testing.T) {
	var count uint64

	fn := func(message []byte) {
		t.Logf("message: %v", string(message))
		atomic.AddUint64(&count, 1)
		time.Sleep(time.Millisecond)
	}

	tmpFile := fmt.Sprintf("/tmp/test-socket-%v", time.Now().Nanosecond())

	l, err := net.Listen("unix", tmpFile)
	if err != nil {
		t.Fatalf("unexpected error creating unix socket: %v", err)
	}
	if l == nil {
		t.Fatalf("expected a listener but none returned")
	}

	defer l.Close()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}

			go handleMessages(conn, fn)
		}
	}()

	conn, _ := net.Dial("unix", tmpFile)
	conn.Write([]byte("message"))
	conn.Close()

	time.Sleep(1 * time.Millisecond)
	if atomic.LoadUint64(&count) != 1 {
		t.Errorf("expected only one message from the UDP listern but %v returned", atomic.LoadUint64(&count))
	}
}

const metricExample = `
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.005"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.01"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.025"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.05"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.1"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.25"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="0.5"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="1"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="2.5"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="5"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="10"} 2
upstream_response_time_seconds_bucket{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200",le="+Inf"} 2
upstream_response_time_seconds_sum{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200"} 0.001
upstream_response_time_seconds_count{host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200"} 2
`

func TestExtractMetrics(t *testing.T) {
	execCommand = mockRootForExecCommand
	defer func() {
		execCommand = exec.Command
	}()

	m := extractMetrics("172.17.0.4:8080", "upstream_response_time_seconds_bucket", metricExample)

	em := []string{
		`host="echoheaders",ingress="http-svc",method="GET",namespace="default",path="/",protocol="HTTP/1.1",service="http-svc",status="200",upstream_ip="172.17.0.4:8080",upstream_name="default-http-svc-80",upstream_status="200"|`,
	}

	if !reflect.DeepEqual(em, m) {
		t.Errorf("unexpected metrics extracted \n%v\n%v\n", em, m)
	}
}

func mockRootForExecCommand(command string, args ...string) *exec.Cmd {
	if strings.Index(command, ".sh") != -1 {
		command = path.Join("../../../../rootfs", command)
	}
	cmd := exec.Command(command, args...)
	return cmd
}
