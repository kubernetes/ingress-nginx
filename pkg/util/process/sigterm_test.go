/*
Copyright 2022 The Kubernetes Authors.

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

package process

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

type FakeProcess struct {
	shouldError bool
	exitCode    int
}

func (f *FakeProcess) Start() {
}

func (f *FakeProcess) Stop() error {
	if f.shouldError {
		return fmt.Errorf("error")
	}
	return nil
}

func (f *FakeProcess) exiterFunc(code int) {
	f.exitCode = code
}

func sendDelayedSignal(delay time.Duration) {
	time.Sleep(delay * time.Second)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}

func TestHandleSigterm(t *testing.T) {
	tests := []struct {
		name        string
		shouldError bool
		delay       int
	}{
		{
			name:        "should exit without error",
			shouldError: false,
		},
		{
			name:        "should exit with error",
			shouldError: true,
			delay:       2,
		},
	}
	for _, tt := range tests {
		process := &FakeProcess{shouldError: tt.shouldError}
		t.Run(tt.name, func(t *testing.T) {
			go sendDelayedSignal(2) // Send a signal after 2 seconds
			HandleSigterm(process, tt.delay, process.exiterFunc)
		})
		if tt.shouldError && process.exitCode != 1 {
			t.Errorf("wrong return, should be 1 and returned %d", process.exitCode)
		}
		if !tt.shouldError && process.exitCode != 0 {
			t.Errorf("wrong return, should be 0 and returned %d", process.exitCode)
		}
	}
}
