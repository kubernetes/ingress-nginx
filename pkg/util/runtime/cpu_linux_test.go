//go:build linux
// +build linux

/*
Copyright 2018 The Kubernetes Authors.

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

package runtime

import (
	"testing"
)

func TestReadCgroup2StringToInt64Tuple(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedQuota  int64
		expectedPeriod int64
	}{
		{
			name:           "empty string",
			input:          "",
			expectedQuota:  -1,
			expectedPeriod: -1,
		},
		{
			name:           "whitespace only",
			input:          "   ",
			expectedQuota:  -1,
			expectedPeriod: -1,
		},
		{
			name:           "max value",
			input:          "max 100000",
			expectedQuota:  -1,
			expectedPeriod: -1,
		},
		{
			name:           "quota only",
			input:          "50000",
			expectedQuota:  50000,
			expectedPeriod: 100000,
		},
		{
			name:           "quota and period",
			input:          "50000 100000",
			expectedQuota:  50000,
			expectedPeriod: 100000,
		},
		{
			name:           "quota and custom period",
			input:          "100000 200000",
			expectedQuota:  100000,
			expectedPeriod: 200000,
		},
		{
			name:           "invalid quota",
			input:          "invalid 100000",
			expectedQuota:  -1,
			expectedPeriod: -1,
		},
		{
			name:           "invalid period",
			input:          "50000 invalid",
			expectedQuota:  -1,
			expectedPeriod: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			quota, period := readCgroup2StringToInt64Tuple(tc.input)
			if quota != tc.expectedQuota {
				t.Errorf("expected quota %d, got %d", tc.expectedQuota, quota)
			}
			if period != tc.expectedPeriod {
				t.Errorf("expected period %d, got %d", tc.expectedPeriod, period)
			}
		})
	}
}
