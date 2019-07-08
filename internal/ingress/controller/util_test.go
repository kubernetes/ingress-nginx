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

package controller

import (
	"testing"
)

func TestRlimitMaxNumFiles(t *testing.T) {
	i := rlimitMaxNumFiles()
	if i < 1 {
		t.Errorf("returned %v but expected > 0", i)
	}
}

func TestSysctlSomaxconn(t *testing.T) {
	i := sysctlSomaxconn()
	if i < 511 {
		t.Errorf("returned %v but expected >= 511", i)
	}
}
