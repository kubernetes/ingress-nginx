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

package controller

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNginxHashBucketSize(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 32},
		{1, 32},
		{2, 32},
		{3, 32},
		// ...
		{13, 32},
		{14, 32},
		{15, 64},
		{16, 64},
		// ...
		{45, 64},
		{46, 64},
		{47, 128},
		{48, 128},
		// ...
		// ...
		{109, 128},
		{110, 128},
		{111, 256},
		{112, 256},
		// ...
		{237, 256},
		{238, 256},
		{239, 512},
		{240, 512},
	}

	for _, test := range tests {
		actual := nginxHashBucketSize(test.n)
		if actual != test.expected {
			t.Errorf("Test nginxHashBucketSize(%d): expected %d but returned %d", test.n, test.expected, actual)
		}
	}
}

func TestNextPowerOf2(t *testing.T) {
	// Powers of 2
	actual := nextPowerOf2(2)
	if actual != 2 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 2, actual)
	}
	actual = nextPowerOf2(4)
	if actual != 4 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 4, actual)
	}
	actual = nextPowerOf2(32)
	if actual != 32 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 32, actual)
	}
	actual = nextPowerOf2(256)
	if actual != 256 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 256, actual)
	}

	// Not Powers of 2
	actual = nextPowerOf2(7)
	if actual != 8 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 8, actual)
	}
	actual = nextPowerOf2(9)
	if actual != 16 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 16, actual)
	}
	actual = nextPowerOf2(15)
	if actual != 16 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 16, actual)
	}
	actual = nextPowerOf2(17)
	if actual != 32 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 32, actual)
	}
	actual = nextPowerOf2(250)
	if actual != 256 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 256, actual)
	}

	// Other
	actual = nextPowerOf2(0)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
	actual = nextPowerOf2(-1)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
	actual = nextPowerOf2(-2)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
}

func TestCleanTempNginxCfg(t *testing.T) {
	err := cleanTempNginxCfg()
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err := os.CreateTemp("", tempNginxPattern)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	dur, err := time.ParseDuration("-10m")
	if err != nil {
		t.Fatal(err)
	}

	oldTime := time.Now().Add(dur)
	err = os.Chtimes(tmpfile.Name(), oldTime, oldTime)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err = os.CreateTemp("", tempNginxPattern)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	err = cleanTempNginxCfg()
	if err != nil {
		t.Fatal(err)
	}

	var files []string

	err = filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && os.TempDir() != path {
			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), tempNginxPattern) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Errorf("expected one file but %d were found", len(files))
	}
}
