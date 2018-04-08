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
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kubernetes/pkg/util/filesystem"
)

func init() {
	NewFakeFS = newFakeFS
}

// newFakeFS creates an in-memory filesystem with all the required
// paths used by the ingress controller.
// This allows running test without polluting the local machine.
func newFakeFS() (Filesystem, error) {
	fs := filesystem.NewFakeFs()

	for _, directory := range directories {
		err := fs.MkdirAll(directory, 0655)
		if err != nil {
			return nil, err
		}
	}

	for _, file := range files {
		f, err := fs.Create(file)
		if err != nil {
			return nil, err
		}

		_, err = f.Write([]byte(""))
		if err != nil {
			return nil, err
		}

		err = f.Close()
		if err != nil {
			return nil, err
		}
	}

	err := fs.MkdirAll("/proc", 0655)
	if err != nil {
		return nil, err
	}

	for _, assetName := range AssetNames() {
		err := restoreAsset("/", assetName, fs)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// restoreAsset restores an asset under the given directory
func restoreAsset(dir, name string, fs Filesystem) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = fs.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}

	f, err := fs.Create(_filePath(dir, name))
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	//Missing info.Mode()

	err = fs.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

func TestNewFakeFS(t *testing.T) {
	fs, err := NewFakeFS()
	if err != nil {
		t.Fatalf("unexpected error creating filesystem abstraction: %v", err)
	}

	if fs == nil {
		t.Fatal("expected a filesystem but none returned")
	}

	_, err = fs.Stat("/etc/nginx/nginx.conf")
	if err != nil {
		t.Fatalf("unexpected error reading default nginx.conf file: %v", err)
	}
}
