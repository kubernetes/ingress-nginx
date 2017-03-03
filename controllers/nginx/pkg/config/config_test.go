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
		{true, "", fmt.Sprintf(logFormatUpstream, "$proxy_protocol_addr")},
		{false, "", fmt.Sprintf(logFormatUpstream, "$remote_addr")},
		{true, "my-log-format", "my-log-format"},
		{false, "john-log-format", "john-log-format"},
	}

	for _, testCase := range testCases {

		result := BuildLogFormatUpstream(testCase.useProxyProtocol, testCase.curLogFormat)

		if result != testCase.expected {
			t.Errorf(" expected %v but return %v", testCase.expected, result)
		}

	}
}
