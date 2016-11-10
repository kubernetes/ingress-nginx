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

package watch

import (
	"io/ioutil"
	"testing"
)

func TestFileWatcher(t *testing.T) {
	file, err := ioutil.TempFile("", "fw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()
	count := 0
	fw, err := NewFileWatcher(file.Name(), func() {
		count++
		if count != 1 {
			t.Fatalf("expected 1 but returned %v", count)
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer fw.Close()
	if count != 0 {
		t.Fatalf("expected 0 but returned %v", count)
	}
	ioutil.WriteFile(file.Name(), []byte{}, 0644)
}
