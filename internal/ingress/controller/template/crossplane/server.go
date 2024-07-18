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

package crossplane

import (
	"fmt"
	"strings"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"

	"k8s.io/ingress-nginx/pkg/apis/ingress"
	utilingress "k8s.io/ingress-nginx/pkg/util/ingress"
	"k8s.io/utils/ptr"
)

func (c *Template) buildServerDirective(server *ingress.Server) *ngx_crossplane.Directive {
	cfg := c.tplConfig.Cfg
	serverName := buildServerName(server.Hostname)
	serverBlock := ngx_crossplane.Directives{
		buildDirective("server_name", serverName, server.Aliases),
		buildDirective("http2", cfg.UseHTTP2),
		buildDirective("set", "$proxy_upstream_name", "-"),
		buildDirective("ssl_certificate_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_certificate.lua"),
	}

	serverBlock = append(serverBlock, buildListener(c.tplConfig, server.Hostname)...)
	serverBlock = append(serverBlock, c.buildBlockers()...)

	if server.Hostname == "_" {
		serverBlock = append(serverBlock, buildDirective("ssl_reject_handshake", cfg.SSLRejectHandshake))
	}

	if server.CertificateAuth.MatchCN != "" {
		matchCNBlock := buildBlockDirective("if",
			[]string{"$ssl_client_s_dn", "!~", server.CertificateAuth.MatchCN},
			ngx_crossplane.Directives{
				buildDirective("return", "403", "client certificate unauthorized"),
			})
		serverBlock = append(serverBlock, matchCNBlock)
	}

	if server.AuthTLSError != "" {
		serverBlock = append(serverBlock, buildDirective("return", 403))
	} else {
		serverBlock = append(serverBlock, c.buildCertificateDirectives(server)...)
		serverBlock = append(serverBlock, buildCustomErrorLocationsPerServer(server, c.tplConfig.EnableMetrics)...)
		serverBlock = append(serverBlock, buildMirrorLocationDirective(server.Locations)...)

		// The other locations should come here!
		serverBlock = append(serverBlock, c.buildServerLocations(server, server.Locations)...)
	}

	// "/healthz" location
	if server.Hostname == "_" {
		dirs := ngx_crossplane.Directives{
			buildDirective("access_log", "off"),
			buildDirective("return", "200"),
		}
		if cfg.EnableOpentelemetry {
			dirs = append(dirs, buildDirective("opentelemetry", "off"))
		}
		healthLocation := buildBlockDirective("location",
			[]string{c.tplConfig.HealthzURI}, dirs)
		serverBlock = append(serverBlock, healthLocation)

		// "/nginx_status" location
		statusLocationDirs := ngx_crossplane.Directives{}
		if cfg.EnableOpentelemetry {
			statusLocationDirs = append(statusLocationDirs, buildDirective("opentelemetry", "off"))
		}

		for _, v := range c.tplConfig.NginxStatusIpv4Whitelist {
			statusLocationDirs = append(statusLocationDirs, buildDirective("allow", v))
		}

		if c.tplConfig.IsIPV6Enabled {
			for _, v := range c.tplConfig.NginxStatusIpv6Whitelist {
				statusLocationDirs = append(statusLocationDirs, buildDirective("allow", v))
			}
		}
		statusLocationDirs = append(statusLocationDirs,
			buildDirective("deny", "all"),
			buildDirective("access_log", "off"),
			buildDirective("stub_status", "on"))

		// End of "nginx_status" location

		serverBlock = append(serverBlock, buildBlockDirective("location", []string{"/nginx_status"}, statusLocationDirs))
	}

	// DO NOT MOVE! THIS IS THE END DIRECTIVE OF SERVERS
	serverBlock = append(serverBlock, buildCustomErrorLocation("upstream-default-backend", cfg.CustomHTTPErrors, c.tplConfig.EnableMetrics)...)

	return &ngx_crossplane.Directive{
		Directive: "server",
		Block:     serverBlock,
	}
}

func (c *Template) buildCertificateDirectives(server *ingress.Server) ngx_crossplane.Directives {
	certDirectives := make(ngx_crossplane.Directives, 0)

	if server.CertificateAuth.CAFileName != "" {
		certAuth := server.CertificateAuth
		certDirectives = append(certDirectives,
			buildDirective("ssl_client_certificate", certAuth.CAFileName),
			buildDirective("ssl_verify_client", certAuth.VerifyClient),
			buildDirective("ssl_verify_depth", certAuth.ValidationDepth))
		if certAuth.CRLFileName != "" {
			certDirectives = append(certDirectives, buildDirective("ssl_crl", certAuth.CRLFileName))
		}
		if certAuth.ErrorPage != "" {
			certDirectives = append(certDirectives, buildDirective("error_page", "495", "496", "=", certAuth.ErrorPage))
		}
	}

	prxSSL := server.ProxySSL
	if prxSSL.CAFileName != "" {
		certDirectives = append(certDirectives, buildDirective("proxy_ssl_trusted_certificate", prxSSL.CAFileName),
			buildDirective("proxy_ssl_ciphers", prxSSL.Ciphers),
			buildDirective("proxy_ssl_protocols", strings.Split(prxSSL.Protocols, " ")),
			buildDirective("proxy_ssl_verify", prxSSL.Verify),
			buildDirective("proxy_ssl_verify_depth", prxSSL.VerifyDepth),
		)
		if prxSSL.ProxySSLName != "" {
			certDirectives = append(certDirectives,
				buildDirective("proxy_ssl_name", prxSSL.ProxySSLName),
				buildDirective("proxy_ssl_server_name", prxSSL.ProxySSLServerName))
		}
	}
	if prxSSL.PemFileName != "" {
		certDirectives = append(certDirectives,
			buildDirective("proxy_ssl_certificate", prxSSL.PemFileName),
			buildDirective("proxy_ssl_certificate_key", prxSSL.PemFileName))
	}
	if server.SSLCiphers != "" {
		certDirectives = append(certDirectives, buildDirective("ssl_ciphers", server.SSLCiphers))
	}

	if server.SSLPreferServerCiphers != "" {
		certDirectives = append(certDirectives, buildDirective("ssl_prefer_server_ciphers", server.SSLPreferServerCiphers))
	}

	return certDirectives
}

// buildRedirectServer builds the server blocks for redirections
func (c *Template) buildRedirectServer(server *utilingress.Redirect) *ngx_crossplane.Directive {
	serverBlock := ngx_crossplane.Directives{
		buildDirective("server_name", server.From),
		buildDirective("ssl_certificate_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_certificate.lua"),
		buildDirective("set_by_lua_file", "$redirect_to", "/etc/nginx/lua/nginx/ngx_srv_redirect.lua", server.To),
	}
	serverBlock = append(serverBlock, buildListener(c.tplConfig, server.From)...)
	serverBlock = append(serverBlock, c.buildBlockers()...)
	serverBlock = append(serverBlock, buildDirective("return", c.tplConfig.Cfg.HTTPRedirectCode, "$redirect_to"))

	return &ngx_crossplane.Directive{
		Directive: "server",
		Block:     serverBlock,
	}
}

// buildDefaultBackend builds the default catch all server
func (c *Template) buildDefaultBackend() *ngx_crossplane.Directive {
	var reusePort *string
	if c.tplConfig.Cfg.ReusePort {
		reusePort = ptr.To("reuseport")
	}
	serverBlock := ngx_crossplane.Directives{
		buildDirective("listen", c.tplConfig.ListenPorts.Default, "default_server", reusePort, fmt.Sprintf("backlog=%d", c.tplConfig.BacklogSize)),
	}
	if c.tplConfig.IsIPV6Enabled {
		serverBlock = append(serverBlock, buildDirective(
			"listen",
			fmt.Sprintf("[::]:%d", c.tplConfig.ListenPorts.Default),
			"default_server", reusePort,
			fmt.Sprintf("backlog=%d", c.tplConfig.BacklogSize),
		))
	}
	serverBlock = append(serverBlock,
		buildDirective("set", "$proxy_upstream_name", "internal"),
		buildDirective("access_log", "off"),
		buildBlockDirective("location", []string{"/"}, ngx_crossplane.Directives{
			buildDirective("return", "404"),
		}))

	return &ngx_crossplane.Directive{
		Directive: "server",
		Block:     serverBlock,
	}
}

func (c *Template) buildHealthAndStatsServer() *ngx_crossplane.Directive {
	serverBlock := ngx_crossplane.Directives{
		buildDirective("listen", fmt.Sprintf("127.0.0.1:%d", c.tplConfig.StatusPort)),
		buildDirective("set", "$proxy_upstream_name", "internal"),
		buildDirective("keepalive_timeout", "0"),
		buildDirective("gzip", "off"),
		buildDirective("access_log", "off"),
		buildBlockDirective(
			"location",
			[]string{c.tplConfig.HealthzURI}, ngx_crossplane.Directives{
				buildDirective("return", "200"),
			}),
		buildBlockDirective(
			"location",
			[]string{"/is-dynamic-lb-initialized"}, ngx_crossplane.Directives{
				buildDirective("content_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_is_dynamic_lb_initialized.lua"),
			}),
		buildBlockDirective(
			"location",
			[]string{c.tplConfig.StatusPath}, ngx_crossplane.Directives{
				buildDirective("stub_status", "on"),
			}),
		buildBlockDirective(
			"location",
			[]string{"/configuration"}, ngx_crossplane.Directives{
				buildDirective("client_max_body_size", luaConfigurationRequestBodySize(&c.tplConfig.Cfg)),
				buildDirective("client_body_buffer_size", luaConfigurationRequestBodySize(&c.tplConfig.Cfg)),
				buildDirective("proxy_buffering", "off"),
				buildDirective("content_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_configuration.lua"),
			}),
		buildBlockDirective(
			"location",
			[]string{"/"}, ngx_crossplane.Directives{
				buildDirective("return", "404"),
			}),
	}
	if c.tplConfig.Cfg.EnableOpentelemetry {
		serverBlock = append(serverBlock, buildDirective("opentelemetry", "off"))
	}

	return &ngx_crossplane.Directive{
		Directive: "server",
		Block:     serverBlock,
	}
}

func (c *Template) buildBlockers() ngx_crossplane.Directives {
	blockers := make(ngx_crossplane.Directives, 0)
	if len(c.tplConfig.Cfg.BlockUserAgents) > 0 {
		uaDirectives := buildBlockDirective("if", []string{"$block_ua"}, ngx_crossplane.Directives{
			buildDirective("return", "403"),
		})
		blockers = append(blockers, uaDirectives)
	}

	if len(c.tplConfig.Cfg.BlockReferers) > 0 {
		refDirectives := buildBlockDirective("if", []string{"$block_ref"}, ngx_crossplane.Directives{
			buildDirective("return", "403"),
		})
		blockers = append(blockers, refDirectives)
	}
	return blockers
}
