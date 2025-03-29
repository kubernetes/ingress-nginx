/*
Copyright 2015 The Kubernetes Authors.

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

package template

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1" // #nosec
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	text_template "text/template"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

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

const (
	stateCode = iota
	stateComment
)

// Writer is the interface to render a template
type Writer interface {
	// Write renders the template.
	// NOTE: Implementors must ensure that the content of the returned slice is not modified by the implementation
	// after the return of this function.
	Write(conf *config.TemplateConfig) ([]byte, error)
	// Validate is a function that can be called, containing the file name to be tested
	// This function should be used just by specific cases like crossplane, otherwise it can return
	// null error
	Validate(filename string) error
}

// Template ingress template
type Template struct {
	tmpl *text_template.Template

	bp *BufferPool
}

// NewTemplate returns a new Template instance or an
// error if the specified template file contains errors
func NewTemplate(file string) (*Template, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unexpected error reading template %s: %w", file, err)
	}

	tmpl, err := text_template.New("nginx.tmpl").Funcs(funcMap).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl: tmpl,
		bp:   NewBufferPool(defBufferSize),
	}, nil
}

// 1. Removes carriage return symbol (\r)
// 2. Collapses multiple empty lines to single one
// 3. Re-indent
// (ATW: always returns nil)
func cleanConf(in, out *bytes.Buffer) error {
	depth := 0
	lineStarted := false
	emptyLineWritten := false
	state := stateCode
	for {
		c, err := in.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err // unreachable
		}

		needOutput := false
		nextDepth := depth
		nextLineStarted := lineStarted

		switch state {
		case stateCode:
			switch c {
			case '{':
				needOutput = true
				nextDepth = depth + 1
				nextLineStarted = true
			case '}':
				needOutput = true
				depth--
				nextDepth = depth
				nextLineStarted = true
			case ' ', '\t':
				needOutput = lineStarted
			case '\r':
				rest := in.Bytes()
				if len(rest) > 0 {
					if rest[0] != '\n' {
						c = ' '
						needOutput = lineStarted
					}
				}
			case '\n':
				needOutput = !(!lineStarted && emptyLineWritten)
				nextLineStarted = false
			case '#':
				needOutput = true
				nextLineStarted = true
				state = stateComment
			default:
				needOutput = true
				nextLineStarted = true
			}
		case stateComment:
			switch c {
			case '\r':
				rest := in.Bytes()
				if len(rest) > 0 {
					if rest[0] != '\n' {
						c = ' '
						needOutput = lineStarted
					}
				}
			case '\n':
				needOutput = true
				nextLineStarted = false
				state = stateCode
			default:
				needOutput = true
			}
		}

		if needOutput {
			if !lineStarted && (writeIndentOnEmptyLines || c != '\n') {
				for i := 0; i < depth; i++ {
					err = out.WriteByte('\t') // always nil
					if err != nil {
						return err
					}
				}
			}
			emptyLineWritten = !lineStarted
			err = out.WriteByte(c) // always nil
			if err != nil {
				return err
			}
		}

		depth = nextDepth
		lineStarted = nextLineStarted
	}
}

/* LuaConfig defines the structure that will be written as a config for lua scripts
The json format should follow what's expected by lua:
		use_forwarded_headers = %t,
		use_proxy_protocol = %t,
		is_ssl_passthrough_enabled = %t,
		http_redirect_code = %v,
		listen_ports = { ssl_proxy = "%v", https = "%v" },

		hsts = %t,
		hsts_max_age = %v,
		hsts_include_subdomains = %t,
		hsts_preload = %t,
*/

type LuaConfig struct {
	EnableMetrics           bool           `json:"enable_metrics"`
	ListenPorts             LuaListenPorts `json:"listen_ports"`
	UseForwardedHeaders     bool           `json:"use_forwarded_headers"`
	UseProxyProtocol        bool           `json:"use_proxy_protocol"`
	IsSSLPassthroughEnabled bool           `json:"is_ssl_passthrough_enabled"`
	HTTPRedirectCode        int            `json:"http_redirect_code"`
	EnableOCSP              bool           `json:"enable_ocsp"`
	MonitorBatchMaxSize     int            `json:"monitor_batch_max_size"`
	HSTS                    bool           `json:"hsts"`
	HSTSMaxAge              string         `json:"hsts_max_age"`
	HSTSIncludeSubdomains   bool           `json:"hsts_include_subdomains"`
	HSTSPreload             bool           `json:"hsts_preload"`
}

type LuaListenPorts struct {
	HTTPSPort    string `json:"https"`
	StatusPort   string `json:"status_port"`
	SSLProxyPort string `json:"ssl_proxy"`
}

// Validate is no-op at go-template
func (t *Template) Validate(_ string) error {
	return nil
}

// Write populates a buffer using a template with NGINX configuration
// and the servers and upstreams created by Ingress rules
func (t *Template) Write(conf *config.TemplateConfig) ([]byte, error) {
	tmplBuf := t.bp.Get()
	defer t.bp.Put(tmplBuf)

	outCmdBuf := t.bp.Get()
	defer t.bp.Put(outCmdBuf)

	if klog.V(3).Enabled() {
		b, err := json.Marshal(*conf)
		if err != nil {
			klog.Errorf("unexpected error: %v", err)
		}
		klog.InfoS("NGINX", "configuration", string(b))
	}

	err := t.tmpl.Execute(tmplBuf, *conf)
	if err != nil {
		return nil, err
	}

	// squeezes multiple adjacent empty lines to be single
	// spaced this is to avoid the use of regular expressions
	err = cleanConf(tmplBuf, outCmdBuf)
	if err != nil {
		return nil, err
	}

	// make a copy to ensure that we are no longer modifying the content of the buffer
	out := outCmdBuf.Bytes()
	res := make([]byte, len(out))
	copy(res, out)

	return res, nil
}

var funcMap = text_template.FuncMap{
	"empty": func(input interface{}) bool {
		check, ok := input.(string)
		if ok {
			return check == ""
		}
		return true
	},
	"escapeLiteralDollar":             escapeLiteralDollar,
	"buildLuaSharedDictionaries":      buildLuaSharedDictionaries,
	"luaConfigurationRequestBodySize": luaConfigurationRequestBodySize,
	"buildLocation":                   buildLocation,
	"buildAuthLocation":               buildAuthLocation,
	"shouldApplyGlobalAuth":           shouldApplyGlobalAuth,
	"buildAuthResponseHeaders":        buildAuthResponseHeaders,
	"buildAuthUpstreamLuaHeaders":     buildAuthUpstreamLuaHeaders,
	"buildAuthProxySetHeaders":        buildAuthProxySetHeaders,
	"buildAuthUpstreamName":           buildAuthUpstreamName,
	"shouldApplyAuthUpstream":         shouldApplyAuthUpstream,
	"extractHostPort":                 extractHostPort,
	"changeHostPort":                  changeHostPort,
	"buildProxyPass":                  buildProxyPass,
	"filterRateLimits":                filterRateLimits,
	"buildRateLimitZones":             buildRateLimitZones,
	"buildRateLimit":                  buildRateLimit,
	"locationConfigForLua":            locationConfigForLua,
	"buildResolvers":                  buildResolvers,
	"buildUpstreamName":               buildUpstreamName,
	"isLocationInLocationList":        isLocationInLocationList,
	"isLocationAllowed":               isLocationAllowed,
	"buildDenyVariable":               buildDenyVariable,
	"getenv":                          os.Getenv,
	"contains":                        strings.Contains,
	"split":                           strings.Split,
	"hasPrefix":                       strings.HasPrefix,
	"hasSuffix":                       strings.HasSuffix,
	"trimSpace":                       strings.TrimSpace,
	"toUpper":                         strings.ToUpper,
	"toLower":                         strings.ToLower,
	"formatIP":                        formatIP,
	"quote":                           quote,
	"buildNextUpstream":               buildNextUpstream,
	"getIngressInformation":           getIngressInformation,
	"serverConfig": func(all config.TemplateConfig, server *ingress.Server) interface{} {
		return struct{ First, Second interface{} }{all, server}
	},
	"isValidByteSize":                    isValidByteSize,
	"buildForwardedFor":                  buildForwardedFor,
	"buildAuthSignURL":                   buildAuthSignURL,
	"buildAuthSignURLLocation":           buildAuthSignURLLocation,
	"buildOpentelemetry":                 buildOpentelemetry,
	"proxySetHeader":                     proxySetHeader,
	"enforceRegexModifier":               enforceRegexModifier,
	"buildCustomErrorDeps":               buildCustomErrorDeps,
	"buildCustomErrorLocationsPerServer": buildCustomErrorLocationsPerServer,
	"shouldLoadModSecurityModule":        shouldLoadModSecurityModule,
	"buildHTTPListener":                  buildHTTPListener,
	"buildHTTPSListener":                 buildHTTPSListener,
	"buildOpentelemetryForLocation":      buildOpentelemetryForLocation,
	"shouldLoadOpentelemetryModule":      shouldLoadOpentelemetryModule,
	"buildModSecurityForLocation":        buildModSecurityForLocation,
	"buildMirrorLocations":               buildMirrorLocations,
	"shouldLoadAuthDigestModule":         shouldLoadAuthDigestModule,
	"buildServerName":                    buildServerName,
	"buildCorsOriginRegex":               buildCorsOriginRegex,
}

// escapeLiteralDollar will replace the $ character with ${literal_dollar}
// which is made to work via the following configuration in the http section of
// the template:
//
//	geo $literal_dollar {
//	    default "$";
//	}
func escapeLiteralDollar(input interface{}) string {
	inputStr, ok := input.(string)
	if !ok {
		return ""
	}
	return strings.ReplaceAll(inputStr, `$`, `${literal_dollar}`)
}

// formatIP will wrap IPv6 addresses in [] and return IPv4 addresses
// without modification. If the input cannot be parsed as an IP address
// it is returned without modification.
func formatIP(input string) string {
	ip := net.ParseIP(input)
	if ip == nil {
		return input
	}
	if v4 := ip.To4(); v4 != nil {
		return input
	}
	return fmt.Sprintf("[%s]", input)
}

func quote(input interface{}) string {
	var inputStr string
	switch input := input.(type) {
	case string:
		inputStr = input
	case fmt.Stringer:
		inputStr = input.String()
	case *string:
		inputStr = *input
	default:
		inputStr = fmt.Sprintf("%v", input)
	}
	return fmt.Sprintf("%q", inputStr)
}

func buildLuaSharedDictionaries(c, s interface{}) string {
	cfg, ok := c.(config.Configuration)
	if !ok {
		klog.Errorf("expected a 'config.Configuration' type but %T was returned", c)
		return ""
	}

	_, ok = s.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Server' type but %T was returned", s)
		return ""
	}

	out := make([]string, 0, len(cfg.LuaSharedDicts))
	for name, size := range cfg.LuaSharedDicts {
		sizeStr := dictKbToStr(size)
		out = append(out, fmt.Sprintf("lua_shared_dict %s %s", name, sizeStr))
	}

	sort.Strings(out)

	return strings.Join(out, ";\n") + ";\n"
}

func luaConfigurationRequestBodySize(c interface{}) string {
	cfg, ok := c.(config.Configuration)
	if !ok {
		klog.Errorf("expected a 'config.Configuration' type but %T was returned", c)
		return "100M" // just a default number
	}

	size := cfg.LuaSharedDicts["configuration_data"]
	if size < cfg.LuaSharedDicts["certificate_data"] {
		size = cfg.LuaSharedDicts["certificate_data"]
	}
	size += 1024

	return dictKbToStr(size)
}

// locationConfigForLua formats some location specific configuration into Lua table represented as string
func locationConfigForLua(l, a interface{}) string {
	location, ok := l.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was given", l)
		return "{}"
	}

	all, ok := a.(config.TemplateConfig)
	if !ok {
		klog.Errorf("expected a 'config.TemplateConfig' type but %T was given", a)
		return "{}"
	}

	/* Lua expects the following vars
		force_ssl_redirect = string_to_bool(ngx.var.force_ssl_redirect),
	    ssl_redirect = string_to_bool(ngx.var.ssl_redirect),
	    force_no_ssl_redirect = string_to_bool(ngx.var.force_no_ssl_redirect),
	    preserve_trailing_slash = string_to_bool(ngx.var.preserve_trailing_slash),
	    use_port_in_redirects = string_to_bool(ngx.var.use_port_in_redirects),
	*/

	return fmt.Sprintf(`
	    set $force_ssl_redirect "%t";
	    set $ssl_redirect "%t";
	    set $force_no_ssl_redirect "%t";
	    set $preserve_trailing_slash "%t";
	    set $use_port_in_redirects "%t";
	`,
		location.Rewrite.ForceSSLRedirect,
		location.Rewrite.SSLRedirect,
		isLocationInLocationList(l, all.Cfg.NoTLSRedirectLocations),
		location.Rewrite.PreserveTrailingSlash,
		location.UsePortInRedirects,
	)
}

// buildResolvers returns the resolvers reading the /etc/resolv.conf file
func buildResolvers(res, disableIpv6 interface{}) string {
	// NGINX need IPV6 addresses to be surrounded by brackets
	nss, ok := res.([]net.IP)
	if !ok {
		klog.Errorf("expected a '[]net.IP' type but %T was returned", res)
		return ""
	}
	no6, ok := disableIpv6.(bool)
	if !ok {
		klog.Errorf("expected a 'bool' type but %T was returned", disableIpv6)
		return ""
	}

	if len(nss) == 0 {
		return ""
	}

	r := []string{"resolver"}
	for _, ns := range nss {
		if ing_net.IsIPV6(ns) {
			if no6 {
				continue
			}
			r = append(r, fmt.Sprintf("[%v]", ns))
		} else {
			r = append(r, ns.String())
		}
	}
	r = append(r, "valid=30s")

	if no6 {
		r = append(r, "ipv6=off")
	}

	return strings.Join(r, " ") + ";"
}

func needsRewrite(location *ingress.Location) bool {
	if location.Rewrite.Target != "" && location.Rewrite.Target != location.Path {
		return true
	}
	return false
}

// enforceRegexModifier checks if the "rewrite-target" or "use-regex" annotation
// is used on any location path within a server
func enforceRegexModifier(input interface{}) bool {
	locations, ok := input.([]*ingress.Location)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Location' type but %T was returned", input)
		return false
	}

	for _, location := range locations {
		if needsRewrite(location) || location.Rewrite.UseRegex {
			return true
		}
	}
	return false
}

// buildLocation produces the location string, if the ingress has redirects
// (specified through the nginx.ingress.kubernetes.io/rewrite-target annotation)
func buildLocation(input interface{}, enforceRegex bool) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return slash
	}

	path := location.Path
	if enforceRegex {
		return fmt.Sprintf(`~* "^%s"`, path)
	}

	if location.PathType != nil && *location.PathType == networkingv1.PathTypeExact {
		return fmt.Sprintf(`= %s`, path)
	}

	return path
}

func buildAuthLocation(input interface{}, globalExternalAuthURL string) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return ""
	}

	if (location.ExternalAuth.URL == "") && (!shouldApplyGlobalAuth(input, globalExternalAuthURL)) {
		return ""
	}

	str := base64.URLEncoding.EncodeToString([]byte(location.Path))
	// removes "=" after encoding
	str = strings.ReplaceAll(str, "=", "")

	pathType := "default"
	if location.PathType != nil {
		pathType = fmt.Sprintf("%v", *location.PathType)
	}

	return fmt.Sprintf("/_external-auth-%v-%v", str, pathType)
}

// shouldApplyGlobalAuth returns true only in case when ExternalAuth.URL is not set and
// GlobalExternalAuth is set and enabled
func shouldApplyGlobalAuth(input interface{}, globalExternalAuthURL string) bool {
	location, ok := input.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
	}

	if (location.ExternalAuth.URL == "") && (globalExternalAuthURL != "") && (location.EnableGlobalAuth) {
		return true
	}

	return false
}

// buildAuthResponseHeaders sets HTTP response headers when `auth-url` is used.
// Based on `auth-keepalive` value we use auth_request_set Nginx directives, or
// we use Lua and Nginx variables instead.
//
// NOTE: Unfortunately auth_request module ignores the keepalive directive (see:
// https://trac.nginx.org/nginx/ticket/1579), that is why we mimic the same
// functionality with access_by_lua_block.
func buildAuthResponseHeaders(proxySetHeader string, headers []string, lua bool) []string {
	res := []string{}

	if len(headers) == 0 {
		return res
	}

	for i, h := range headers {
		if lua {
			res = append(res, fmt.Sprintf("set $authHeader%d '';", i))
		} else {
			hvar := strings.ToLower(h)
			hvar = strings.NewReplacer("-", "_").Replace(hvar)
			res = append(res, fmt.Sprintf("auth_request_set $authHeader%v $upstream_http_%v;", i, hvar))
		}
		res = append(res, fmt.Sprintf("%s '%v' $authHeader%v;", proxySetHeader, h, i))
	}
	return res
}

func buildAuthUpstreamLuaHeaders(headers []string) string {
	if len(headers) == 0 {
		return ""
	}

	return strings.Join(headers, ",")
}

func buildAuthProxySetHeaders(headers map[string]string) []string {
	res := []string{}

	if len(headers) == 0 {
		return res
	}

	for name, value := range headers {
		res = append(res, fmt.Sprintf("proxy_set_header '%v' '%v';", name, value))
	}
	sort.Strings(res)
	return res
}

func buildAuthUpstreamName(input interface{}, host string) string {
	authPath := buildAuthLocation(input, "")
	if authPath == "" || host == "" {
		return ""
	}

	return fmt.Sprintf("%s-%s", host, authPath[2:])
}

// shouldApplyAuthUpstream returns true only in case when ExternalAuth.URL and
// ExternalAuth.KeepaliveConnections are all set
func shouldApplyAuthUpstream(l, c interface{}) bool {
	location, ok := l.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", l)
		return false
	}

	cfg, ok := c.(config.Configuration)
	if !ok {
		klog.Errorf("expected a 'config.Configuration' type but %T was returned", c)
		return false
	}

	if location.ExternalAuth.URL == "" || location.ExternalAuth.KeepaliveConnections == 0 {
		return false
	}

	// Unfortunately, `auth_request` module ignores keepalive in upstream block: https://trac.nginx.org/nginx/ticket/1579
	// The workaround is to use `ngx.location.capture` Lua subrequests but it is not supported with HTTP/2
	if cfg.UseHTTP2 {
		klog.Warning("Upstream keepalive is not supported with HTTP/2")
		return false
	}

	return true
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

// buildProxyPass produces the proxy pass string, if the ingress has redirects
// (specified through the nginx.ingress.kubernetes.io/rewrite-target annotation)
// If the annotation nginx.ingress.kubernetes.io/add-base-url:"true" is specified it will
// add a base tag in the head of the response from the service
func buildProxyPass(_ string, b, loc interface{}) string {
	backends, ok := b.([]*ingress.Backend)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Backend' type but %T was returned", b)
		return ""
	}

	location, ok := loc.(*ingress.Location)
	if !ok {
		klog.Errorf("expected a '*ingress.Location' type but %T was returned", loc)
		return ""
	}

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

	// TODO: add support for custom protocols
	if location.Backend == "upstream-default-backend" {
		proto = "http://"
		proxyPass = "proxy_pass"
	}

	// defProxyPass returns the default proxy_pass, just the name of the upstream
	defProxyPass := fmt.Sprintf("%v %s%s;", proxyPass, proto, upstreamName)

	// if the path in the ingress rule is equals to the target: no special rewrite
	if path == location.Rewrite.Target {
		return defProxyPass
	}

	if location.Rewrite.Target != "" {
		var xForwardedPrefix string

		if location.XForwardedPrefix != "" {
			xForwardedPrefix = fmt.Sprintf("%s X-Forwarded-Prefix %q;\n", proxySetHeader(location), location.XForwardedPrefix)
		}

		return fmt.Sprintf(`
rewrite "(?i)%s" %s break;
%v%v %s%s;`, path, location.Rewrite.Target, xForwardedPrefix, proxyPass, proto, upstreamName)
	}

	// default proxy_pass
	return defProxyPass
}

func filterRateLimits(input interface{}) []ratelimit.Config {
	ratelimits := []ratelimit.Config{}
	found := sets.Set[string]{}

	servers, ok := input.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected a '[]ratelimit.RateLimit' type but %T was returned", input)
		return ratelimits
	}
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
func buildRateLimitZones(input interface{}) []string {
	zones := sets.Set[string]{}

	servers, ok := input.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected a '[]*ingress.Server' type but %T was returned", input)
		return zones.UnsortedList()
	}

	for _, server := range servers {
		for _, loc := range server.Locations {
			if loc.RateLimit.Connections.Limit > 0 {
				zone := fmt.Sprintf("limit_conn_zone $limit_%s zone=%v:%vm;",
					loc.RateLimit.ID,
					loc.RateLimit.Connections.Name,
					loc.RateLimit.Connections.SharedSize)
				if !zones.Has(zone) {
					zones.Insert(zone)
				}
			}

			if loc.RateLimit.RPM.Limit > 0 {
				zone := fmt.Sprintf("limit_req_zone $limit_%s zone=%v:%vm rate=%vr/m;",
					loc.RateLimit.ID,
					loc.RateLimit.RPM.Name,
					loc.RateLimit.RPM.SharedSize,
					loc.RateLimit.RPM.Limit)
				if !zones.Has(zone) {
					zones.Insert(zone)
				}
			}

			if loc.RateLimit.RPS.Limit > 0 {
				zone := fmt.Sprintf("limit_req_zone $limit_%s zone=%v:%vm rate=%vr/s;",
					loc.RateLimit.ID,
					loc.RateLimit.RPS.Name,
					loc.RateLimit.RPS.SharedSize,
					loc.RateLimit.RPS.Limit)
				if !zones.Has(zone) {
					zones.Insert(zone)
				}
			}
		}
	}

	return zones.UnsortedList()
}

// buildRateLimit produces an array of limit_req to be used inside the Path of
// Ingress rules. The order: connections by IP first, then RPS, and RPM last.
func buildRateLimit(input interface{}) []string {
	limits := []string{}

	loc, ok := input.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return limits
	}

	if loc.RateLimit.Connections.Limit > 0 {
		limit := fmt.Sprintf("limit_conn %v %v;",
			loc.RateLimit.Connections.Name, loc.RateLimit.Connections.Limit)
		limits = append(limits, limit)
	}

	if loc.RateLimit.RPS.Limit > 0 {
		limit := fmt.Sprintf("limit_req zone=%v burst=%v nodelay;",
			loc.RateLimit.RPS.Name, loc.RateLimit.RPS.Burst)
		limits = append(limits, limit)
	}

	if loc.RateLimit.RPM.Limit > 0 {
		limit := fmt.Sprintf("limit_req zone=%v burst=%v nodelay;",
			loc.RateLimit.RPM.Name, loc.RateLimit.RPM.Burst)
		limits = append(limits, limit)
	}

	if loc.RateLimit.LimitRateAfter > 0 {
		limit := fmt.Sprintf("limit_rate_after %vk;",
			loc.RateLimit.LimitRateAfter)
		limits = append(limits, limit)
	}

	if loc.RateLimit.LimitRate > 0 {
		limit := fmt.Sprintf("limit_rate %vk;",
			loc.RateLimit.LimitRate)
		limits = append(limits, limit)
	}

	return limits
}

func isLocationInLocationList(location interface{}, rawLocationList string) bool {
	loc, ok := location.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", location)
		return false
	}

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

func isLocationAllowed(input interface{}) bool {
	loc, ok := input.(*ingress.Location)
	if !ok {
		klog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return false
	}

	return loc.Denied == nil
}

var denyPathSlugMap = map[string]string{}

// buildDenyVariable returns a nginx variable for a location in a
// server to be used in the whitelist check
// This method uses a unique id generator library to reduce the
// size of the string to be used as a variable in nginx to avoid
// issue with the size of the variable bucket size directive
func buildDenyVariable(a interface{}) string {
	l, ok := a.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", a)
		return ""
	}

	if _, ok := denyPathSlugMap[l]; !ok {
		denyPathSlugMap[l] = randomString()
	}

	return fmt.Sprintf("$deny_%v", denyPathSlugMap[l])
}

func buildUpstreamName(loc interface{}) string {
	location, ok := loc.(*ingress.Location)
	if !ok {
		klog.Errorf("expected a '*ingress.Location' type but %T was returned", loc)
		return ""
	}

	upstreamName := location.Backend

	return upstreamName
}

func buildNextUpstream(i, r interface{}) string {
	nextUpstream, ok := i.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", i)
		return ""
	}

	retryNonIdempotent, ok := r.(bool)
	if !ok {
		klog.Errorf("expected a 'bool' type but %T was returned", i)
		return ""
	}

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

	return strings.Join(nextUpstreamCodes, " ")
}

// refer to http://nginx.org/en/docs/syntax.html
// Nginx differentiates between size and offset
// offset directives support gigabytes in addition
var (
	nginxSizeRegex   = regexp.MustCompile(`^\d+[kKmM]?$`)
	nginxOffsetRegex = regexp.MustCompile(`^\d+[kKmMgG]?$`)
)

// isValidByteSize validates size units valid in nginx
// http://nginx.org/en/docs/syntax.html
func isValidByteSize(input interface{}, isOffset bool) bool {
	s, ok := input.(string)
	if !ok {
		klog.Errorf("expected an 'string' type but %T was returned", input)
		return false
	}

	s = strings.TrimSpace(s)
	if s == "" {
		klog.V(2).Info("empty byte size, hence it will not be set")
		return false
	}

	if isOffset {
		return nginxOffsetRegex.MatchString(s)
	}

	return nginxSizeRegex.MatchString(s)
}

type ingressInformation struct {
	Namespace   string
	Path        string
	Rule        string
	Service     string
	ServicePort string
	Annotations map[string]string
}

func (info *ingressInformation) Equal(other *ingressInformation) bool {
	if info.Namespace != other.Namespace {
		return false
	}
	if info.Rule != other.Rule {
		return false
	}
	if info.Service != other.Service {
		return false
	}
	if info.ServicePort != other.ServicePort {
		return false
	}
	if !reflect.DeepEqual(info.Annotations, other.Annotations) {
		return false
	}

	return true
}

func getIngressInformation(i, h, p interface{}) *ingressInformation {
	ing, ok := i.(*ingress.Ingress)
	if !ok {
		klog.Errorf("expected an '*ingress.Ingress' type but %T was returned", i)
		return &ingressInformation{}
	}

	hostname, ok := h.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", h)
		return &ingressInformation{}
	}

	ingressPath, ok := p.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", p)
		return &ingressInformation{}
	}

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

func buildForwardedFor(input interface{}) string {
	s, ok := input.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", input)
		return ""
	}

	ffh := strings.ReplaceAll(s, "-", "_")
	ffh = strings.ToLower(ffh)
	return fmt.Sprintf("$http_%v", ffh)
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

func buildAuthSignURLLocation(location, authSignURL string) string {
	hasher := sha1.New() // #nosec
	hasher.Write([]byte(location))
	hasher.Write([]byte(authSignURL))
	return "@" + hex.EncodeToString(hasher.Sum(nil))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString() string {
	b := make([]rune, 32)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			klog.Errorf("unexpected error generating random index: %v", err)
			return ""
		}
		b[i] = letters[idx.Int64()]
	}

	return string(b)
}

func buildOpentelemetry(c, s interface{}) string {
	cfg, ok := c.(config.Configuration)
	if !ok {
		klog.Errorf("expected a 'config.Configuration' type but %T was returned", c)
		return ""
	}

	servers, ok := s.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Server' type but %T was returned", s)
		return ""
	}

	if !shouldLoadOpentelemetryModule(cfg, servers) {
		return ""
	}

	buf := bytes.NewBufferString("")

	buf.WriteString("\r\n")

	if cfg.OpentelemetryOperationName != "" {
		fmt.Fprintf(buf, "opentelemetry_operation_name \"%s\";\n", cfg.OpentelemetryOperationName)
	}
	return buf.String()
}

func proxySetHeader(loc interface{}) string {
	location, ok := loc.(*ingress.Location)
	if !ok {
		klog.Errorf("expected a '*ingress.Location' type but %T was returned", loc)
		return "proxy_set_header"
	}

	if location.BackendProtocol == grpcProtocol || location.BackendProtocol == grpcsProtocol {
		return "grpc_set_header"
	}

	return "proxy_set_header"
}

// buildCustomErrorDeps is a utility function returning a struct wrapper with
// the data required to build the 'CUSTOM_ERRORS' template
func buildCustomErrorDeps(upstreamName string, errorCodes []int, enableMetrics, modsecurityEnabled bool) interface{} {
	return struct {
		UpstreamName       string
		ErrorCodes         []int
		EnableMetrics      bool
		ModsecurityEnabled bool
	}{
		UpstreamName:       upstreamName,
		ErrorCodes:         errorCodes,
		EnableMetrics:      enableMetrics,
		ModsecurityEnabled: modsecurityEnabled,
	}
}

type errorLocation struct {
	UpstreamName string
	Codes        []int
}

// buildCustomErrorLocationsPerServer is a utility function which will collect all
// custom error codes for all locations of a server block, deduplicates them,
// and returns a set which is unique by default-upstream and error code. It returns an array
// of errorLocations, each of which contain the upstream name and a list of
// error codes for that given upstream, so that sufficiently unique
// @custom error location blocks can be created in the template
func buildCustomErrorLocationsPerServer(input interface{}) []errorLocation {
	server, ok := input.(*ingress.Server)
	if !ok {
		klog.Errorf("expected a '*ingress.Server' type but %T was returned", input)
		return nil
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

	return errorLocations
}

func opentelemetryPropagateContext(location *ingress.Location) string {
	if location == nil {
		return ""
	}
	return "opentelemetry_propagate;"
}

// shouldLoadModSecurityModule determines whether or not the ModSecurity module needs to be loaded.
// First, it checks if `enable-modsecurity` is set in the ConfigMap. If it is not, it iterates over all locations to
// check if ModSecurity is enabled by the annotation `nginx.ingress.kubernetes.io/enable-modsecurity`.
func shouldLoadModSecurityModule(c, s interface{}) bool {
	cfg, ok := c.(config.Configuration)
	if !ok {
		klog.Errorf("expected a 'config.Configuration' type but %T was returned", c)
		return false
	}

	servers, ok := s.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Server' type but %T was returned", s)
		return false
	}

	// Determine if ModSecurity is enabled globally.
	if cfg.EnableModsecurity {
		return true
	}

	// If ModSecurity is not enabled globally, check if any location has it enabled via annotation.
	for _, server := range servers {
		for _, location := range server.Locations {
			if location.ModSecurity.Enable {
				return true
			}
		}
	}

	// Not enabled globally nor via annotation on a location, no need to load the module.
	return false
}

func buildHTTPListener(t, s interface{}) string {
	var out []string

	tc, ok := t.(config.TemplateConfig)
	if !ok {
		klog.Errorf("expected a 'config.TemplateConfig' type but %T was returned", t)
		return ""
	}

	hostname, ok := s.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", s)
		return ""
	}

	addrV4 := []string{""}
	if len(tc.Cfg.BindAddressIpv4) > 0 {
		addrV4 = tc.Cfg.BindAddressIpv4
	}

	co := commonListenOptions(&tc, hostname)

	out = append(out, httpListener(addrV4, co, &tc)...)

	if !tc.IsIPV6Enabled {
		return strings.Join(out, "\n")
	}

	addrV6 := []string{"[::]"}
	if len(tc.Cfg.BindAddressIpv6) > 0 {
		addrV6 = tc.Cfg.BindAddressIpv6
	}

	out = append(out, httpListener(addrV6, co, &tc)...)

	return strings.Join(out, "\n")
}

func buildHTTPSListener(t, s interface{}) string {
	var out []string

	tc, ok := t.(config.TemplateConfig)
	if !ok {
		klog.Errorf("expected a 'config.TemplateConfig' type but %T was returned", t)
		return ""
	}

	hostname, ok := s.(string)
	if !ok {
		klog.Errorf("expected a 'string' type but %T was returned", s)
		return ""
	}

	co := commonListenOptions(&tc, hostname)

	addrV4 := []string{""}
	if len(tc.Cfg.BindAddressIpv4) > 0 {
		addrV4 = tc.Cfg.BindAddressIpv4
	}

	out = append(out, httpsListener(addrV4, co, &tc)...)

	if !tc.IsIPV6Enabled {
		return strings.Join(out, "\n")
	}

	addrV6 := []string{"[::]"}
	if len(tc.Cfg.BindAddressIpv6) > 0 {
		addrV6 = tc.Cfg.BindAddressIpv6
	}

	out = append(out, httpsListener(addrV6, co, &tc)...)

	return strings.Join(out, "\n")
}

func commonListenOptions(template *config.TemplateConfig, hostname string) string {
	var out []string

	if template.Cfg.UseProxyProtocol {
		out = append(out, "proxy_protocol")
	}

	if hostname != "_" {
		return strings.Join(out, " ")
	}

	// setup options that are valid only once per port

	out = append(out, "default_server")

	if template.Cfg.ReusePort {
		out = append(out, "reuseport")
	}

	out = append(out, fmt.Sprintf("backlog=%v", template.BacklogSize))

	return strings.Join(out, " ")
}

func httpListener(addresses []string, co string, tc *config.TemplateConfig) []string {
	out := make([]string, 0)
	for _, address := range addresses {
		lo := []string{"listen"}

		if address == "" {
			lo = append(lo, fmt.Sprintf("%v", tc.ListenPorts.HTTP))
		} else {
			lo = append(lo, fmt.Sprintf("%v:%v", address, tc.ListenPorts.HTTP))
		}

		lo = append(lo, co, ";")
		out = append(out, strings.Join(lo, " "))
	}

	return out
}

func httpsListener(addresses []string, co string, tc *config.TemplateConfig) []string {
	out := make([]string, 0)
	for _, address := range addresses {
		lo := []string{"listen"}

		if tc.IsSSLPassthroughEnabled {
			if address == "" {
				lo = append(lo, fmt.Sprintf("%v", tc.ListenPorts.SSLProxy))
			} else {
				lo = append(lo, fmt.Sprintf("%v:%v", address, tc.ListenPorts.SSLProxy))
			}

			if !strings.Contains(co, "proxy_protocol") {
				lo = append(lo, "proxy_protocol")
			}
		} else {
			if address == "" {
				lo = append(lo, fmt.Sprintf("%v", tc.ListenPorts.HTTPS))
			} else {
				lo = append(lo, fmt.Sprintf("%v:%v", address, tc.ListenPorts.HTTPS))
			}
		}

		lo = append(lo, co, "ssl;")

		out = append(out, strings.Join(lo, " "))
	}

	return out
}

func buildOpentelemetryForLocation(isOTEnabled, isOTTrustSet bool, location *ingress.Location) string {
	isOTEnabledInLoc := location.Opentelemetry.Enabled
	isOTSetInLoc := location.Opentelemetry.Set

	if isOTEnabled {
		if isOTSetInLoc && !isOTEnabledInLoc {
			return "opentelemetry off;"
		}
	} else if !isOTSetInLoc || !isOTEnabledInLoc {
		return ""
	}

	opc := opentelemetryPropagateContext(location)
	if opc != "" {
		opc = fmt.Sprintf("opentelemetry on;\n%v", opc)
	}

	if location.Opentelemetry.OperationName != "" {
		opc += "\nopentelemetry_operation_name " + location.Opentelemetry.OperationName + ";"
	}

	if (!isOTTrustSet && !location.Opentelemetry.TrustSet) ||
		(location.Opentelemetry.TrustSet && !location.Opentelemetry.TrustEnabled) {
		opc += "\nopentelemetry_trust_incoming_spans off;"
	} else {
		opc += "\nopentelemetry_trust_incoming_spans on;"
	}
	return opc
}

// shouldLoadOpentelemetryModule determines whether or not the Opentelemetry module needs to be loaded.
// It checks if `enable-opentelemetry` is set in the ConfigMap.
func shouldLoadOpentelemetryModule(c, s interface{}) bool {
	cfg, ok := c.(config.Configuration)
	if !ok {
		klog.Errorf("expected a 'config.Configuration' type but %T was returned", c)
		return false
	}

	servers, ok := s.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Server' type but %T was returned", s)
		return false
	}

	if cfg.EnableOpentelemetry {
		return true
	}

	for _, server := range servers {
		for _, location := range server.Locations {
			if location.Opentelemetry.Enabled {
				return true
			}
		}
	}
	return false
}

//nolint:gocritic // Ignore passing cfg by pointer error
func buildModSecurityForLocation(cfg config.Configuration, location *ingress.Location) string {
	isMSEnabledInLoc := location.ModSecurity.Enable
	isMSEnableSetInLoc := location.ModSecurity.EnableSet
	isMSEnabled := cfg.EnableModsecurity

	if !isMSEnabled && !isMSEnabledInLoc {
		return ""
	}

	if isMSEnableSetInLoc && !isMSEnabledInLoc {
		return "modsecurity off;"
	}

	var buffer bytes.Buffer

	if !isMSEnabled {
		buffer.WriteString(`modsecurity on;
`)
	}

	if location.ModSecurity.Snippet != "" {
		buffer.WriteString(fmt.Sprintf(`modsecurity_rules '
%v
';
`, location.ModSecurity.Snippet))
	}

	if location.ModSecurity.TransactionID != "" {
		buffer.WriteString(fmt.Sprintf(`modsecurity_transaction_id "%v";
`, location.ModSecurity.TransactionID))
	}

	if !isMSEnabled && location.ModSecurity.Snippet == "" {
		buffer.WriteString(`modsecurity_rules_file /etc/nginx/modsecurity/modsecurity.conf;
`)
	}

	if !cfg.EnableOWASPCoreRules && location.ModSecurity.OWASPRules {
		buffer.WriteString(`modsecurity_rules_file /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf;
`)
	}

	return buffer.String()
}

func buildMirrorLocations(locs []*ingress.Location) string {
	var buffer bytes.Buffer

	mapped := sets.Set[string]{}

	for _, loc := range locs {
		if loc.Mirror.Source == "" || loc.Mirror.Target == "" || loc.Mirror.Host == "" {
			continue
		}

		if mapped.Has(loc.Mirror.Source) {
			continue
		}

		mapped.Insert(loc.Mirror.Source)
		buffer.WriteString(fmt.Sprintf(`location = %v {
internal;
proxy_set_header Host %v;
proxy_pass %v;
}

`, strconv.Quote(loc.Mirror.Source), strconv.Quote(loc.Mirror.Host), strconv.Quote(loc.Mirror.Target)))
	}

	return buffer.String()
}

// shouldLoadAuthDigestModule determines whether or not the ngx_http_auth_digest_module module needs to be loaded.
func shouldLoadAuthDigestModule(s interface{}) bool {
	servers, ok := s.([]*ingress.Server)
	if !ok {
		klog.Errorf("expected an '[]*ingress.Server' type but %T was returned", s)
		return false
	}

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

// buildServerName ensures wildcard hostnames are valid
func buildServerName(hostname string) string {
	if !strings.HasPrefix(hostname, "*") {
		return hostname
	}

	hostname = strings.Replace(hostname, "*.", "", 1)
	parts := strings.Split(hostname, ".")

	return `~^(?<subdomain>[\w-]+)\.` + strings.Join(parts, "\\.") + `$`
}

func buildOriginRegex(origin string) string {
	origin = regexp.QuoteMeta(origin)
	origin = strings.Replace(origin, "\\*", `[A-Za-z0-9\-]+`, 1)
	return fmt.Sprintf("(%s)", origin)
}

func buildCorsOriginRegex(corsOrigins []string) string {
	if len(corsOrigins) == 1 && corsOrigins[0] == "*" {
		return "set $http_origin *;\nset $cors 'true';"
	}

	originsRegex := "if ($http_origin ~* ("
	for i, origin := range corsOrigins {
		originTrimmed := strings.TrimSpace(origin)
		if originTrimmed != "" {
			builtOrigin := buildOriginRegex(originTrimmed)
			originsRegex += builtOrigin
			if i != len(corsOrigins)-1 {
				originsRegex += "|"
			}
		}
	}
	originsRegex += ")$ ) { set $cors 'true'; }"
	return originsRegex
}
