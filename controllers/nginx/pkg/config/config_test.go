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
		{true, logFormatUpstream, fmt.Sprintf(logFormatUpstream, "$proxy_protocol_addr")},
		{false, logFormatUpstream, fmt.Sprintf(logFormatUpstream, "$remote_addr")},
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
