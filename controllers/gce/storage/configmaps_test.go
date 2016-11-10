/*
Copyright 2016 The Kubernetes Authors.

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

package storage

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func TestConfigMapUID(t *testing.T) {
	vault := NewFakeConfigMapVault(api.NamespaceSystem, "ingress-uid")
	uid := ""
	k, exists, err := vault.Get()
	if exists {
		t.Errorf("Got a key from an empyt vault")
	}
	vault.Put(uid)
	k, exists, err = vault.Get()
	if !exists || err != nil {
		t.Errorf("Failed to retrieve value from vault")
	}
	if k != "" {
		t.Errorf("Failed to store empty string as a key in the vault")
	}
	vault.Put("newuid")
	k, exists, err = vault.Get()
	if !exists || err != nil {
		t.Errorf("Failed to retrieve value from vault")
	}
	if k != "newuid" {
		t.Errorf("Failed to modify uid")
	}
	if err := vault.Delete(); err != nil {
		t.Errorf("Failed to delete uid %v", err)
	}
	if uid, exists, _ := vault.Get(); exists {
		t.Errorf("Found uid %v, expected none", uid)
	}
}
