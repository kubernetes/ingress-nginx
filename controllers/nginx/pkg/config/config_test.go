package config

import (
	"fmt"
	"testing"
)

func TestBuildLogFormatUpstream(t *testing.T) {

	testCases := []struct {
		useProxyProtocol bool // use proxy protocol
		expected         string
	}{
		{true, fmt.Sprintf(logFormatUpstream, "$proxy_protocol_addr")},
		{false, fmt.Sprintf(logFormatUpstream, "$remote_addr")},
	}

	for _, testCase := range testCases {

		result := BuildLogFormatUpstream(testCase.useProxyProtocol)

		if result != testCase.expected {
			t.Errorf(" expected %v but return %v", testCase.expected, result)
		}

	}
}
