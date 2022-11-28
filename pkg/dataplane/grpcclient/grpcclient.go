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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ingressconfig "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/klog/v2"
)

/* TODO for this client:
- Implement LB: https://github.com/grpc/grpc-go/blob/master/examples/features/load_balancing/client/main.go
- Allow usage of xDS: https://github.com/grpc/grpc-go/blob/master/examples/features/xds/server/main.go
- Check keepalive usage https://github.com/grpc/grpc-go/tree/master/examples/features/keepalive
- Allow usage of gzip or snappy: https://github.com/grpc/grpc-go/tree/master/examples/features/compression
- Authenticate using mTLS: https://github.com/grpc/grpc-go/tree/master/examples/features/encryption/mTLS
- Implement health service: https://github.com/grpc/grpc-go/tree/master/examples/features/health
- Implement retry config: https://github.com/grpc/grpc-go/tree/master/examples/features/retry
*/

// Config defines the gRPC configuration to be used
type Config struct {
	// Backendname *ingress.BackendName // TODO: This is not used yet
	ErrorCh   chan error
	GRPCErrCh chan error
	Options   GRPCDialOptions
}

type GRPCDialOptions struct {
	Address     string
	Keepalive   bool   // TODO: Not used yet
	RetryPolicy string // TODO: Not used yet, will be a struct
}

// Client defines the client structure
type Client struct {
	Options      GRPCDialOptions
	connection   *grpc.ClientConn
	Backendname  *ingress.BackendName
	ShutdownFunc func() error
	// EventCh is used between goroutines to publish new events via gRPC
	EventCh chan ingress.EventMessage
	// ConfigCh is used between goroutines to get new configurations published via gRPC
	ConfigCh            chan *ingressconfig.TemplateConfig
	ErrorCh             chan error
	grpcErrCh           chan error
	EventClient         ingress.EventServiceClient
	ConfigurationClient ingress.ConfigurationClient
	ctx                 context.Context
}

var (
	retryPolicy = `{
		"methodConfig": [{
		  "name": [{"service": "ingressv1.Configuration"}],
		  "name": [{"service": "ingressv1.EventService"}],
		  "waitForReady": true,
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": "1s",
			  "MaxBackoff": "2s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`

	// TODO: Turn configurable
	kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
)

// NewGRPCClient receives the gRPC configuration and returns the client to be used
func NewGRPCClient(config Config) (*Client, error) {
	var cli Client

	cli.ctx = context.TODO()
	cli.ErrorCh = config.ErrorCh
	cli.EventCh = make(chan ingress.EventMessage)
	cli.ConfigCh = make(chan *ingressconfig.TemplateConfig)
	cli.grpcErrCh = config.GRPCErrCh

	cli.ShutdownFunc = func() error {
		return cli.connection.Close()
	}

	cli.Options = config.Options

	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")
	cli.Backendname = &ingress.BackendName{
		Name:      podName,
		Namespace: podNs,
	}

	return &cli, nil
}

func (c *Client) Start() {

	// Because we may start the dataplane before the control plane, Kubernetes service may not
	// be ready or return a DNS error when dp tries to connect to cp.
	// Usually this is fast, but we can/should wait until the control plane is ready
	// to move forward

	var newConn *grpc.ClientConn
	err := wait.Poll(1*time.Second, 30*time.Second, func() (bool, error) {
		var err error
		newConn, err = grpc.Dial(
			c.Options.Address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(retryPolicy),
			grpc.WithKeepaliveParams(kacp),
		)
		if err != nil {
			klog.Warningf("error while trying to connect to controller, retrying: %s", err)
			return false, nil
		}

		c.EventClient = ingress.NewEventServiceClient(newConn)
		c.ConfigurationClient = ingress.NewConfigurationClient(newConn)
		cfg, err := c.ConfigurationClient.GetConfigurations(c.ctx, c.Backendname)
		if err != nil {
			klog.Warningf("failed to get initial configuration: %s", err)
			return false, nil
		}

		switch op := cfg.Op.(type) {
		case *ingress.Configurations_FullconfigOp:
			var configBackend *config.TemplateConfig
			if err := json.Unmarshal(op.FullconfigOp.Configuration, &configBackend); err != nil {
				klog.Fatalf("error unmarshalling config: %s", err)
			}
			if len(configBackend.Servers) < 1 || len(configBackend.Backends) < 1 {
				klog.Warning("controller not ready, retrying in 5s")
				return false, nil
			}
			klog.Info("controller become ready, moving forward")

		default:
			klog.Fatalf("controller returned an invalid operation: %v", op)
		}

		return true, nil

	})
	if err != nil {
		c.grpcErrCh <- fmt.Errorf("gRPC server didn't became ready, will retry: %s", err)
	}

	go c.EventService()
	go c.ConfigurationService()
}
