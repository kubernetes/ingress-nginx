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
	"os"
	"testing"
	"time"
)

func prepareTimeout() chan bool {
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		timeoutChan <- true
	}()
	return timeoutChan
}

func TestFileWatcher(t *testing.T) {
	file, err := ioutil.TempFile("", "fw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer file.Close()
	defer os.Remove(file.Name())
	count := 0
	events := make(chan bool, 10)
	fw, err := NewFileWatcher(file.Name(), func() {
		count++
		if count != 1 {
			t.Fatalf("expected 1 but returned %v", count)
		}
		events <- true
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer fw.Close()
	timeoutChan := prepareTimeout()
	select {
	case <-events:
		t.Fatalf("expected no events before writing a file")
	case <-timeoutChan:
	}
	ioutil.WriteFile(file.Name(), []byte{}, 0644)
	select {
	case <-events:
	case <-timeoutChan:
		t.Fatalf("expected an event shortly after writing a file")
	}
}
