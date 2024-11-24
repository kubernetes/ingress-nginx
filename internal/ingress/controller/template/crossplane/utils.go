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
	"crypto/sha1" //nolint:gosec // We cannot move away from sha1
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ratelimit"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

const (
	slash                   = "/"
	nonIdempotent           = "non_idempotent"
	defBufferSize           = 65535
	writeIndentOnEmptyLines = true // backward-compatibility
	httpProtocol            = "HTTP"
	autoHTTPProtocol        = "AUTO_HTTP"
	httpsProtocol           = "HTTPS"
	grpcProtocol            = "GRPC"
	grpcsProtocol           = "GRPCS"
	fcgiProtocol            = "FCGI"
)

var (
	nginxSizeRegex                 = regexp.MustCompile(`^\d+[kKmM]?$`)
	nginxOffsetRegex               = regexp.MustCompile(`^\d+[kKmMgG]?$`)
	defaultGlobalAuthRedirectParam = "rd"
)

type (
	seconds int
	minutes int
)

func buildDirectiveWithComment(directive, comment string, args ...any) *ngx_crossplane.Directive {
	dir := buildDirective(directive, args...)
	dir.Comment = ptr.To(comment)
	return dir
}

func buildStartServer(name string) *ngx_crossplane.Directive {
	return buildDirective("##", "start", "server", name)
}

func buildEndServer(name string) *ngx_crossplane.Directive {
	return buildDirective("##", "end", "server", name)
}

func buildStartAuthUpstream(name, location string) *ngx_crossplane.Directive {
	return buildDirective("##", "start", "auth", "upstream", name, location)
}

func buildEndAuthUpstream(name, location string) *ngx_crossplane.Directive {
	return buildDirective("##", "end", "auth", "upstream", name, location)
}

func buildDirective(directive string, args ...any) *ngx_crossplane.Directive {
	argsVal := make([]string, 0)
	for k := range args {
		switch v := args[k].(type) {
		case string:
			argsVal = append(argsVal, v)
		case *string:
			if v != nil {
				argsVal = append(argsVal, *v)
			}
		case []string:
			argsVal = append(argsVal, v...)
		case int:
			argsVal = append(argsVal, strconv.Itoa(v))
		case bool:
			argsVal = append(argsVal, boolToStr(v))
		case seconds:
			argsVal = append(argsVal, strconv.Itoa(int(v))+"s")
		case minutes:
			argsVal = append(argsVal, strconv.Itoa(int(v))+"m")
		}
	}
	return &ngx_crossplane.Directive{
		Directive: directive,
		Args:      argsVal,
	}
}

func buildLuaSharedDictionaries(cfg *config.Configuration) []*ngx_crossplane.Directive {
	out := make([]*ngx_crossplane.Directive, 0, len(cfg.LuaSharedDicts))
	for name, size := range cfg.LuaSharedDicts {
		sizeStr := dictKbToStr(size)
		out = append(out, buildDirective("lua_shared_dict", name, sizeStr))
	}

	return out
}

// TODO: The utils below should be moved to a level where they can be consumed by any template writer

// buildResolvers returns the resolvers reading the /etc/resolv.conf file
func buildResolversInternal(res []net.IP, disableIpv6 bool) []string {
	r := make([]string, 0)
	for _, ns := range res {
		if ing_net.IsIPV6(ns) {
			if disableIpv6 {
				continue
			}
			r = append(r, fmt.Sprintf("[%s]", ns))
		} else {
			r = append(r, ns.String())
		}
	}
	r = append(r, "valid=30s")

	if disableIpv6 {
		r = append(r, "ipv6=off")
	}

	return r
}

// buildBlockDirective is used to build a block directive
func buildBlockDirective(blockName string, args []string, block ngx_crossplane.Directives) *ngx_crossplane.Directive {
	return &ngx_crossplane.Directive{
		Directive: blockName,
		Args:      args,
		Block:     block,
	}
}

// buildMapDirective is used to build a map directive
func buildMapDirective(name, variable string, block ngx_crossplane.Directives) *ngx_crossplane.Directive {
	return buildBlockDirective("map", []string{name, variable}, block)
}

func boolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func dictKbToStr(size int) string {
	if size%1024 == 0 {
		return fmt.Sprintf("%dM", size/1024)
	}
	return fmt.Sprintf("%dK", size)
}

func shouldLoadAuthDigestModule(servers []*ingress.Server) bool {
	for _, server := range servers {
		for _, location := range server.Locations {
			if !location.BasicDigestAuth.Secured {
				continue
			}

			if location.BasicDigestAuth.Type == "digest" {
				return true
			}
		}
	}
	return false
}

// shouldLoadOpentelemetryModule determines whether or not the Opentelemetry module needs to be loaded.
// It checks if `enable-opentelemetry` is set in the ConfigMap.
func shouldLoadOpentelemetryModule(servers []*ingress.Server) bool {
	for _, server := range servers {
		for _, location := range server.Locations {
			if location.Opentelemetry.Enabled {
				return true
			}
		}
	}
	return false
}

func buildServerName(hostname string) string {
	if !strings.HasPrefix(hostname, "*") {
		return hostname
	}

	hostname = strings.Replace(hostname, "*.", "", 1)
	parts := strings.Split(hostname, ".")

	return `~^(?<subdomain>[\w-]+)\.` + strings.Join(parts, "\\.") + `$`
}

func buildListener(tc *config.TemplateConfig, hostname string) ngx_crossplane.Directives {
	listenDirectives := make(ngx_crossplane.Directives, 0)

	co := commonListenOptions(tc, hostname)

	addrV4 := []string{""}
	if len(tc.Cfg.BindAddressIpv4) > 0 {
		addrV4 = tc.Cfg.BindAddressIpv4
	}
	listenDirectives = append(listenDirectives, httpListener(addrV4, co, tc, false)...)
	listenDirectives = append(listenDirectives, httpListener(addrV4, co, tc, true)...)

	if tc.IsIPV6Enabled {
		addrV6 := []string{"[::]"}
		if len(tc.Cfg.BindAddressIpv6) > 0 {
			addrV6 = tc.Cfg.BindAddressIpv6
		}
		listenDirectives = append(listenDirectives, httpListener(addrV6, co, tc, false)...)
		listenDirectives = append(listenDirectives, httpListener(addrV6, co, tc, true)...)
	}

	return listenDirectives
}

// commonListenOptions defines the common directives that should be added to NGINX listeners
func commonListenOptions(template *config.TemplateConfig, hostname string) []string {
	var out []string

	if template.Cfg.UseProxyProtocol {
		out = append(out, "proxy_protocol")
	}

	if hostname != "_" {
		return out
	}

	out = append(out, "default_server")

	if template.Cfg.ReusePort {
		out = append(out, "reuseport")
	}
	out = append(out, fmt.Sprintf("backlog=%d", template.BacklogSize))
	return out
}

func httpListener(addresses, co []string, tc *config.TemplateConfig, ssl bool) ngx_crossplane.Directives {
	listeners := make(ngx_crossplane.Directives, 0)
	port := tc.ListenPorts.HTTP
	isTLSProxy := tc.IsSSLPassthroughEnabled
	// If this is a SSL listener we should mutate the port properly
	if ssl {
		port = tc.ListenPorts.HTTPS
		if isTLSProxy {
			port = tc.ListenPorts.SSLProxy
		}
	}
	for _, address := range addresses {
		var listenAddress string
		if address == "" {
			listenAddress = fmt.Sprintf("%d", port)
		} else {
			listenAddress = fmt.Sprintf("%s:%d", address, port)
		}
		if ssl {
			if isTLSProxy {
				co = append(co, "proxy_protocol")
			}
			co = append(co, "ssl")
		}
		listenDirective := buildDirective("listen", listenAddress, co)
		listeners = append(listeners, listenDirective)
	}

	return listeners
}

func luaConfigurationRequestBodySize(cfg *config.Configuration) string {
	size := cfg.LuaSharedDicts["configuration_data"]
	if size < cfg.LuaSharedDicts["certificate_data"] {
		size = cfg.LuaSharedDicts["certificate_data"]
	}
	size += 1024

	return dictKbToStr(size)
}

func buildLocation(location *ingress.Location, enforceRegex bool) []string {
	path := location.Path
	if enforceRegex {
		return []string{"~*", fmt.Sprintf("^%s", path)}
	}

	if location.PathType != nil && *location.PathType == networkingv1.PathTypeExact {
		return []string{"=", path}
	}

	return []string{path}
}

func getProxySetHeader(location *ingress.Location) string {
	if location.BackendProtocol == grpcProtocol || location.BackendProtocol == grpcsProtocol {
		return "grpc_set_header"
	}

	return "proxy_set_header"
}

func buildAuthLocation(location *ingress.Location, globalExternalAuthURL string) string {
	if (location.ExternalAuth.URL == "") && (!shouldApplyGlobalAuth(location, globalExternalAuthURL)) {
		return ""
	}

	str := base64.URLEncoding.EncodeToString([]byte(location.Path))
	// removes "=" after encoding
	str = strings.ReplaceAll(str, "=", "")

	pathType := "default"
	if location.PathType != nil {
		pathType = string(*location.PathType)
	}

	return fmt.Sprintf("/_external-auth-%v-%v", str, pathType)
}

// shouldApplyGlobalAuth returns true only in case when ExternalAuth.URL is not set and
// GlobalExternalAuth is set and enabled
func shouldApplyGlobalAuth(location *ingress.Location, globalExternalAuthURL string) bool {
	return location.ExternalAuth.URL == "" &&
		globalExternalAuthURL != "" &&
		location.EnableGlobalAuth
}

// shouldApplyAuthUpstream returns true only in case when ExternalAuth.URL and
// ExternalAuth.KeepaliveConnections are all set
func shouldApplyAuthUpstream(location *ingress.Location, cfg *config.Configuration) bool {
	if location.ExternalAuth.URL == "" || location.ExternalAuth.KeepaliveConnections == 0 {
		return false
	}

	// Unfortunately, `auth_request` module ignores keepalive in upstream block: https://trac.nginx.org/nginx/ticket/1579
	// The workaround is to use `ngx.location.capture` Lua subrequests but it is not supported with HTTP/2
	if cfg.UseHTTP2 {
		return false
	}
	return true
}

func isValidByteSize(s string, isOffset bool) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	if isOffset {
		return nginxOffsetRegex.MatchString(s)
	}

	return nginxSizeRegex.MatchString(s)
}

func buildAuthUpstreamName(input *ingress.Location, host string) string {
	authPath := buildAuthLocation(input, "")
	if authPath == "" || host == "" {
		return ""
	}

	return fmt.Sprintf("%s-%s", host, authPath[2:])
}

// changeHostPort will change the host:port part of the url to value
func changeHostPort(newURL, value string) string {
	if newURL == "" {
		return ""
	}

	authURL, err := parser.StringToURL(newURL)
	if err != nil {
		klog.Errorf("expected a valid URL but %s was returned", newURL)
		return ""
	}

	authURL.Host = value

	return authURL.String()
}

func buildAuthSignURLLocation(location, authSignURL string) string {
	hasher := sha1.New() //nolint:gosec // We cannot move away from sha1
	hasher.Write([]byte(location))
	hasher.Write([]byte(authSignURL))
	return "@" + hex.EncodeToString(hasher.Sum(nil))
}

func buildAuthSignURL(authSignURL, authRedirectParam string) string {
	u, err := url.Parse(authSignURL)
	if err != nil {
		klog.Errorf("error parsing authSignURL: %v", err)
		return ""
	}
	q := u.Query()
	if authRedirectParam == "" {
		authRedirectParam = defaultGlobalAuthRedirectParam
	}
	if len(q) == 0 {
		return fmt.Sprintf("%v?%v=$pass_access_scheme://$http_host$escaped_request_uri", authSignURL, authRedirectParam)
	}

	if q.Get(authRedirectParam) != "" {
		return authSignURL
	}

	return fmt.Sprintf("%v&%v=$pass_access_scheme://$http_host$escaped_request_uri", authSignURL, authRedirectParam)
}

func buildCorsOriginRegex(corsOrigins []string) ngx_crossplane.Directives {
	if len(corsOrigins) == 1 && corsOrigins[0] == "*" {
		return ngx_crossplane.Directives{
			buildDirective("set", "$http_origin", "*"),
			buildDirective("set", "$cors", "true"),
		}
	}

	originsArray := []string{"("}
	for i, origin := range corsOrigins {
		originTrimmed := strings.TrimSpace(origin)
		if originTrimmed != "" {
			originsArray = append(originsArray, buildOriginRegex(originTrimmed))
		}
		if i != len(corsOrigins)-1 {
			originsArray = append(originsArray, "|")
		}
	}
	originsArray = append(originsArray, ")$")

	// originsArray should be converted to a single string, as it is a single directive for if.
	origins := strings.Join(originsArray, "")
	return ngx_crossplane.Directives{
		buildBlockDirective("if", []string{"$http_origin", "~*", origins}, ngx_crossplane.Directives{
			buildDirective("set", "$cors", "true"),
		}),
	}
}

func buildOriginRegex(origin string) string {
	origin = regexp.QuoteMeta(origin)
	origin = strings.Replace(origin, "\\*", `[A-Za-z0-9\-]+`, 1)
	return fmt.Sprintf("(%s)", origin)
}

func buildNextUpstream(nextUpstream string, retryNonIdempotent bool) []string {
	parts := strings.Split(nextUpstream, " ")

	nextUpstreamCodes := make([]string, 0, len(parts))
	for _, v := range parts {
		if v != "" && v != nonIdempotent {
			nextUpstreamCodes = append(nextUpstreamCodes, v)
		}

		if v == nonIdempotent {
			retryNonIdempotent = true
		}
	}

	if retryNonIdempotent {
		nextUpstreamCodes = append(nextUpstreamCodes, nonIdempotent)
	}

	return nextUpstreamCodes
}

func buildProxyPass(backends []*ingress.Backend, location *ingress.Location) ngx_crossplane.Directives {
	path := location.Path
	proto := "http://"
	proxyPass := "proxy_pass"

	switch strings.ToUpper(location.BackendProtocol) {
	case autoHTTPProtocol:
		proto = "$scheme://"
	case httpsProtocol:
		proto = "https://"
	case grpcProtocol:
		proto = "grpc://"
		proxyPass = "grpc_pass"
	case grpcsProtocol:
		proto = "grpcs://"
		proxyPass = "grpc_pass"
	case fcgiProtocol:
		proto = ""
		proxyPass = "fastcgi_pass"
	}

	upstreamName := "upstream_balancer"

	for _, backend := range backends {
		if backend.Name == location.Backend {
			if backend.SSLPassthrough {
				proto = "https://"

				if location.BackendProtocol == grpcsProtocol {
					proto = "grpcs://"
				}
			}

			break
		}
	}

	if location.Backend == "upstream-default-backend" {
		proto = "http://"
		proxyPass = "proxy_pass"
	}

	// defProxyPass returns the default proxy_pass, just the name of the upstream
	defProxyPass := buildDirective(proxyPass, fmt.Sprintf("%s%s", proto, upstreamName))

	// if the path in the ingress rule is equals to the target: no special rewrite
	if path == location.Rewrite.Target {
		return ngx_crossplane.Directives{defProxyPass}
	}

	if location.Rewrite.Target != "" {
		proxySetHeader := "proxy_set_header"
		dir := make(ngx_crossplane.Directives, 0)
		if location.BackendProtocol == grpcProtocol || location.BackendProtocol == grpcsProtocol {
			proxySetHeader = "grpc_set_header"
		}

		if location.XForwardedPrefix != "" {
			dir = append(dir,
				buildDirective(proxySetHeader, "X-Forwarded-Prefix", location.XForwardedPrefix),
			)
		}

		dir = append(dir,
			buildDirective("rewrite", fmt.Sprintf("(?i)%s", path), location.Rewrite.Target, "break"),
			buildDirective(proxyPass, fmt.Sprintf("%s%s", proto, upstreamName)),
		)
		return dir
	}

	// default proxy_pass
	return ngx_crossplane.Directives{defProxyPass}
}

func buildGeoIPDirectives(reloadTime int, files []string) ngx_crossplane.Directives {
	directives := make(ngx_crossplane.Directives, 0)
	buildGeoIPBlock := func(file string, directives ngx_crossplane.Directives) *ngx_crossplane.Directive {
		if reloadTime > 0 && file != "GeoIP2-Connection-Type.mmdb" {
			directives = append(directives, buildDirective("auto_reload", minutes(reloadTime)))
		}
		fileName := fmt.Sprintf("/etc/ingress-controller/geoip/%s", file)
		return buildBlockDirective("geoip2", []string{fileName}, directives)
	}

	for _, file := range files {
		if file == "GeoLite2-Country.mmdb" || file == "GeoIP2-Country.mmdb" {
			directives = append(directives, buildGeoIPBlock(file, ngx_crossplane.Directives{
				buildDirective("$geoip2_country_code", "source=$remote_addr", "country", "iso_code"),
				buildDirective("$geoip2_country_name", "source=$remote_addr", "country", "names", "en"),
				buildDirective("$geoip2_country_geoname_id", "source=$remote_addr", "country", "geoname_id"),
				buildDirective("$geoip2_continent_code", "source=$remote_addr", "continent", "code"),
				buildDirective("$geoip2_continent_name", "source=$remote_addr", "continent", "names", "en"),
				buildDirective("$geoip2_continent_geoname_id", "source=$remote_addr", "continent", "geoname_id"),
			}))
		}
		if file == "GeoLite2-City.mmdb" || file == "GeoIP2-City.mmdb" {
			directives = append(directives, buildGeoIPBlock(file, ngx_crossplane.Directives{
				buildDirective("$geoip2_city_country_code", "source=$remote_addr", "country", "iso_code"),
				buildDirective("$geoip2_city_country_name", "source=$remote_addr", "country", "names", "en"),
				buildDirective("$geoip2_city_country_geoname_id", "source=$remote_addr", "country", "geoname_id"),
				buildDirective("$geoip2_city_continent_code", "source=$remote_addr", "continent", "code"),
				buildDirective("$geoip2_city_continent_name", "source=$remote_addr", "continent", "names", "en"),
				buildDirective("$geoip2_city", "source=$remote_addr", "city", "names", "en"),
				buildDirective("$geoip2_city_geoname_id", "source=$remote_addr", "city", "geoname_id"),
				buildDirective("$geoip2_postal_code", "source=$remote_addr", "postal", "code"),
				buildDirective("$geoip2_dma_code", "source=$remote_addr", "location", "metro_code"),
				buildDirective("$geoip2_latitude", "source=$remote_addr", "location", "latitude"),
				buildDirective("$geoip2_longitude", "source=$remote_addr", "location", "longitude"),
				buildDirective("$geoip2_time_zone", "source=$remote_addr", "location", "time_zone"),
				buildDirective("$geoip2_region_code", "source=$remote_addr", "subdivisions", "0", "iso_code"),
				buildDirective("$geoip2_region_name", "source=$remote_addr", "subdivisions", "0", "names", "en"),
				buildDirective("$geoip2_region_geoname_id", "source=$remote_addr", "subdivisions", "0", "geoname_id"),
				buildDirective("$geoip2_subregion_code", "source=$remote_addr", "subdivisions", "1", "iso_code"),
				buildDirective("$geoip2_subregion_name", "source=$remote_addr", "subdivisions", "1", "names", "en"),
				buildDirective("$geoip2_subregion_geoname_id", "source=$remote_addr", "subdivisions", "1", "geoname_id"),
			}))
		}
		if file == "GeoLite2-ASN.mmdb" || file == "GeoIP2-ASN.mmdb" {
			directives = append(directives, buildGeoIPBlock(file, ngx_crossplane.Directives{
				buildDirective("$geoip2_asn", "source=$remote_addr", "autonomous_system_number"),
				buildDirective("$geoip2_org", "source=$remote_addr", "autonomous_system_organization"),
			}))
		}
		if file == "GeoIP2-ISP.mmdb" {
			directives = append(directives, buildGeoIPBlock(file, ngx_crossplane.Directives{
				buildDirective("$geoip2_isp", "source=$remote_addr", "isp"),
				buildDirective("$geoip2_isp_org", "source=$remote_addr", "organization"),
				buildDirective("$geoip2_asn", "source=$remote_addr", "autonomous_system_number"),
			}))
		}
		if file == "GeoIP2-Anonymous-IP.mmdb" {
			directives = append(directives, buildGeoIPBlock(file, ngx_crossplane.Directives{
				buildDirective("$geoip2_is_anon", "source=$remote_addr", "is_anonymous"),
				buildDirective("$geoip2_is_anonymous", "source=$remote_addr", "default=0", "is_anonymous"),
				buildDirective("$geoip2_is_anonymous_vpn", "source=$remote_addr", "default=0", "is_anonymous_vpn"),
				buildDirective("$geoip2_is_hosting_provider", "source=$remote_addr", "default=0", "is_hosting_provider"),
				buildDirective("$geoip2_is_public_proxy", "source=$remote_addr", "default=0", "is_public_proxy"),
				buildDirective("$geoip2_is_tor_exit_node", "source=$remote_addr", "default=0", "is_tor_exit_node"),
			}))
		}
		if file == "GeoIP2-Connection-Type.mmdb" {
			directives = append(directives, buildGeoIPBlock(file, ngx_crossplane.Directives{
				buildDirective("$geoip2_connection_type", "connection_type"),
			}))
		}
	}
	return directives
}

func filterRateLimits(servers []*ingress.Server) []ratelimit.Config {
	ratelimits := []ratelimit.Config{}
	found := sets.Set[string]{}

	for _, server := range servers {
		for _, loc := range server.Locations {
			if loc.RateLimit.ID != "" && !found.Has(loc.RateLimit.ID) {
				found.Insert(loc.RateLimit.ID)
				ratelimits = append(ratelimits, loc.RateLimit)
			}
		}
	}
	return ratelimits
}

// buildRateLimitZones produces an array of limit_conn_zone in order to allow
// rate limiting of request. Each Ingress rule could have up to three zones, one
// for connection limit by IP address, one for limiting requests per minute, and
// one for limiting requests per second.
func buildRateLimitZones(servers []*ingress.Server) ngx_crossplane.Directives {
	zones := make(map[string]bool)
	directives := make(ngx_crossplane.Directives, 0)
	for _, server := range servers {
		for _, loc := range server.Locations {
			zoneID := fmt.Sprintf("$limit_%s", loc.RateLimit.ID)
			if loc.RateLimit.Connections.Limit > 0 {
				zoneArg := fmt.Sprintf("zone=%s:%dm", loc.RateLimit.Connections.Name, loc.RateLimit.Connections.SharedSize)
				zone := fmt.Sprintf("limit_conn_zone %s %s", zoneID, zoneArg)
				if _, ok := zones[zone]; !ok {
					zones[zone] = true
					directives = append(directives, buildDirective("limit_conn_zone", zoneID, zoneArg))
				}
			}

			if loc.RateLimit.RPM.Limit > 0 {
				zoneArg := fmt.Sprintf("zone=%s:%dm", loc.RateLimit.RPM.Name, loc.RateLimit.RPM.SharedSize)
				zoneRate := fmt.Sprintf("rate=%dr/m", loc.RateLimit.RPM.Limit)
				zone := fmt.Sprintf("limit_req_zone %s %s %s", zoneID, zoneArg, zoneRate)
				if _, ok := zones[zone]; !ok {
					zones[zone] = true
					directives = append(directives, buildDirective("limit_req_zone", zoneID, zoneArg, zoneRate))
				}
			}

			if loc.RateLimit.RPS.Limit > 0 {
				zoneArg := fmt.Sprintf("zone=%s:%dm", loc.RateLimit.RPS.Name, loc.RateLimit.RPS.SharedSize)
				zoneRate := fmt.Sprintf("rate=%dr/s", loc.RateLimit.RPS.Limit)
				zone := fmt.Sprintf("limit_req_zone %s %s %s", zoneID, zoneArg, zoneRate)
				if _, ok := zones[zone]; !ok {
					zones[zone] = true
					directives = append(directives, buildDirective("limit_req_zone", zoneID, zoneArg, zoneRate))
				}
			}
		}
	}
	return directives
}

// buildAuthResponseHeaders sets HTTP response headers when `auth-url` is used.
// Based on `auth-keepalive` value we use auth_request_set Nginx directives, or
// we use Lua and Nginx variables instead.
//
// NOTE: Unfortunately auth_request module ignores the keepalive directive (see:
// https://trac.nginx.org/nginx/ticket/1579), that is why we mimic the same
// functionality with access_by_lua_block.
// TODO: This function is duplicated with the non-crossplane and we should consolidate
func buildAuthResponseHeaders(proxySetHeader string, headers []string, lua bool) ngx_crossplane.Directives {
	res := make(ngx_crossplane.Directives, 0)

	if len(headers) == 0 {
		return res
	}

	for i, h := range headers {
		authHeader := fmt.Sprintf("$authHeader%d", i)
		if lua {
			res = append(res, buildDirective("set", authHeader, ""))
		} else {
			hvar := strings.ToLower(h)
			hvar = strings.NewReplacer("-", "_").Replace(hvar)
			res = append(res, buildDirective("auth_request_set",
				authHeader,
				fmt.Sprintf("$upstream_http_%s", hvar)))
		}
		res = append(res, buildDirective(proxySetHeader, h, authHeader))
	}
	return res
}

// extractHostPort will extract the host:port part from the URL specified by url
func extractHostPort(newURL string) string {
	if newURL == "" {
		return ""
	}

	authURL, err := parser.StringToURL(newURL)
	if err != nil {
		klog.Errorf("expected a valid URL but %s was returned", newURL)
		return ""
	}

	return authURL.Host
}
