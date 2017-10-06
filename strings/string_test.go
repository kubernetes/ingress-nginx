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

package strings

import (
	"testing"
)

var testDatas = []struct {
	a     string
	slice []string
	er    bool
}{
	{"first", []string{"first", "second"}, true},
	{"FIRST", []string{"first", "second"}, false},
	{"third", []string{"first", "second"}, false},
	{"first", nil, false},

	{"", []string{"first", "second"}, false},
	{"", []string{"first", "second", ""}, true},
	{"", nil, false},
}

func TestStringInSlice(t *testing.T) {
	for _, testData := range testDatas {
		r := StringInSlice(testData.a, testData.slice)
		if r != testData.er {
			t.Errorf("getted result is '%t', but expected is '%t'", r, testData.er)
		}
	}
}
