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
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func TestConfigMapUID(t *testing.T) {
	vault := NewFakeConfigMapVault(api.NamespaceSystem, "ingress-uid")
	//uid := ""
	keyvals, exists, err := vault.Get()
	if exists {
		t.Errorf("Got keyvals from an empty vault")
	}

	// Store empty value for UidDataKey.
	uidmap := map[string]string{UidDataKey: ""}
	vault.Put(uidmap)
	keyvals, exists, err = vault.Get()
	if !exists || err != nil {
		t.Errorf("Failed to retrieve value from vault")
	}
	if val, ok := keyvals[UidDataKey]; !ok {
		t.Errorf("Failed to retried UidDataKey from vault")
	} else if val != "" {
		t.Errorf("Failed to store empty string as a key in the vault")
	}

	// Store actual value in key.
	uidmap[UidDataKey] = "newuid"
	vault.Put(uidmap)
	keyvals, exists, err = vault.Get()
	if !exists || err != nil {
		t.Errorf("Failed to retrieve value from vault")
	} else if val, ok := keyvals[UidDataKey]; !ok {
		t.Errorf("Failed to retried UidDataKey from vault")
	} else if val != "newuid" {
		t.Errorf("Failed to store empty string as a key in the vault")
	}

	// Delete value.
	if err := vault.Delete(); err != nil {
		t.Errorf("Failed to delete uid %v", err)
	}
	if keyvals, exists, _ = vault.Get(); exists {
		t.Errorf("Found uid but expected none after deletion")
	}

	// Ensure Keystore is not wiped on second update.
	uidmap[UidDataKey] = "newuid"
	uidmap[FirewallRuleKey] = "fwrule"
	vault.Put(uidmap)
	keyvals, exists, err = vault.Get()
	if !exists || err != nil || len(keyvals) != 2 {
		t.Errorf("Failed to retrieve value from vault")
	}
	uidmap[UidDataKey] = "newnewuid"
	vault.Put(uidmap)
	keyvals, exists, err = vault.Get()
	if !exists || err != nil || len(keyvals) != 2 {
		t.Errorf("Failed to retrieve value from vault")
	}
	if !reflect.DeepEqual(keyvals, uidmap) {
		t.Errorf("Failed to provide equal maps from vault after a partial update")
	}

}
