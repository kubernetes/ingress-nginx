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

package config

import (
	"fmt"
	"testing"
)

func TestBuildLogFormatUpstream(t *testing.T) {

	testCases := []struct {
		useProxyProtocol bool // use proxy protocol
		curLogFormat     string
		expected         string
	}{
		{true, logFormatUpstream, fmt.Sprintf(logFormatUpstream, "$the_real_ip")},
		{false, logFormatUpstream, fmt.Sprintf(logFormatUpstream, "$the_real_ip")},
		{true, "my-log-format", "my-log-format"},
		{false, "john-log-format", "john-log-format"},
	}

	for _, testCase := range testCases {
		cfg := NewDefault()
		cfg.UseProxyProtocol = testCase.useProxyProtocol
		cfg.LogFormatUpstream = testCase.curLogFormat
		result := cfg.BuildLogFormatUpstream()
		if result != testCase.expected {
			t.Errorf(" expected %v but return %v", testCase.expected, result)
		}
	}
}
