/*
Copyright 2024 The Kubernetes Authors.

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

package crossplane_test

import (
	"net"
	"os"
	"testing"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
	"github.com/stretchr/testify/require"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/template/crossplane"
	"k8s.io/ingress-nginx/internal/ingress/controller/template/crossplane/extramodules"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

const mockMimeTypes = `
types {
    text/html                                        html htm shtml;
    text/css                                         css;
    text/xml                                         xml;
}
`

var resolvers = []net.IP{net.ParseIP("::1"), net.ParseIP("192.168.20.10")}

// TestTemplate should be a roundtrip test.
// We should initialize the scenarios based on the template configuration
// Then Parse and write a crossplane configuration, and roundtrip/parse back to check
// if the directives matches
// we should ignore line numbers and comments
func TestCrossplaneTemplate(t *testing.T) {
	lua := ngx_crossplane.Lua{}
	options := ngx_crossplane.ParseOptions{
		ErrorOnUnknownDirectives: true,
		StopParsingOnError:       true,
		IgnoreDirectives:         []string{"more_clear_headers", "more_set_headers"}, // TODO: Add more_set_headers
		DirectiveSources: []ngx_crossplane.MatchFunc{
			ngx_crossplane.DefaultDirectivesMatchFunc,
			ngx_crossplane.LuaDirectivesMatchFn,
			extramodules.BrotliMatchFn,
		},
		LexOptions: ngx_crossplane.LexOptions{
			Lexers: []ngx_crossplane.RegisterLexer{lua.RegisterLexer()},
		},
	}
	defaultCertificate := &ingress.SSLCert{
		PemFileName: "bla.crt",
		PemCertKey:  "bla.key",
	}

	mimeFile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	_, err = mimeFile.WriteString(mockMimeTypes)
	require.NoError(t, err)
	require.NoError(t, mimeFile.Close())

	tpl := crossplane.NewTemplate()

	t.Run("it should be able to marshall and unmarshall the default configuration", func(t *testing.T) {
		tplConfig := &config.TemplateConfig{
			Cfg: config.NewDefault(),
		}
		tplConfig.Cfg.DefaultSSLCertificate = defaultCertificate
		tplConfig.Cfg.EnableBrotli = true
		tplConfig.Cfg.HideHeaders = []string{"x-fake-header", "x-another-fake-header"}
		tplConfig.Cfg.Resolver = resolvers
		tplConfig.Cfg.DisableIpv6DNS = true
		tplConfig.Cfg.UseForwardedHeaders = true
		tplConfig.Cfg.LogFormatEscapeNone = true
		tplConfig.Cfg.DisableAccessLog = true
		tplConfig.Cfg.UpstreamKeepaliveConnections = 0

		tpl.SetMimeFile(mimeFile.Name())
		content, err := tpl.Write(tplConfig)
		require.NoError(t, err)

		tmpFile, err := os.CreateTemp("", "")
		require.NoError(t, err)
		_, err = tmpFile.Write(content)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		_, err = ngx_crossplane.Parse(tmpFile.Name(), &options)
		require.NoError(t, err)
	})

	t.Run("it should set the right logging configs", func(t *testing.T) {
		tplConfig := &config.TemplateConfig{
			Cfg: config.NewDefault(),
		}
		tplConfig.Cfg.DefaultSSLCertificate = defaultCertificate
		tplConfig.Cfg.DisableAccessLog = false
		tplConfig.Cfg.HTTPAccessLogPath = "/lalala.log"

		tpl.SetMimeFile(mimeFile.Name())
		content, err := tpl.Write(tplConfig)
		require.NoError(t, err)

		tmpFile, err := os.CreateTemp("", "")
		require.NoError(t, err)
		_, err = tmpFile.Write(content)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		_, err = ngx_crossplane.Parse(tmpFile.Name(), &options)
		require.NoError(t, err)
	})

	t.Run("it should be able to marshall and unmarshall the specified configuration", func(t *testing.T) {
		tplConfig := &config.TemplateConfig{
			Cfg: config.NewDefault(),
		}
		tplConfig.Cfg.DefaultSSLCertificate = defaultCertificate
		tplConfig.Cfg.WorkerCPUAffinity = "0001 0010 0100 1000"
		tplConfig.Cfg.LuaSharedDicts = map[string]int{
			"configuration_data": 10240,
			"certificate_data":   50,
		}
		tplConfig.Cfg.Resolver = resolvers
		tplConfig.Cfg.DisableIpv6DNS = false

		tplConfig.Cfg.UseProxyProtocol = true
		tplConfig.Cfg.ProxyRealIPCIDR = []string{"192.168.0.20", "200.200.200.200"}
		tplConfig.Cfg.LogFormatEscapeJSON = true

		tplConfig.Cfg.GRPCBufferSizeKb = 10 // default 0

		tplConfig.Cfg.HTTP2MaxHeaderSize = "10" // default ""
		tplConfig.Cfg.HTTP2MaxFieldSize = "10"  // default ""
		tplConfig.Cfg.HTTP2MaxRequests = 1      // default 0

		tplConfig.Cfg.UseGzip = true // default false
		tplConfig.Cfg.GzipDisable = "enable"

		tplConfig.Cfg.ShowServerTokens = true // default false

		tplConfig.Cfg.DisableAccessLog = false // TODO: test true
		tplConfig.Cfg.DisableHTTPAccessLog = false
		tplConfig.Cfg.EnableSyslog = true
		tplConfig.Cfg.SyslogHost = "localhost"
		tplConfig.Cfg.SkipAccessLogURLs = []string{"aaa.a", "bbb.b"}
		tplConfig.Cfg.SSLDHParam = "/some/dh.pem"

		// Example: openssl rand 80 | openssl enc -A -base64
		tplConfig.Cfg.SSLSessionTicketKey = "lOj3+7Xe21K9GapKqqPIw/gCQm5S4C2lK8pVne6drEik0QqOQHAw1AaPSMdbAvXx2zZKKPCEG98+g3hzftmrfnePSIvokIIE+hHto3Kj1HQ="

		tplConfig.Cfg.CustomHTTPErrors = []int{1024, 2048}

		tplConfig.Cfg.AllowBackendServerHeader = true                                      // default false
		tplConfig.Cfg.BlockCIDRs = []string{"192.168.0.0/24", " 200.200.0.0/16 "}          // default 0
		tplConfig.Cfg.BlockUserAgents = []string{"someuseragent", " another/user-agent  "} // default 0
		tplConfig.Cfg.BlockReferers = []string{"someref", "  anotherref", "escape\nref"}

		tplConfig.AddHeaders = map[string]string{
			"someheader":    "xpto",
			"anotherheader": "blabla",
		}

		tplConfig.Cfg.EnableBrotli = true
		tplConfig.Cfg.BrotliLevel = 7
		tplConfig.Cfg.BrotliMinLength = 2
		tplConfig.Cfg.BrotliTypes = "application/xml+rss application/atom+xml"

		tplConfig.Cfg.HideHeaders = []string{"x-fake-header", "x-another-fake-header"}
		tplConfig.Cfg.UpstreamKeepaliveConnections = 15

		tplConfig.Cfg.UpstreamKeepaliveConnections = 200
		tplConfig.Cfg.UpstreamKeepaliveTime = "60s"
		tplConfig.Cfg.UpstreamKeepaliveTimeout = 200
		tplConfig.Cfg.UpstreamKeepaliveRequests = 15

		tplConfig.Cfg.CustomHTTPErrors = []int{400, 404, 501}

		tpl = crossplane.NewTemplate()
		tpl.SetMimeFile(mimeFile.Name())
		content, err := tpl.Write(tplConfig)
		require.NoError(t, err)

		tmpFile, err := os.CreateTemp("", "")
		require.NoError(t, err)
		_, err = tmpFile.Write(content)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		_, err = ngx_crossplane.Parse(tmpFile.Name(), &options)
		require.NoError(t, err)
		// require.Equal(t, " ", string(content))
	})
}
