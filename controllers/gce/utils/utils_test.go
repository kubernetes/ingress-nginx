/*
Copyright 2015 The Kubernetes Authors.

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

package utils

import (
	"testing"
)

const (
	ClusterId = "0123456789abcdef"
)

func TestNEGName(t *testing.T) {
	testCases := []struct {
		namespace string
		name      string
		port      string
		expect    string
	}{
		// simple case
		{
			"namespace",
			"name",
			"80",
			"k8s-0123456789abcdef-namespace-name-80-1e047e33",
		},
		// 63 characters
		{
			"0123456789a",
			"0123456789",
			"0123456789",
			"k8s-0123456789abcdef-0123456789a-0123456789-0123456789-49c9993e",
		},
		// long namespace
		{
			"01234567890123456789012345678901234567890123456789",
			"0",
			"0",
			"k8s-0123456789abcdef-01234567890123456789012345678-0-0-44255b67",
		},
		// long name
		{
			"0",
			"01234567890123456789012345678901234567890123456789",
			"0",
			"k8s-0123456789abcdef--012345678901234567890123456789-0-8a22658c",
		},
		// long name and namespace
		{
			"01234567890123456789012345678901234567890123456789",
			"01234567890123456789012345678901234567890123456789",
			"0",
			"k8s-0123456789abcdef-01234567890123-0123456789012345-0-525cce3d",
		},
		// long name, namespace and port
		{
			"01234567890123456789012345678901234567890123456789",
			"0123456789012345678901234567890123456789",
			"012345678901234567890123456789",
			"k8s-0123456789abcdef-012345678901-01234567890-01234567-71877a60",
		},
	}

	namer := NewNamer(ClusterId, "")
	for _, tc := range testCases {
		res := namer.NEGName(tc.namespace, tc.name, tc.port)
		if len(res) > 63 {
			t.Errorf("NEG name is longer than 63 charaters: %v", len(res))
		}
		if res != tc.expect {
			t.Errorf("Expect NEG name to be %q, but got %q.", tc.expect, res)
		}
	}
}
