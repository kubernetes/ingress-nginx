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
limitations under the License.
*/

package watch

// DummyFileWatcher noop implementation of a file watcher
type DummyFileWatcher struct{}

// NewDummyFileWatcher creates a FileWatcher using the DummyFileWatcher
func NewDummyFileWatcher(file string, onEvent func()) FileWatcher {
	return DummyFileWatcher{}
}

// Close ends the watch
func (f DummyFileWatcher) Close() error {
	return nil
}
