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

package file

import (
	"fmt"

	"k8s.io/kubernetes/pkg/util/filesystem"
)

// Filesystem is an interface that we can use to mock various filesystem operations
type Filesystem interface {
	filesystem.Filesystem
}

// NewLocalFS implements Filesystem using same-named functions from "os" and "io/ioutil".
func NewLocalFS() (Filesystem, error) {
	fs := filesystem.DefaultFs{}

	for _, directory := range directories {
		err := fs.MkdirAll(directory, 0655)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// NewFakeFS returns a function to build a fake filesystem
var NewFakeFS = func() (Filesystem, error) {
	return nil, fmt.Errorf("fake filesystem is available only in tests")
}
