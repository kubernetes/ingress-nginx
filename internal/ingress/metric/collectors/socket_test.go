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

package collectors

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestNewUDPLogListener(t *testing.T) {
	var count uint64

	fn := func(message []byte) {
		atomic.AddUint64(&count, 1)
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
		t.Errorf("expected only one message from the socket listener but %v returned", atomic.LoadUint64(&count))
	}
}

func TestCollector(t *testing.T) {

	buckets := struct {
		TimeBuckets   []float64
		LengthBuckets []float64
		SizeBuckets   []float64
	}{
		prometheus.DefBuckets,
		prometheus.LinearBuckets(10, 10, 10),
		prometheus.ExponentialBuckets(10, 10, 7),
	}

	cases := []struct {
		name            string
		data            []string
		metrics         []string
		wantBefore      string
		removeIngresses []string
		wantAfter       string
	}{
		{
			name: "invalid metric object should not increase prometheus metrics",
			data: []string{`#missing {
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":""
			}`},
			metrics: []string{"nginx_ingress_controller_response_duration_seconds"},
			wantBefore: `

			`,
		},
		{
			name: "valid metric object should update prometheus metrics",
			data: []string{`[{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":""
			}]`},
			metrics: []string{"nginx_ingress_controller_response_duration_seconds"},
			wantBefore: `
				# HELP nginx_ingress_controller_response_duration_seconds The time spent on receiving the response from the upstream server
				# TYPE nginx_ingress_controller_response_duration_seconds histogram
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.005"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.01"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.025"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.05"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.25"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="2.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="10"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="+Inf"} 1
				nginx_ingress_controller_response_duration_seconds_sum{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 200
				nginx_ingress_controller_response_duration_seconds_count{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 1
			`,
			removeIngresses: []string{"test-app-production/web-yml"},
			wantAfter: `
			`,
		},
		{
			name: "valid metric object should update requests metrics",
			data: []string{`[{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":""
			}]`},
			metrics: []string{"nginx_ingress_controller_requests"},
			wantBefore: `
				# HELP nginx_ingress_controller_requests The total number of client requests.
				# TYPE nginx_ingress_controller_requests counter
				nginx_ingress_controller_requests{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 1
			`,
			removeIngresses: []string{"test-app-production/web-yml"},
			wantAfter: `
			`,
		},
		{
			name: "valid metric object with canary information should update prometheus metrics",
			data: []string{`[{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":"test-app-production-test-app-canary-80"
			}]`},
			metrics: []string{"nginx_ingress_controller_response_duration_seconds"},
			wantBefore: `
				# HELP nginx_ingress_controller_response_duration_seconds The time spent on receiving the response from the upstream server
				# TYPE nginx_ingress_controller_response_duration_seconds histogram
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.005"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.01"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.025"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.05"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.25"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="2.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="10"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="+Inf"} 1
				nginx_ingress_controller_response_duration_seconds_sum{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 200
				nginx_ingress_controller_response_duration_seconds_count{canary="test-app-production-test-app-canary-80",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 1
			`,
			removeIngresses: []string{"test-app-production/web-yml"},
			wantAfter: `
			`,
		},

		{
			name: "multiple messages should increase prometheus metric by two",
			data: []string{`[{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":""
			}]`, `[{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-qa",
				"ingress":"web-yml-qa",
				"service":"test-app-qa",
				"canary":""
			}]`, `[{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-qa",
				"ingress":"web-yml-qa",
				"service":"test-app-qa",
				"canary":""
			}]`},
			metrics: []string{"nginx_ingress_controller_response_duration_seconds"},
			wantBefore: `
				# HELP nginx_ingress_controller_response_duration_seconds The time spent on receiving the response from the upstream server
				# TYPE nginx_ingress_controller_response_duration_seconds histogram
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.005"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.01"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.025"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.05"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.25"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="2.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="10"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="+Inf"} 1
				nginx_ingress_controller_response_duration_seconds_sum{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 200
				nginx_ingress_controller_response_duration_seconds_count{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 1
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.005"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.01"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.025"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.05"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.25"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="0.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="2.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="10"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200",le="+Inf"} 2
				nginx_ingress_controller_response_duration_seconds_sum{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200"} 400
				nginx_ingress_controller_response_duration_seconds_count{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml-qa",method="GET",namespace="test-app-qa",path="/admin",service="test-app-qa",status="200"} 2			
			`,
		},

		{
			name: "collector should be able to handle batched metrics correctly",
			data: []string{`[
			{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":200,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":""
			},
			{
				"host":"testshop.com",
				"status":"200",
				"bytesSent":150.0,
				"method":"GET",
				"path":"/admin",
				"requestLength":300.0,
				"requestTime":60.0,
				"upstreamName":"test-upstream",
				"upstreamIP":"1.1.1.1:8080",
				"upstreamResponseTime":100,
				"upstreamStatus":"220",
				"namespace":"test-app-production",
				"ingress":"web-yml",
				"service":"test-app",
				"canary":""
			}]`},
			metrics: []string{"nginx_ingress_controller_response_duration_seconds"},
			wantBefore: `
				# HELP nginx_ingress_controller_response_duration_seconds The time spent on receiving the response from the upstream server
				# TYPE nginx_ingress_controller_response_duration_seconds histogram
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.005"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.01"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.025"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.05"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.25"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="0.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="1"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="2.5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="5"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="10"} 0
				nginx_ingress_controller_response_duration_seconds_bucket{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200",le="+Inf"} 2
				nginx_ingress_controller_response_duration_seconds_sum{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 300
				nginx_ingress_controller_response_duration_seconds_count{canary="",controller_class="ingress",controller_namespace="default",controller_pod="pod",host="testshop.com",ingress="web-yml",method="GET",namespace="test-app-production",path="/admin",service="test-app",status="200"} 2
			`,
			removeIngresses: []string{"test-app-production/web-yml"},
			wantAfter: `
			`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			registry := prometheus.NewPedanticRegistry()

			sc, err := NewSocketCollector("pod", "default", "ingress", true, buckets)
			if err != nil {
				t.Errorf("%v: unexpected error creating new SocketCollector: %v", c.name, err)
			}

			if err := registry.Register(sc); err != nil {
				t.Errorf("registering collector failed: %s", err)
			}

			sc.SetHosts(sets.NewString("testshop.com"))

			for _, d := range c.data {
				sc.handleMessage([]byte(d))
			}

			if err := GatherAndCompare(sc, c.wantBefore, c.metrics, registry); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}

			if len(c.removeIngresses) > 0 {
				sc.RemoveMetrics(c.removeIngresses, registry)
				time.Sleep(1 * time.Second)

				if err := GatherAndCompare(sc, c.wantAfter, c.metrics, registry); err != nil {
					t.Errorf("unexpected collecting result:\n%s", err)
				}
			}

			sc.Stop()

			registry.Unregister(sc)
		})
	}
}
