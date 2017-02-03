/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	"k8s.io/ingress/core/pkg/ingress"
)

type (
	configuration struct {
		Backends            []*ingress.Backend
		Servers             []*ingress.Server
		TCPEndpoints        []*ingress.Location
		UDPEndpoints        []*ingress.Location
		PassthroughBackends []*ingress.SSLPassthroughBackend
		Syslog              string `json:"syslog-endpoint"`
	}
)

func newConfig(cfg *ingress.Configuration, data map[string]string) *configuration {
	conf := configuration{
		Backends:            cfg.Backends,
		Servers:             cfg.Servers,
		TCPEndpoints:        cfg.TCPEndpoints,
		UDPEndpoints:        cfg.UPDEndpoints,
		PassthroughBackends: cfg.PassthroughBackends,
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &conf,
		TagName:          "json",
	})
	if err != nil {
		glog.Warningf("error configuring decoder: %v", err)
	}
	if err = decoder.Decode(data); err != nil {
		glog.Warningf("error decoding config: %v", err)
	}
	return &conf
}
