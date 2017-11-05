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
	"io/ioutil"
	"testing"
)

func TestSHA1(t *testing.T) {
	tests := []struct {
		content []byte
		sha     string
	}{
		{[]byte(""), "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{[]byte("hello world"), "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"},
	}

	for _, test := range tests {
		f, err := ioutil.TempFile("", "sha-test")
		if err != nil {
			t.Fatal(err)
		}
		f.Write(test.content)
		f.Sync()

		sha := SHA1(f.Name())
		f.Close()

		if sha != test.sha {
			t.Fatalf("expected %v but returned %s", test.sha, sha)
		}
	}

	sha := SHA1("")
	if sha != "" {
		t.Fatalf("expected an empty sha but returned %s", sha)
	}
}
