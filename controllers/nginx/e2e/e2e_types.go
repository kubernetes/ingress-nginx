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

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
)

// IngressTestCase defines a test case for Ingress
type IngressTestCase struct {
	Name                  string                       `json:"name"`
	Description           string                       `json:"description"`
	Ingress               *extensions.Ingress          `json:"ingress"`
	ReplicationController *apiv1.ReplicationController `json:"replicationController,omitempty"`
	Deployment            *extensions.Deployment       `json:"deployment,omitempty"`
	Service               *apiv1.Service               `json:"service"`
	Assert                []*Assert                    `json:"tests"`
}

// Assert defines a verification over the
type Assert struct {
	Name    string        `json:"name"`
	Request Request       `json:"request"`
	Expect  []*Expect     `json:"expect"`
	Timeout time.Duration `json:"timeout"`
}

// Request defines a HTTP/s request to be executed against an Ingress
type Request struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Query   map[string]interface{} `json:"query"`
	Form    map[string]interface{} `json:"form"`
	Body    interface{}            `json:"body"`
	Headers map[string]string      `json:"headers"`
}

// Expect defines the required conditions that must be true from a request response
type Expect struct {
	Body           []byte            `json:"body"`
	ContentType    string            `json:"contentType"`
	Header         []string          `json:"header"`
	HeaderAndValue map[string]string `json:"headerAndValue"`
	Statuscode     int               `json:"statusCode"`
}
