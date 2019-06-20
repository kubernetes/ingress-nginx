/*
Copyright 2019 The Kubernetes Authors.

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

package sets

import (
	"testing"
)

var (
	testCasesElementMatch = []struct {
		listA    []string
		listB    []string
		expected bool
	}{
		{nil, nil, true},
		{[]string{"1"}, nil, false},
		{[]string{"1"}, []string{"1"}, true},
		{[]string{"1", "2", "1"}, []string{"1", "1", "2"}, true},
		{[]string{"1", "3", "1"}, []string{"1", "1", "2"}, false},
		{[]string{"1", "1"}, []string{"1", "2"}, false},
	}
)

func TestElementsMatch(t *testing.T) {

	for _, testCase := range testCasesElementMatch {
		result := StringElementsMatch(testCase.listA, testCase.listB)
		if result != testCase.expected {
			t.Errorf("expected %v but returned %v (%v - %v)", testCase.expected, result, testCase.listA, testCase.listB)
		}

		sameResult := StringElementsMatch(testCase.listB, testCase.listA)
		if sameResult != testCase.expected {
			t.Errorf("expected %v but returned %v (%v - %v)", testCase.expected, sameResult, testCase.listA, testCase.listB)
		}
	}
}

func BenchmarkStringElementsMatchReflection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range testCasesElementMatch {
			StringElementsMatch(test.listA, test.listB)
		}
	}
}
