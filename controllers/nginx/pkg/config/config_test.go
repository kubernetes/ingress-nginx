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
		{true, logFormatUpstream, fmt.Sprintf("$proxy_protocol_addr - %s", logFormatUpstream)},
		{false, logFormatUpstream, fmt.Sprintf("$remote_addr - %s", logFormatUpstream)},
		{true, "my-log-format", "$proxy_protocol_addr - my-log-format"},
		{false, "john-log-format", "$remote_addr - john-log-format"},
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
