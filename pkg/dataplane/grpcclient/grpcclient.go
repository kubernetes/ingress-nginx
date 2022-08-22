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
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	ingressconfig "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
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
	Address     string
	Backendname *ingress.BackendName
	Keepalive   bool
	ErrorCh     chan error
}

// Client defines the client structure
type Client struct {
	connection   *grpc.ClientConn
	Backendname  *ingress.BackendName
	ShutdownFunc func() error
	// EventCh is used between goroutines to publish new events via gRPC
	EventCh chan ingress.EventMessage
	// ConfigCh is used between goroutines to get new configurations published via gRPC
	ConfigCh            chan *ingressconfig.TemplateConfig
	ErrorCh             chan error
	EventClient         ingress.EventServiceClient
	ConfigurationClient ingress.ConfigurationClient
	ctx                 context.Context
}

// NewGRPCClient receives the gRPC configuration and returns the client to be used
func NewGRPCClient(config Config) (*Client, error) {
	var cli Client
	var err error

	cli.ctx = context.TODO()
	cli.ErrorCh = config.ErrorCh
	cli.EventCh = make(chan ingress.EventMessage)
	cli.ConfigCh = make(chan *ingressconfig.TemplateConfig)
	//TODO:  WE WONT USE INSECURE IN PRODUCTION!!!
	cli.connection, err = grpc.Dial(config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	cli.ShutdownFunc = func() error {
		return cli.connection.Close()
	}
	cli.EventClient = ingress.NewEventServiceClient(cli.connection)
	cli.ConfigurationClient = ingress.NewConfigurationClient(cli.connection)

	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")
	cli.Backendname = &ingress.BackendName{
		Name:      podName,
		Namespace: podNs,
	}

	return &cli, nil
}

func (c *Client) Start() {
	go c.EventService()
	go c.ConfigurationService()
}
