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

	"k8s.io/ingress-nginx/internal/file"
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
	f, err := ioutil.TempFile("", "fw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())
	count := 0
	events := make(chan bool, 10)
	fw, err := NewFileWatcher(f.Name(), func() {
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
	ioutil.WriteFile(f.Name(), []byte{}, file.ReadWriteByUser)
	select {
	case <-events:
	case <-timeoutChan:
		t.Fatalf("expected an event shortly after writing a file")
	}
}
