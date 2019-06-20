/*
Copyright 2015 The Kubernetes Authors.

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

package collector

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		in  string
		out *basicStatus
	}{
		{`Active connections: 43
server accepts handled requests
 7368 7368 10993
Reading: 0 Writing: 5 Waiting: 38`,
			&basicStatus{43, 7368, 7368, 10993, 0, 5, 38},
		},
		{`Active connections: 0
server accepts handled requests
 1 7 0
Reading: A Writing: B Waiting: 38`,
			&basicStatus{0, 1, 7, 0, 0, 0, 38},
		},
	}

	for _, test := range tests {
		r := parse(test.in)
		if diff := pretty.Compare(r, test.out); diff != "" {
			t.Logf("%v", diff)
			t.Fatalf("expected %v but returned  %v", test.out, r)
		}
	}
}

func TestToint(t *testing.T) {
	tests := []struct {
		in  []string
		pos int
		exp int
	}{
		{[]string{}, 0, 0},
		{[]string{}, 1, 0},
		{[]string{"A"}, 0, 0},
		{[]string{"1"}, 0, 1},
		{[]string{"a", "2"}, 1, 2},
	}

	for _, test := range tests {
		v := toInt(test.in, test.pos)
		if v != test.exp {
			t.Fatalf("expected %v but returned  %v", test.exp, v)
		}
	}
}
