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
	"testing"
)

func TestReadYamlCase(t *testing.T) {
	itc, err := parseTestCase("suite/0001.yaml")
	if err != nil {
		t.Fatalf("unexpected error reading test case 0001: %v", err)
	}

	if itc == nil {
		t.Fatal("unexpected decoding of test case 0001")
	}

	if itc.ReplicationController != nil {
		t.Fatal("unexpected replication controller in test case 0001")
	}

	if len(itc.Assert) != 1 {
		t.Fatalf("expected 1 tests but %v returned", len(itc.Assert))
	}
}
