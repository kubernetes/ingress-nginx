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
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"k8s.io/klog"
)

// SHA1 returns the SHA1 of a file.
func SHA1(filename string) string {
	hasher := sha1.New()
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		klog.Errorf("Error reading file %v", err)
		return ""
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil))
}
