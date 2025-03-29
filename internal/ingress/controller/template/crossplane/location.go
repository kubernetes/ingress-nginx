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
	"sort"
	"strconv"
	"strings"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

func buildMirrorLocationDirective(locs []*ingress.Location) ngx_crossplane.Directives {
	mirrorDirectives := make(ngx_crossplane.Directives, 0)

	mapped := sets.Set[string]{}

	for _, loc := range locs {
		if loc.Mirror.Source == "" || loc.Mirror.Target == "" || loc.Mirror.Host == "" {
			continue
		}

		if mapped.Has(loc.Mirror.Source) {
			continue
		}

		mapped.Insert(loc.Mirror.Source)
		mirrorDirectives = append(mirrorDirectives, buildBlockDirective("location",
			[]string{"=", loc.Mirror.Source},
			ngx_crossplane.Directives{
				buildDirective("internal"),
				buildDirective("proxy_set_header", "Host", loc.Mirror.Host),
				buildDirective("proxy_pass", loc.Mirror.Target),
			}))
	}
	return mirrorDirectives
}

// buildCustomErrorLocationsPerServer is a utility function which will collect all
// custom error codes for all locations of a server block, deduplicates them,
// and returns a set which is unique by default-upstream and error code. It returns an array
// of errorLocations, each of which contain the upstream name and a list of
// error codes for that given upstream, so that sufficiently unique
// @custom error location blocks can be created in the template
func buildCustomErrorLocationsPerServer(server *ingress.Server, enableMetrics bool) ngx_crossplane.Directives {
	type errorLocation struct {
		UpstreamName string
		Codes        []int
	}

	codesMap := make(map[string]map[int]bool)
	for _, loc := range server.Locations {
		backendUpstream := loc.DefaultBackendUpstreamName

		var dedupedCodes map[int]bool
		if existingMap, ok := codesMap[backendUpstream]; ok {
			dedupedCodes = existingMap
		} else {
			dedupedCodes = make(map[int]bool)
		}

		for _, code := range loc.CustomHTTPErrors {
			dedupedCodes[code] = true
		}
		codesMap[backendUpstream] = dedupedCodes
	}

	errorLocations := []errorLocation{}

	for upstream, dedupedCodes := range codesMap {
		codesForUpstream := []int{}
		for code := range dedupedCodes {
			codesForUpstream = append(codesForUpstream, code)
		}
		sort.Ints(codesForUpstream)
		errorLocations = append(errorLocations, errorLocation{
			UpstreamName: upstream,
			Codes:        codesForUpstream,
		})
	}

	sort.Slice(errorLocations, func(i, j int) bool {
		return errorLocations[i].UpstreamName < errorLocations[j].UpstreamName
	})

	errorLocationsDirectives := make(ngx_crossplane.Directives, 0)
	for i := range errorLocations {
		errorLocationsDirectives = append(errorLocationsDirectives, buildCustomErrorLocation(errorLocations[i].UpstreamName, errorLocations[i].Codes, enableMetrics)...)
	}
	return errorLocationsDirectives
}

func buildCustomErrorLocation(upstreamName string, errorCodes []int, enableMetrics bool) ngx_crossplane.Directives {
	directives := make(ngx_crossplane.Directives, len(errorCodes))
	for i := range errorCodes {
		locationDirectives := ngx_crossplane.Directives{
			buildDirective("internal"),
			buildDirective("proxy_intercept_errors", "off"),
			buildDirective("proxy_set_header", "X-Code", errorCodes[i]),
			buildDirective("proxy_set_header", "X-Format", "$http_accept"),
			buildDirective("proxy_set_header", "X-Original-URI", "$request_uri"),
			buildDirective("proxy_set_header", "X-Namespace", "$namespace"),
			buildDirective("proxy_set_header", "X-Ingress-Name", "$ingress_name"),
			buildDirective("proxy_set_header", "X-Service-Name", "$service_name"),
			buildDirective("proxy_set_header", "X-Service-Port", "$service_port"),
			buildDirective("proxy_set_header", "X-Request-ID", "$req_id"),
			buildDirective("proxy_set_header", "X-Forwarded-For", "$remote_addr"),
			buildDirective("proxy_set_header", "Host", "$best_http_host"),
			buildDirective("set", "$proxy_upstream_name", upstreamName),
			buildDirective("rewrite", "(.*)", "/", "break"),
			buildDirective("proxy_pass", "http://upstream_balancer"),
		}

		if enableMetrics {
			locationDirectives = append(locationDirectives, buildDirective("log_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_log.lua"))
		}
		locationName := fmt.Sprintf("@custom_%s_%d", upstreamName, errorCodes[i])
		directives[i] = buildBlockDirective("location", []string{locationName}, locationDirectives)
	}

	return directives
}

type locationCfg struct {
	pathLocation      []string
	authPath          string
	externalAuth      *externalAuth
	proxySetHeader    string
	applyGlobalAuth   bool
	applyAuthUpstream bool
}

func (c *Template) buildServerLocations(server *ingress.Server, locations []*ingress.Location) ngx_crossplane.Directives {
	serverLocations := make(ngx_crossplane.Directives, 0)

	cfg := c.tplConfig.Cfg
	enforceRegexModifier := false
	needsRewrite := func(loc *ingress.Location) bool {
		return loc.Rewrite.Target != "" &&
			loc.Rewrite.Target != loc.Path
	}

	for _, location := range locations {
		if needsRewrite(location) || location.Rewrite.UseRegex {
			enforceRegexModifier = true
			break
		}
	}

	for _, location := range locations {
		locationConfig := locationCfg{
			pathLocation:      buildLocation(location, enforceRegexModifier),
			proxySetHeader:    getProxySetHeader(location),
			authPath:          buildAuthLocation(location, cfg.GlobalExternalAuth.URL),
			applyGlobalAuth:   shouldApplyGlobalAuth(location, cfg.GlobalExternalAuth.URL),
			applyAuthUpstream: shouldApplyAuthUpstream(location, &cfg),
			externalAuth:      &externalAuth{},
		}

		if location.Rewrite.AppRoot != "" {
			serverLocations = append(serverLocations,
				buildBlockDirective("if", []string{"$uri", "=", "/"},
					ngx_crossplane.Directives{
						buildDirective("return", "302", fmt.Sprintf("$scheme://$http_host%s", location.Rewrite.AppRoot)),
					}))
		}

		if locationConfig.applyGlobalAuth {
			locationConfig.externalAuth = buildExternalAuth(cfg.GlobalExternalAuth)
		} else {
			locationConfig.externalAuth = buildExternalAuth(location.ExternalAuth)
		}
		if locationConfig.authPath != "" {
			serverLocations = append(serverLocations, c.buildAuthLocation(server, location, locationConfig))
		}
		if location.Denied == nil && locationConfig.externalAuth != nil && locationConfig.externalAuth.SigninURL != "" {
			directives := ngx_crossplane.Directives{
				buildDirective("internal"),
				buildDirective("add_header", "Set-Cookie", "$auth_cookie"),
			}
			if location.CorsConfig.CorsEnabled {
				directives = append(directives, buildCorsDirectives(&location.CorsConfig)...)
			}
			directives = append(directives,
				buildDirective("return",
					"302",
					buildAuthSignURL(locationConfig.externalAuth.SigninURL, locationConfig.externalAuth.SigninURLRedirectParam)))

			serverLocations = append(serverLocations, buildBlockDirective("location",
				[]string{buildAuthSignURLLocation(location.Path, locationConfig.externalAuth.SigninURL)}, directives))
		}
		serverLocations = append(serverLocations, c.buildLocation(server, location, locationConfig))
	}
	return serverLocations
}

func (c *Template) buildLocation(server *ingress.Server,
	location *ingress.Location, locationConfig locationCfg,
) *ngx_crossplane.Directive {
	ing := getIngressInformation(location.Ingress, server.Hostname, location.IngressPath)
	cfg := c.tplConfig
	locationDirectives := ngx_crossplane.Directives{
		buildDirective("set", "$namespace", ing.Namespace),
		buildDirective("set", "$ingress_name", ing.Rule),
		buildDirective("set", "$service_name", ing.Service),
		buildDirective("set", "$service_port", ing.ServicePort),
		buildDirective("set", "$balancer_ewma_score", "-1"),
		buildDirective("set", "$proxy_upstream_name", location.Backend),
		buildDirective("set", "$proxy_host", "$proxy_upstream_name"),
		buildDirective("set", "$pass_access_scheme", "$scheme"),
		buildDirective("set", "$best_http_host", "$http_host"),
		buildDirective("set", "$pass_port", "$pass_server_port"),
		buildDirective("set", "$proxy_alternative_upstream_name", ""),
		buildDirective("set", "$location_path", strings.ReplaceAll(ing.Path, `$`, `${literal_dollar}`)),
	}

	locationDirectives = append(locationDirectives, locationConfigForLua(location, c.tplConfig)...)
	locationDirectives = append(locationDirectives, buildCertificateDirectives(location)...)

	if cfg.Cfg.UseProxyProtocol {
		locationDirectives = append(locationDirectives,
			buildDirective("set", "$pass_server_port", "$proxy_protocol_server_port"))
	} else {
		locationDirectives = append(locationDirectives,
			buildDirective("set", "$pass_server_port", "$server_port"))
	}

	locationDirectives = append(locationDirectives,
		buildOpentelemetryForLocationDirectives(cfg.Cfg.EnableOpentelemetry, cfg.Cfg.OpentelemetryTrustIncomingSpan, location)...)

	locationDirectives = append(locationDirectives,
		buildDirective("rewrite_by_lua_file", "/etc/nginx/lua/nginx/ngx_rewrite.lua"),
		buildDirective("header_filter_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_srv_hdr_filter.lua"),
		buildDirective("log_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_log_block.lua"),
		buildDirective("rewrite_log", location.Logs.Rewrite),
		// buildDirective("http2_push_preload", location.HTTP2PushPreload), // This directive is deprecated, keeping out of new crossplane
		buildDirective("port_in_redirect", location.UsePortInRedirects))

	if location.Mirror.Source != "" {
		locationDirectives = append(locationDirectives,
			buildDirective("mirror", location.Mirror.Source),
			buildDirective("mirror_request_body", location.Mirror.RequestBody),
		)
	}

	if !location.Logs.Access {
		locationDirectives = append(locationDirectives,
			buildDirective("access_log", "off"),
		)
	}
	if location.Denied != nil {
		locationDirectives = append(locationDirectives,
			buildDirectiveWithComment("return", fmt.Sprintf("Location denied. Reason: %s", *location.Denied), "503"))
	} else {
		locationDirectives = append(locationDirectives, c.buildAllowedLocation(server, location, locationConfig)...)
	}

	return buildBlockDirective("location", locationConfig.pathLocation, locationDirectives)
}

func (c *Template) buildAllowedLocation(server *ingress.Server, location *ingress.Location, locationConfig locationCfg) ngx_crossplane.Directives {
	dir := make(ngx_crossplane.Directives, 0)
	proxySetHeader := locationConfig.proxySetHeader
	for _, ip := range location.Denylist.CIDR {
		dir = append(dir, buildDirective("deny", ip))
	}
	if len(location.Allowlist.CIDR) > 0 {
		for _, ip := range location.Allowlist.CIDR {
			dir = append(dir, buildDirective("allow", ip))
		}
		dir = append(dir, buildDirective("deny", "all"))
	}

	if location.CorsConfig.CorsEnabled {
		dir = append(dir, buildCorsDirectives(&location.CorsConfig)...)
	}

	if !isLocationInLocationList(location, c.tplConfig.Cfg.NoAuthLocations) {
		dir = append(dir, buildAuthLocationConfig(location, locationConfig)...)
	}

	dir = append(dir, buildRateLimit(location)...)

	if isValidByteSize(location.Proxy.BodySize, true) {
		dir = append(dir, buildDirective("client_max_body_size", location.Proxy.BodySize))
	}
	if isValidByteSize(location.ClientBodyBufferSize, false) {
		dir = append(dir, buildDirective("client_body_buffer_size", location.ClientBodyBufferSize))
	}

	if location.UpstreamVhost != "" {
		dir = append(dir, buildDirective(proxySetHeader, "Host", location.UpstreamVhost))
	} else {
		dir = append(dir, buildDirective(proxySetHeader, "Host", "$best_http_host"))
	}

	if server.CertificateAuth.CAFileName != "" {
		dir = append(dir,
			buildDirective(proxySetHeader, "ssl-client-verify", "$ssl_client_verify"),
			buildDirective(proxySetHeader, "ssl-client-subject-dn", "$ssl_client_s_dn"),
			buildDirective(proxySetHeader, "ssl-client-issuer-dn", "$ssl_client_i_dn"),
		)

		if server.CertificateAuth.PassCertToUpstream {
			dir = append(dir, buildDirective(proxySetHeader, "ssl-client-cert", "$ssl_client_escaped_cert"))
		}
	}

	dir = append(dir,
		buildDirective(proxySetHeader, "Upgrade", "$http_upgrade"),
		buildDirective(proxySetHeader, "X-Request-ID", "$req_id"),
		buildDirective(proxySetHeader, "X-Real-IP", "$remote_addr"),
		buildDirective(proxySetHeader, "X-Forwarded-Host", "$best_http_host"),
		buildDirective(proxySetHeader, "X-Forwarded-Port", "$pass_port"),
		buildDirective(proxySetHeader, "X-Forwarded-Proto", "$pass_access_scheme"),
		buildDirective(proxySetHeader, "X-Forwarded-Scheme", "$pass_access_scheme"),
		buildDirective(proxySetHeader, "X-Real-IP", "$remote_addr"),
		buildDirective(proxySetHeader, "X-Scheme", "$pass_access_scheme"),
		buildDirective(proxySetHeader, "X-Original-Forwarded-For",
			fmt.Sprintf("$http_%s", strings.ToLower(strings.ReplaceAll(c.tplConfig.Cfg.ForwardedForHeader, "-", "_")))),
		buildDirectiveWithComment(proxySetHeader,
			"mitigate HTTProxy Vulnerability - https://www.nginx.com/blog/mitigating-the-httpoxy-vulnerability-with-nginx/", "Proxy", ""),
		buildDirective("proxy_connect_timeout", seconds(location.Proxy.ConnectTimeout)),
		buildDirective("proxy_read_timeout", seconds(location.Proxy.ReadTimeout)),
		buildDirective("proxy_send_timeout", seconds(location.Proxy.SendTimeout)),
		buildDirective("proxy_buffering", location.Proxy.ProxyBuffering),
		buildDirective("proxy_buffer_size", location.Proxy.BufferSize),
		buildDirective("proxy_buffers", location.Proxy.BuffersNumber, location.Proxy.BufferSize),
		buildDirective("proxy_request_buffering", location.Proxy.RequestBuffering),
		buildDirective("proxy_busy_buffers_size", location.Proxy.BusyBuffersSize),
		buildDirective("proxy_http_version", location.Proxy.ProxyHTTPVersion),
		buildDirective("proxy_cookie_domain", strings.Split(location.Proxy.CookieDomain, " ")),
		buildDirective("proxy_cookie_path", strings.Split(location.Proxy.CookiePath, " ")),
		buildDirective("proxy_next_upstream_timeout", location.Proxy.NextUpstreamTimeout),
		buildDirective("proxy_next_upstream_tries", location.Proxy.NextUpstreamTries),
		buildDirective("proxy_next_upstream", buildNextUpstream(location.Proxy.NextUpstream, c.tplConfig.Cfg.RetryNonIdempotent)),
	)

	if isValidByteSize(location.Proxy.ProxyMaxTempFileSize, true) {
		dir = append(dir, buildDirective("proxy_max_temp_file_size", location.Proxy.ProxyMaxTempFileSize))
	}

	if c.tplConfig.Cfg.UseForwardedHeaders && c.tplConfig.Cfg.ComputeFullForwardedFor {
		dir = append(dir, buildDirective(proxySetHeader, "X-Forwarded-For", "$full_x_forwarded_for"))
	} else {
		dir = append(dir, buildDirective(proxySetHeader, "X-Forwarded-For", "$remote_addr"))
	}

	if c.tplConfig.Cfg.ProxyAddOriginalURIHeader {
		dir = append(dir, buildDirective(proxySetHeader, "X-Original-URI", "$request_uri"))
	}

	if location.Connection.Enabled {
		dir = append(dir, buildDirective(proxySetHeader, "Connection", location.Connection.Header))
	} else {
		dir = append(dir, buildDirective(proxySetHeader, "Connection", "$connection_upgrade"))
	}

	for k, v := range c.tplConfig.ProxySetHeaders {
		dir = append(dir, buildDirective(proxySetHeader, k, v))
	}

	for k, v := range location.CustomHeaders.Headers {
		dir = append(dir, buildDirective("more_set_headers", fmt.Sprintf("%s: %s", k, strings.ReplaceAll(v, `$`, `${literal_dollar}`))))
	}

	if strings.HasPrefix(location.Backend, "custom-default-backend-") {
		dir = append(dir,
			buildDirective("proxy_set_header", "X-Code", "503"),
			buildDirective("proxy_set_header", "X-Format", "$http_accept"),
			buildDirective("proxy_set_header", "X-Namespace", "$namespace"),
			buildDirective("proxy_set_header", "X-Ingress-Name", "$ingress_name"),
			buildDirective("proxy_set_header", "X-Service-Name", "$service_name"),
			buildDirective("proxy_set_header", "X-Service-Port", "$service_port"),
			buildDirective("proxy_set_header", "X-Request-ID", "$req_id"),
		)
	}

	if location.Satisfy != "" {
		dir = append(dir, buildDirective("satisfy", location.Satisfy))
	}

	if location.Redirect.Relative {
		dir = append(dir, buildDirective("absolute_redirect", false))
	}

	if len(location.CustomHTTPErrors) > 0 && !location.DisableProxyInterceptErrors {
		dir = append(dir, buildDirective("proxy_intercept_errors", "on"))
	}

	for _, errorcode := range location.CustomHTTPErrors {
		dir = append(dir, buildDirective(
			"error_page",
			errorcode, "=",
			fmt.Sprintf("@custom_%s_%d", location.DefaultBackendUpstreamName, errorcode)),
		)
	}

	switch location.BackendProtocol {
	case "GRPC", "GRPCS":
		dir = append(dir,
			buildDirective("grpc_connect_timeout", seconds(location.Proxy.ConnectTimeout)),
			buildDirective("grpc_send_timeout", seconds(location.Proxy.SendTimeout)),
			buildDirective("grpc_read_timeout", seconds(location.Proxy.ReadTimeout)),
		)
	case "FCGI":
		dir = append(dir, buildDirective("include", "/etc/nginx/fastcgi_params"))
		if location.FastCGI.Index != "" {
			dir = append(dir, buildDirective("fastcgi_index", location.FastCGI.Index))
		}
		for k, v := range location.FastCGI.Params {
			dir = append(dir, buildDirective("fastcgi_param", k, v))
		}
	}

	if location.Redirect.URL != "" {
		dir = append(dir, buildDirective("return", location.Redirect.Code, location.Redirect.URL))
	}

	dir = append(dir, buildProxyPass(c.tplConfig.Backends, location)...)

	if location.Proxy.ProxyRedirectFrom == "default" || location.Proxy.ProxyRedirectFrom == "off" {
		dir = append(dir, buildDirective("proxy_redirect", location.Proxy.ProxyRedirectFrom))
	} else if location.Proxy.ProxyRedirectTo != "off" {
		dir = append(dir, buildDirective("proxy_redirect", location.Proxy.ProxyRedirectFrom, location.Proxy.ProxyRedirectTo))
	}

	return dir
}

func buildCertificateDirectives(location *ingress.Location) ngx_crossplane.Directives {
	cert := make(ngx_crossplane.Directives, 0)
	if location.ProxySSL.CAFileName != "" {
		cert = append(cert,
			buildDirectiveWithComment(
				"proxy_ssl_trusted_certificate",
				fmt.Sprintf("#PEM sha: %s", location.ProxySSL.CASHA),
				location.ProxySSL.CAFileName,
			),
			buildDirective("proxy_ssl_ciphers", location.ProxySSL.Ciphers),
			buildDirective("proxy_ssl_protocols", strings.Split(location.ProxySSL.Protocols, " ")),
			buildDirective("proxy_ssl_verify", location.ProxySSL.Verify),
			buildDirective("proxy_ssl_verify_depth", location.ProxySSL.VerifyDepth),
		)
	}
	if location.ProxySSL.ProxySSLName != "" {
		cert = append(cert, buildDirective("proxy_ssl_name", location.ProxySSL.ProxySSLName))
	}
	if location.ProxySSL.ProxySSLServerName != "" {
		cert = append(cert, buildDirective("proxy_ssl_server_name", location.ProxySSL.ProxySSLServerName))
	}
	if location.ProxySSL.PemFileName != "" {
		cert = append(cert,
			buildDirective("proxy_ssl_certificate", location.ProxySSL.PemFileName),
			buildDirective("proxy_ssl_certificate_key", location.ProxySSL.PemFileName),
		)
	}
	return cert
}

type ingressInformation struct {
	Namespace   string
	Path        string
	Rule        string
	Service     string
	ServicePort string
	Annotations map[string]string
}

func getIngressInformation(ing *ingress.Ingress, hostname, ingressPath string) *ingressInformation {
	if ing == nil {
		return &ingressInformation{}
	}

	info := &ingressInformation{
		Namespace:   ing.GetNamespace(),
		Rule:        ing.GetName(),
		Annotations: ing.Annotations,
		Path:        ingressPath,
	}

	if ingressPath == "" {
		ingressPath = "/"
		info.Path = "/"
	}

	if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
		info.Service = ing.Spec.DefaultBackend.Service.Name
		if ing.Spec.DefaultBackend.Service.Port.Number > 0 {
			info.ServicePort = strconv.Itoa(int(ing.Spec.DefaultBackend.Service.Port.Number))
		} else {
			info.ServicePort = ing.Spec.DefaultBackend.Service.Port.Name
		}
	}

	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		if hostname != "_" && rule.Host == "" {
			continue
		}

		host := "_"
		if rule.Host != "" {
			host = rule.Host
		}

		if hostname != host {
			continue
		}

		for _, rPath := range rule.HTTP.Paths {
			if ingressPath != rPath.Path {
				continue
			}

			if rPath.Backend.Service == nil {
				continue
			}

			if info.Service != "" && rPath.Backend.Service.Name == "" {
				// empty rule. Only contains a Path and PathType
				return info
			}

			info.Service = rPath.Backend.Service.Name
			if rPath.Backend.Service.Port.Number > 0 {
				info.ServicePort = strconv.Itoa(int(rPath.Backend.Service.Port.Number))
			} else {
				info.ServicePort = rPath.Backend.Service.Port.Name
			}

			return info
		}
	}

	return info
}

func buildOpentelemetryForLocationDirectives(isOTEnabled, isOTTrustSet bool, location *ingress.Location) ngx_crossplane.Directives {
	isOTEnabledInLoc := location.Opentelemetry.Enabled
	isOTSetInLoc := location.Opentelemetry.Set
	directives := make(ngx_crossplane.Directives, 0)
	if isOTEnabled {
		if isOTSetInLoc && !isOTEnabledInLoc {
			return ngx_crossplane.Directives{
				buildDirective("opentelemetry", "off"),
			}
		}
	} else if !isOTSetInLoc || !isOTEnabledInLoc {
		return directives
	}

	if location != nil {
		directives = append(directives,
			buildDirective("opentelemetry", "on"),
			buildDirective("opentelemetry_propagate"),
		)
		if location.Opentelemetry.OperationName != "" {
			directives = append(directives,
				buildDirective("opentelemetry_operation_name", location.Opentelemetry.OperationName))
		}

		if (!isOTTrustSet && !location.Opentelemetry.TrustSet) ||
			(location.Opentelemetry.TrustSet && !location.Opentelemetry.TrustEnabled) {
			directives = append(directives,
				buildDirective("opentelemetry_trust_incoming_spans", "off"),
			)
		} else {
			directives = append(directives,
				buildDirective("opentelemetry_trust_incoming_spans", "on"),
			)
		}
	}

	return directives
}

// buildRateLimit produces an array of limit_req to be used inside the Path of
// Ingress rules. The order: connections by IP first, then RPS, and RPM last.
func buildRateLimit(loc *ingress.Location) ngx_crossplane.Directives {
	limits := make(ngx_crossplane.Directives, 0)

	if loc.RateLimit.Connections.Limit > 0 {
		limits = append(limits, buildDirective("limit_conn", loc.RateLimit.Connections.Name, loc.RateLimit.Connections.Limit))
	}

	if loc.RateLimit.RPS.Limit > 0 {
		limits = append(limits,
			buildDirective(
				"limit_req",
				fmt.Sprintf("zone=%s", loc.RateLimit.RPS.Name),
				fmt.Sprintf("burst=%d", loc.RateLimit.RPS.Burst),
				"nodelay",
			),
		)
	}

	if loc.RateLimit.RPM.Limit > 0 {
		limits = append(limits,
			buildDirective(
				"limit_req",
				fmt.Sprintf("zone=%s", loc.RateLimit.RPM.Name),
				fmt.Sprintf("burst=%d", loc.RateLimit.RPM.Burst),
				"nodelay",
			),
		)
	}

	if loc.RateLimit.LimitRateAfter > 0 {
		limits = append(limits,
			buildDirective(
				"limit_rate_after",
				fmt.Sprintf("%dk", loc.RateLimit.LimitRateAfter),
			),
		)
	}

	if loc.RateLimit.LimitRate > 0 {
		limits = append(limits,
			buildDirective(
				"limit_rate",
				fmt.Sprintf("%dk", loc.RateLimit.LimitRate),
			),
		)
	}

	return limits
}

// locationConfigForLua formats some location specific configuration into Lua table represented as string
func locationConfigForLua(location *ingress.Location, all *config.TemplateConfig) ngx_crossplane.Directives {
	/* Lua expects the following vars
		force_ssl_redirect = string_to_bool(ngx.var.force_ssl_redirect),
	    ssl_redirect = string_to_bool(ngx.var.ssl_redirect),
	    force_no_ssl_redirect = string_to_bool(ngx.var.force_no_ssl_redirect),
	    preserve_trailing_slash = string_to_bool(ngx.var.preserve_trailing_slash),
	    use_port_in_redirects = string_to_bool(ngx.var.use_port_in_redirects),
	*/

	return ngx_crossplane.Directives{
		buildDirective("set", "$force_ssl_redirect", strconv.FormatBool(location.Rewrite.ForceSSLRedirect)),
		buildDirective("set", "$ssl_redirect", strconv.FormatBool(location.Rewrite.SSLRedirect)),
		buildDirective("set", "$force_no_ssl_redirect", strconv.FormatBool(isLocationInLocationList(location, all.Cfg.NoTLSRedirectLocations))),
		buildDirective("set", "$preserve_trailing_slash", strconv.FormatBool(location.Rewrite.PreserveTrailingSlash)),
		buildDirective("set", "$use_port_in_redirects", strconv.FormatBool(location.UsePortInRedirects)),
	}
}

func isLocationInLocationList(loc *ingress.Location, rawLocationList string) bool {
	locationList := strings.Split(rawLocationList, ",")

	for _, locationListItem := range locationList {
		locationListItem = strings.Trim(locationListItem, " ")
		if locationListItem == "" {
			continue
		}
		if strings.HasPrefix(loc.Path, locationListItem) {
			return true
		}
	}

	return false
}

func buildAuthLocationConfig(location *ingress.Location, locationConfig locationCfg) ngx_crossplane.Directives {
	directives := make(ngx_crossplane.Directives, 0)
	if locationConfig.authPath != "" {
		if locationConfig.applyAuthUpstream && !locationConfig.applyGlobalAuth {
			directives = append(directives, buildDirective("set", "$auth_cookie", ""),
				buildDirective("add_header", "Set-Cookie", "$auth_cookie"))
			directives = append(directives, buildAuthResponseHeaders(locationConfig.proxySetHeader, locationConfig.externalAuth.ResponseHeaders, true)...)
			if len(locationConfig.externalAuth.ResponseHeaders) > 0 {
				directives = append(directives, buildDirective("set", "$auth_response_headers", strings.Join(locationConfig.externalAuth.ResponseHeaders, ",")))
			}
			directives = append(directives,
				buildDirective("set", "$auth_path", locationConfig.authPath),
				buildDirective("set", "$auth_keepalive_share_vars", strconv.FormatBool(locationConfig.externalAuth.KeepaliveShareVars)),
				buildDirective("access_by_lua_file", "/etc/nginx/lua/nginx/ngx_conf_external_auth.lua"),
			)
		} else {
			directives = append(directives,
				buildDirective("auth_request", locationConfig.authPath),
				buildDirective("auth_request_set", "$auth_cookie", "$upstream_http_set_cookie"),
			)
			cookieDirective := buildDirective("add_header", "Set-Cookie", "$auth_cookie")
			if locationConfig.externalAuth.AlwaysSetCookie {
				cookieDirective.Args = append(cookieDirective.Args, "always")
			}
			directives = append(directives, cookieDirective)
			directives = append(directives, buildAuthResponseHeaders(locationConfig.proxySetHeader, locationConfig.externalAuth.ResponseHeaders, false)...)
		}
	}

	if locationConfig.externalAuth.SigninURL != "" {
		directives = append(directives,
			buildDirective("set_escape_uri", "$escaped_request_uri", "$request_uri"),
			buildDirective("error_page", "401", "=", buildAuthSignURLLocation(location.Path, locationConfig.externalAuth.SigninURL)),
		)
	}
	if location.BasicDigestAuth.Secured {
		var authDirective, authFileDirective string
		if location.BasicDigestAuth.Type == "basic" {
			authDirective, authFileDirective = "auth_basic", "auth_basic_user_file"
		} else {
			authDirective, authFileDirective = "auth_digest", "auth_digest_user_file"
		}

		directives = append(directives,
			buildDirective(authDirective, location.BasicDigestAuth.Realm),
			buildDirective(authFileDirective, location.BasicDigestAuth.File),
			buildDirective(locationConfig.proxySetHeader, "Authorization", ""),
		)
	}

	return directives
}
