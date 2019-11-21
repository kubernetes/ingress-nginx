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
	"reflect"
)

type equalFunction func(e1, e2 interface{}) bool

// Compare checks if the parameters are iterable and contains the same elements
func Compare(listA, listB interface{}, eq equalFunction) bool {
	ok := isIterable(listA)
	if !ok {
		return false
	}

	ok = isIterable(listB)
	if !ok {
		return false
	}

	a := reflect.ValueOf(listA)
	b := reflect.ValueOf(listB)

	if a.IsNil() && b.IsNil() {
		return true
	}

	if a.IsNil() != b.IsNil() {
		return false
	}

	if a.Len() != b.Len() {
		return false
	}

	visited := make([]bool, b.Len())

	for i := 0; i < a.Len(); i++ {
		found := false
		for j := 0; j < b.Len(); j++ {
			if visited[j] {
				continue
			}

			if eq(a.Index(i).Interface(), b.Index(j).Interface()) {
				visited[j] = true
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

var compareStrings = func(e1, e2 interface{}) bool {
	s1, ok := e1.(string)
	if !ok {
		return false
	}

	s2, ok := e2.(string)
	if !ok {
		return false
	}

	return s1 == s2
}

// StringElementsMatch compares two string slices and returns if are equals
func StringElementsMatch(a, b []string) bool {
	return Compare(a, b, compareStrings)
}

func isIterable(obj interface{}) bool {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}
