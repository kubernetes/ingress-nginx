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

package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/klog/v2"
)

const (
	// FullConfiguration is the request of a full configuration
	FullConfiguration = iota
	// DynamicConfiguration should return only if endpoints changes
	// not used right now.
	DynamicConfiguration
)

// This file should be in the same directory as controller otherwise we will end up with cyclic imports

// ConfigurationServer defines a new gRPC Service responsible for the configuration exchange between control-plane and data-plane
type ConfigurationServer struct {
	ingress.UnimplementedConfigurationServer

	n *NGINXController
}

// GetConfigurations is the service used to get a full configuration during initial config and periodic sync
func (s *ConfigurationServer) GetConfigurations(ctx context.Context, backend *ingress.BackendName) (*ingress.Configurations, error) {

	if err := s.checkNilConfiguration(); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(s.n.templateConfig)
	if err != nil {
		klog.ErrorS(err, "error marshalling config json")
		return nil, fmt.Errorf("failed marshalling config json: %w", err)
	}

	op := ingress.Configurations_FullconfigOp{FullconfigOp: &ingress.Configurations_FullConfiguration{Configuration: payload}}
	return &ingress.Configurations{Op: &op}, nil
}

// WatchConfigurations is the stream of changed configurations
func (s *ConfigurationServer) WatchConfigurations(backend *ingress.BackendName, stream ingress.Configuration_WatchConfigurationsServer) error {

	if err := s.checkNilConfiguration(); err != nil {
		return err
	}
	backendName := fmt.Sprintf("%s/%s", backend.Namespace, backend.Name)

	// TODO: Validate backend name to avoid colisions in map
	s.n.GRPCSubscribers.Lock.Lock()
	s.n.GRPCSubscribers.Clients[backendName] = make(chan int)
	s.n.GRPCSubscribers.Lock.Unlock()

	defer func() {
		s.n.GRPCSubscribers.Lock.Lock()
		close(s.n.GRPCSubscribers.Clients[backendName])
		delete(s.n.GRPCSubscribers.Clients, backendName)
		s.n.GRPCSubscribers.Lock.Unlock()
	}()

	for {
		syncType := <-s.n.GRPCSubscribers.Clients[backendName]
		var err error
		var payload []byte
		if err = stream.Context().Err(); err != nil {
			return fmt.Errorf("context error: %s", err)
		}

		// This is a Switch as we want to predict other configuration types in the future
		switch syncType {
		case FullConfiguration:
			payload, err = json.Marshal(s.n.templateConfig)
			if err != nil {
				klog.ErrorS(err, "error marshalling config json")
				return fmt.Errorf("failed marshalling config json: %w", err)
			}
			op := &ingress.Configurations_FullconfigOp{FullconfigOp: &ingress.Configurations_FullConfiguration{Configuration: payload}}
			err = stream.Send(&ingress.Configurations{Op: op})

		default:
			klog.ErrorS(fmt.Errorf("invalid operation"), "error getting dynamic configuration")
		}

		if err != nil {
			klog.ErrorS(err, "failed to send configuration")
			continue
		}
	}
}

func (s *ConfigurationServer) checkNilConfiguration() error {
	if s.n == nil {
		klog.ErrorS(fmt.Errorf("no config available"), "error generating grpc answer")
		return fmt.Errorf("no configuration is available yet")
	}
	if s.n.templateConfig == nil {
		klog.ErrorS(fmt.Errorf("no config available"), "error generating grpc answer")
		return fmt.Errorf("no configuration is available yet")
	}
	return nil
}
