/*
Copyright 2022 The Kubernetes Authors.

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

package grpcclient

import (
	"fmt"

	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/klog/v2"
)

func (c *Client) HealthService() {
	request := healthgrpc.HealthCheckRequest{Service: ""}
	stream, err := c.HealthClient.Watch(c.ctx, &request)
	if err != nil {
		klog.Errorf("error creating event client: %s", err)
		return
	}

	for {
		health, err := stream.Recv()
		if err != nil {
			c.grpcErrCh <- fmt.Errorf("error getting health: %w", err)
			return
		}
		if health.Status != healthgrpc.HealthCheckResponse_SERVING {
			c.grpcErrCh <- fmt.Errorf("error on health: %s", health.String())
			return
		}
	}
}
