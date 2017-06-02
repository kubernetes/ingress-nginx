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
limitations
*/

package main

import (
	"time"

	"k8s.io/client-go/pkg/api"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// IngressTestCase defines a test case for Ingress
type IngressTestCase struct {
	Name                  string                     `yaml:"name"`
	Description           string                     `yaml:"description"`
	Pod                   *api.Pod                   `yaml:"pod"`
	ReplicationController *api.ReplicationController `yaml:"replicationController"`
	Deployment            *extensions.Deployment     `yaml:"deployment"`
	Assert                []*Assert                  `yaml:"tests"`
}

// Assert defines a verification over the
type Assert struct {
	Name    string        `yaml:"name"`
	Request Request       `yaml:"request"`
	Expect  []*Expect     `yaml:"expect"`
	Timeout time.Duration `yaml:"timeout"`
}

// Request defines a HTTP/s request to be executed against an Ingress
type Request struct {
	Method  string                 `yaml:"method"`
	URL     string                 `yaml:"url"`
	Query   map[string]interface{} `yaml:"query"`
	Form    map[string]interface{} `yaml:"form"`
	Body    interface{}            `yaml:"body"`
	Headers map[string]string      `yaml:"headers"`
}

// Expect defines the required conditions that must be true from a request response
type Expect struct {
	Body           []byte            `yaml:"body"`
	ContentType    string            `yaml:"contentType"`
	Header         []string          `yaml:"header"`
	HeaderAndValue map[string]string `yaml:"headerAndValue"`
	Statuscode     int               `yaml:"statucDode"`
}
