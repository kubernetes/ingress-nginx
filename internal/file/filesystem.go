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
	"os"
	"path/filepath"
	"strings"

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
		err := fs.MkdirAll(directory, 0777)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// NewFakeFS creates an in-memory filesystem with all the required
// paths used by the ingress controller.
// This allows running test without polluting the local machine.
func NewFakeFS() (Filesystem, error) {
	osFs := filesystem.DefaultFs{}
	fakeFs := filesystem.NewFakeFs()

	//TODO: find another way to do this
	rootFS := filepath.Clean(fmt.Sprintf("%v/%v", os.Getenv("GOPATH"), "src/k8s.io/ingress-nginx/rootfs"))

	var fileList []string
	err := filepath.Walk(rootFS, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		file := strings.TrimPrefix(path, rootFS)
		if file == "" {
			return nil
		}

		fileList = append(fileList, file)

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, file := range fileList {
		realPath := fmt.Sprintf("%v%v", rootFS, file)

		data, err := osFs.ReadFile(realPath)
		if err != nil {
			return nil, err
		}

		fakeFile, err := fakeFs.Create(file)
		if err != nil {
			return nil, err
		}

		_, err = fakeFile.Write(data)
		if err != nil {
			return nil, err
		}
	}

	return fakeFs, nil
}
