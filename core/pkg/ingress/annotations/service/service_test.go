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

package service

import (
	"encoding/json"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"testing"
)

func fakeService(npa bool, ps bool, expectedP string) *api.Service {
	// fake name for the map of ports
	fakeNpa := NamedPortAnnotation
	if !npa {
		fakeNpa = "fake" + NamedPortAnnotation
	}

	// fake ports
	fakePs, _ := json.Marshal(map[string]string{
		"port1": expectedP,
		"port2": "10211",
	})
	if !ps {
		fakePs, _ = json.Marshal(expectedP)
	}

	// fake service
	return &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ingress",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Annotations: map[string]string{
				fakeNpa: string(fakePs),
			},
			Namespace: api.NamespaceDefault,
			Name:      "named-port-test",
		},
	}
}

func TestGetPortMappingSuccess(t *testing.T) {
	fakeS := fakeService(true, true, "10011")
	port, err := GetPortMapping("port1", fakeS)
	if err != nil {
		t.Errorf("failed to get port with name %s, error: %v", "port1", err)
		return
	}
	if port != 10011 {
		t.Errorf("%s: expected %d but returned %d", "port1", 10011, port)
	}
}

func TestGetPortMappingFailedNamedPortMappingNotExist(t *testing.T) {
	fakeS := fakeService(false, true, "10011")
	_, err := GetPortMapping("port1", fakeS)
	if err == nil {
		t.Errorf("%s: expected error but returned nil", "port1")
	}
}

func TestGetPortMappingFailedPortNotExist(t *testing.T) {
	fakeS := fakeService(true, true, "10011")
	_, err := GetPortMapping("port3", fakeS)
	if err == nil {
		t.Errorf("%s: expected error but returned nil", "port3")
	}
}

func TestGetPortMappingFailedPortInvalid(t *testing.T) {
	fakeS := fakeService(true, true, "s2017")
	_, err := GetPortMapping("port1", fakeS)
	if err == nil {
		t.Errorf("%s: expected error but returned nil", "port1")
	}
}

func TestGetPortMappingFailedApiServiceIsNil(t *testing.T) {
	_, err := GetPortMapping("port1", nil)
	if err == nil {
		t.Errorf("%s: expected error but returned nil", "port1")
	}
}
