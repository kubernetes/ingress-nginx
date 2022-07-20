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

package file

import (
	"os"
	"path"
	"path/filepath"
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
	f, err := os.CreateTemp("", "fw")
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
	os.WriteFile(f.Name(), []byte{}, ReadWriteByUser)
	select {
	case <-events:
	case <-timeoutChan:
		t.Fatalf("expected an event shortly after writing a file")
	}
}

func TestFileWatcherWithNestedSymlink(t *testing.T) {
	target1, err := os.CreateTemp("", "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer target1.Close()
	defer os.Remove(target1.Name())
	dir := path.Dir(target1.Name())

	innerLink := path.Join(dir, "innerLink")
	if err = os.Symlink(target1.Name(), innerLink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(innerLink)
	mainLink := path.Join(dir, "mainLink")
	if err = os.Symlink(innerLink, mainLink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(mainLink)

	targetName, err := filepath.EvalSymlinks(mainLink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetName != target1.Name() {
		t.Fatalf("expected symlink to point to %v, not %v", target1.Name(), targetName)
	}

	count := 0
	events := make(chan bool, 10)
	fw, err := NewFileWatcher(mainLink, func() {
		count++
		if count != 1 {
			t.Fatalf("expected 1 but returned %v", count)
		}
		if targetName, err = filepath.EvalSymlinks(mainLink); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		events <- true
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer fw.Close()

	target2, err := os.CreateTemp("", "t2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer target2.Close()
	defer os.Remove(target2.Name())

	if err = os.Remove(innerLink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err = os.Symlink(target2.Name(), innerLink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	timeoutChan := prepareTimeout()
	select {
	case <-events:
	case <-timeoutChan:
		t.Fatalf("expected an event shortly after creating a file and relinking")
	}
	if targetName != target2.Name() {
		t.Fatalf("expected symlink to point to %v, not %v", target2.Name(), targetName)
	}
}
