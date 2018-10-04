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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	text_template "text/template"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/influxdb"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ratelimit"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
)

const (
	slash         = "/"
	nonIdempotent = "non_idempotent"
	defBufferSize = 65535
)

// Template ...
type Template struct {
	tmpl *text_template.Template
	//fw   watch.FileWatcher
	bp *BufferPool
}

//NewTemplate returns a new Template instance or an
//error if the specified template file contains errors
func NewTemplate(file string, fs file.Filesystem) (*Template, error) {
	data, err := fs.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "unexpected error reading template %v", file)
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

// Write populates a buffer using a template with NGINX configuration
// and the servers and upstreams created by Ingress rules
func (t *Template) Write(conf config.TemplateConfig) ([]byte, error) {
	tmplBuf := t.bp.Get()
	defer t.bp.Put(tmplBuf)

	outCmdBuf := t.bp.Get()
	defer t.bp.Put(outCmdBuf)

	if glog.V(3) {
		b, err := json.Marshal(conf)
		if err != nil {
			glog.Errorf("unexpected error: %v", err)
		}
		glog.Infof("NGINX configuration: %v", string(b))
	}

	err := t.tmpl.Execute(tmplBuf, conf)
	if err != nil {
		return nil, err
	}

	// squeezes multiple adjacent empty lines to be single
	// spaced this is to avoid the use of regular expressions
	cmd := exec.Command("/ingress-controller/clean-nginx-conf.sh")
	cmd.Stdin = tmplBuf
	cmd.Stdout = outCmdBuf
	if err := cmd.Run(); err != nil {
		glog.Warningf("unexpected error cleaning template: %v", err)
		return tmplBuf.Bytes(), nil
	}

	return outCmdBuf.Bytes(), nil
}

var (
	funcMap = text_template.FuncMap{
		"empty": func(input interface{}) bool {
			check, ok := input.(string)
			if ok {
				return len(check) == 0
			}
			return true
		},
		"shouldConfigureLuaRestyWAF": shouldConfigureLuaRestyWAF,
		"buildLuaSharedDictionaries": buildLuaSharedDictionaries,
		"buildLocation":              buildLocation,
		"buildAuthLocation":          buildAuthLocation,
		"buildAuthResponseHeaders":   buildAuthResponseHeaders,
		"buildLoadBalancingConfig":   buildLoadBalancingConfig,
		"buildProxyPass":             buildProxyPass,
		"filterRateLimits":           filterRateLimits,
		"buildRateLimitZones":        buildRateLimitZones,
		"buildRateLimit":             buildRateLimit,
		"buildResolversForLua":       buildResolversForLua,
		"buildResolvers":             buildResolvers,
		"buildUpstreamName":          buildUpstreamName,
		"isLocationInLocationList":   isLocationInLocationList,
		"isLocationAllowed":          isLocationAllowed,
		"buildLogFormatUpstream":     buildLogFormatUpstream,
		"buildDenyVariable":          buildDenyVariable,
		"getenv":                     os.Getenv,
		"contains":                   strings.Contains,
		"hasPrefix":                  strings.HasPrefix,
		"hasSuffix":                  strings.HasSuffix,
		"trimSpace":                  strings.TrimSpace,
		"toUpper":                    strings.ToUpper,
		"toLower":                    strings.ToLower,
		"formatIP":                   formatIP,
		"buildNextUpstream":          buildNextUpstream,
		"getIngressInformation":      getIngressInformation,
		"serverConfig": func(all config.TemplateConfig, server *ingress.Server) interface{} {
			return struct{ First, Second interface{} }{all, server}
		},
		"isValidClientBodyBufferSize": isValidClientBodyBufferSize,
		"buildForwardedFor":           buildForwardedFor,
		"buildAuthSignURL":            buildAuthSignURL,
		"buildOpentracing":            buildOpentracing,
		"proxySetHeader":              proxySetHeader,
		"buildInfluxDB":               buildInfluxDB,
		"enforceRegexModifier":        enforceRegexModifier,
		"stripLocationModifer":        stripLocationModifer,
	}
)

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

func shouldConfigureLuaRestyWAF(disableLuaRestyWAF bool, mode string) bool {
	if !disableLuaRestyWAF && len(mode) > 0 {
		return true
	}

	return false
}

func buildLuaSharedDictionaries(s interface{}, dynamicConfigurationEnabled bool, disableLuaRestyWAF bool) string {
	servers, ok := s.([]*ingress.Server)
	if !ok {
		glog.Errorf("expected an '[]*ingress.Server' type but %T was returned", s)
		return ""
	}

	out := []string{}

	if dynamicConfigurationEnabled {
		out = append(out,
			"lua_shared_dict configuration_data 5M",
			"lua_shared_dict certificate_data 16M",
			"lua_shared_dict locks 512k",
			"lua_shared_dict balancer_ewma 1M",
			"lua_shared_dict balancer_ewma_last_touched_at 1M",
			"lua_shared_dict sticky_sessions 1M",
		)
	}

	if !disableLuaRestyWAF {
		luaRestyWAFEnabled := func() bool {
			for _, server := range servers {
				for _, location := range server.Locations {
					if len(location.LuaRestyWAF.Mode) > 0 {
						return true
					}
				}
			}
			return false
		}()
		if luaRestyWAFEnabled {
			out = append(out, "lua_shared_dict waf_storage 64M")
		}
	}

	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, ";\n\r") + ";"
}

func buildResolversForLua(res interface{}, disableIpv6 interface{}) string {
	nss, ok := res.([]net.IP)
	if !ok {
		glog.Errorf("expected a '[]net.IP' type but %T was returned", res)
		return ""
	}
	no6, ok := disableIpv6.(bool)
	if !ok {
		glog.Errorf("expected a 'bool' type but %T was returned", disableIpv6)
		return ""
	}

	if len(nss) == 0 {
		return ""
	}

	r := []string{}
	for _, ns := range nss {
		if ing_net.IsIPV6(ns) && no6 {
			continue
		}
		r = append(r, fmt.Sprintf("\"%v\"", ns))
	}

	return strings.Join(r, ", ")
}

// buildResolvers returns the resolvers reading the /etc/resolv.conf file
func buildResolvers(res interface{}, disableIpv6 interface{}) string {
	// NGINX need IPV6 addresses to be surrounded by brackets
	nss, ok := res.([]net.IP)
	if !ok {
		glog.Errorf("expected a '[]net.IP' type but %T was returned", res)
		return ""
	}
	no6, ok := disableIpv6.(bool)
	if !ok {
		glog.Errorf("expected a 'bool' type but %T was returned", disableIpv6)
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
			r = append(r, fmt.Sprintf("%v", ns))
		}
	}
	r = append(r, "valid=30s")

	if no6 {
		r = append(r, "ipv6=off")
	}

	return strings.Join(r, " ") + ";"
}

func needsRewrite(location *ingress.Location) bool {
	if len(location.Rewrite.Target) > 0 && location.Rewrite.Target != location.Path {
		return true
	}
	return false
}

func stripLocationModifer(path string) string {
	return strings.TrimLeft(path, "~* ")
}

// enforceRegexModifier checks if the "rewrite-target" or "use-regex" annotation
// is used on any location path within a server
func enforceRegexModifier(input interface{}) bool {
	locations, ok := input.([]*ingress.Location)
	if !ok {
		glog.Errorf("expected an '[]*ingress.Location' type but %T was returned", input)
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
		glog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return slash
	}

	path := location.Path
	if needsRewrite(location) {
		if path == slash {
			return fmt.Sprintf("~* ^%s", path)
		}
		// baseuri regex will parse basename from the given location
		baseuri := `(?<baseuri>.*)`
		if !strings.HasSuffix(path, slash) {
			// Not treat the slash after "location path" as a part of baseuri
			baseuri = fmt.Sprintf(`\/?%s`, baseuri)
		}
		return fmt.Sprintf(`~* "^%s%s"`, path, baseuri)
	}

	if enforceRegex {
		return fmt.Sprintf(`~* "^%s"`, path)
	}
	return path
}

func buildAuthLocation(input interface{}) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		glog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return ""
	}

	if location.ExternalAuth.URL == "" {
		return ""
	}

	str := base64.URLEncoding.EncodeToString([]byte(location.Path))
	// removes "=" after encoding
	str = strings.Replace(str, "=", "", -1)
	return fmt.Sprintf("/_external-auth-%v", str)
}

func buildAuthResponseHeaders(input interface{}) []string {
	location, ok := input.(*ingress.Location)
	res := []string{}
	if !ok {
		glog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return res
	}

	if len(location.ExternalAuth.ResponseHeaders) == 0 {
		return res
	}

	for i, h := range location.ExternalAuth.ResponseHeaders {
		hvar := strings.ToLower(h)
		hvar = strings.NewReplacer("-", "_").Replace(hvar)
		res = append(res, fmt.Sprintf("auth_request_set $authHeader%v $upstream_http_%v;", i, hvar))
		res = append(res, fmt.Sprintf("proxy_set_header '%v' $authHeader%v;", h, i))
	}
	return res
}

func buildLogFormatUpstream(input interface{}) string {
	cfg, ok := input.(config.Configuration)
	if !ok {
		glog.Errorf("expected a 'config.Configuration' type but %T was returned", input)
		return ""
	}

	return cfg.BuildLogFormatUpstream()
}

func buildLoadBalancingConfig(b interface{}, fallbackLoadBalancing string) string {
	backend, ok := b.(*ingress.Backend)
	if !ok {
		glog.Errorf("expected an '*ingress.Backend' type but %T was returned", b)
		return ""
	}

	if backend.UpstreamHashBy != "" {
		return fmt.Sprintf("hash %s consistent;", backend.UpstreamHashBy)
	}

	if backend.LoadBalancing != "" {
		if backend.LoadBalancing == "round_robin" {
			return ""
		}
		return fmt.Sprintf("%s;", backend.LoadBalancing)
	}

	if fallbackLoadBalancing == "round_robin" || fallbackLoadBalancing == "" {
		return ""
	}

	return fmt.Sprintf("%s;", fallbackLoadBalancing)
}

// buildProxyPass produces the proxy pass string, if the ingress has redirects
// (specified through the nginx.ingress.kubernetes.io/rewrite-target annotation)
// If the annotation nginx.ingress.kubernetes.io/add-base-url:"true" is specified it will
// add a base tag in the head of the response from the service
func buildProxyPass(host string, b interface{}, loc interface{}, dynamicConfigurationEnabled bool) string {
	backends, ok := b.([]*ingress.Backend)
	if !ok {
		glog.Errorf("expected an '[]*ingress.Backend' type but %T was returned", b)
		return ""
	}

	location, ok := loc.(*ingress.Location)
	if !ok {
		glog.Errorf("expected a '*ingress.Location' type but %T was returned", loc)
		return ""
	}

	path := location.Path
	proto := "http://"

	proxyPass := "proxy_pass"

	switch location.BackendProtocol {
	case "HTTPS":
		proto = "https://"
	case "GRPC":
		proto = "grpc://"
		proxyPass = "grpc_pass"
	case "GRPCS":
		proto = "grpcs://"
		proxyPass = "grpc_pass"
	case "AJP":
		proto = ""
		proxyPass = "ajp_pass"
	}

	// TODO: Remove after the deprecation of grpc-backend annotation
	if location.GRPC {
		proxyPass = "grpc_pass"
		proto = "grpc://"
	}

	upstreamName := "upstream_balancer"

	if !dynamicConfigurationEnabled {
		upstreamName = location.Backend
	}

	for _, backend := range backends {
		if backend.Name == location.Backend {
			if backend.Secure || backend.SSLPassthrough {
				// TODO: Remove after the deprecation of secure-backend annotation
				proto = "https://"
				// TODO: Remove after the deprecation of grpc-backend annotation
				if location.GRPC {
					proto = "grpcs://"
				}
			}

			if !dynamicConfigurationEnabled && isSticky(host, location, backend.SessionAffinity.CookieSessionAffinity.Locations) {
				upstreamName = fmt.Sprintf("sticky-%v", upstreamName)
			}

			break
		}
	}

	// defProxyPass returns the default proxy_pass, just the name of the upstream
	defProxyPass := fmt.Sprintf("%v %s%s;", proxyPass, proto, upstreamName)

	// if the path in the ingress rule is equals to the target: no special rewrite
	if path == location.Rewrite.Target {
		return defProxyPass
	}

	if !strings.HasSuffix(path, slash) {
		path = fmt.Sprintf("%s/", path)
	}

	if len(location.Rewrite.Target) > 0 {
		var abu string
		var xForwardedPrefix string

		if location.Rewrite.AddBaseURL {
			bPath := fmt.Sprintf("%s$escaped_base_uri", path)
			regex := `(<(?:H|h)(?:E|e)(?:A|a)(?:D|d)(?:[^">]|"[^"]*")*>)`
			scheme := "$scheme"

			if len(location.Rewrite.BaseURLScheme) > 0 {
				scheme = location.Rewrite.BaseURLScheme
			}

			abu = fmt.Sprintf(`
set_escape_uri $escaped_base_uri $baseuri;
subs_filter '%v' '$1<base href="%v://$http_host%v">' ro;
`, regex, scheme, bPath)
		}

		if location.XForwardedPrefix {
			xForwardedPrefix = fmt.Sprintf("proxy_set_header X-Forwarded-Prefix \"%s\";\n", path)
		}

		if location.Rewrite.Target == slash {
			// special case redirect to /
			// ie /something to /
			return fmt.Sprintf(`
rewrite "(?i)%s(.*)" /$1 break;
rewrite "(?i)%s$" / break;
%v%v %s%s;
%v`, path, location.Path, xForwardedPrefix, proxyPass, proto, upstreamName, abu)
		}

		return fmt.Sprintf(`
rewrite "(?i)%s(.*)" %s/$1 break;
rewrite "(?i)%s$" %s/ break;
%v%v %s%s;
%v`, path, location.Rewrite.Target, location.Path, location.Rewrite.Target, xForwardedPrefix, proxyPass, proto, upstreamName, abu)
	}

	// default proxy_pass
	return defProxyPass
}

// TODO: Needs Unit Tests
func filterRateLimits(input interface{}) []ratelimit.Config {
	ratelimits := []ratelimit.Config{}
	found := sets.String{}

	servers, ok := input.([]*ingress.Server)
	if !ok {
		glog.Errorf("expected a '[]ratelimit.RateLimit' type but %T was returned", input)
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

// TODO: Needs Unit Tests
// buildRateLimitZones produces an array of limit_conn_zone in order to allow
// rate limiting of request. Each Ingress rule could have up to three zones, one
// for connection limit by IP address, one for limiting requests per minute, and
// one for limiting requests per second.
func buildRateLimitZones(input interface{}) []string {
	zones := sets.String{}

	servers, ok := input.([]*ingress.Server)
	if !ok {
		glog.Errorf("expected a '[]*ingress.Server' type but %T was returned", input)
		return zones.List()
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

	return zones.List()
}

// buildRateLimit produces an array of limit_req to be used inside the Path of
// Ingress rules. The order: connections by IP first, then RPS, and RPM last.
func buildRateLimit(input interface{}) []string {
	limits := []string{}

	loc, ok := input.(*ingress.Location)
	if !ok {
		glog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
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
		glog.Errorf("expected an '*ingress.Location' type but %T was returned", location)
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
		glog.Errorf("expected an '*ingress.Location' type but %T was returned", input)
		return false
	}

	return loc.Denied == nil
}

var (
	denyPathSlugMap = map[string]string{}
)

// buildDenyVariable returns a nginx variable for a location in a
// server to be used in the whitelist check
// This method uses a unique id generator library to reduce the
// size of the string to be used as a variable in nginx to avoid
// issue with the size of the variable bucket size directive
func buildDenyVariable(a interface{}) string {
	l, ok := a.(string)
	if !ok {
		glog.Errorf("expected a 'string' type but %T was returned", a)
		return ""
	}

	if _, ok := denyPathSlugMap[l]; !ok {
		denyPathSlugMap[l] = randomString()
	}

	return fmt.Sprintf("$deny_%v", denyPathSlugMap[l])
}

func buildUpstreamName(host string, b interface{}, loc interface{}, dynamicConfigurationEnabled bool) string {

	backends, ok := b.([]*ingress.Backend)
	if !ok {
		glog.Errorf("expected an '[]*ingress.Backend' type but %T was returned", b)
		return ""
	}

	location, ok := loc.(*ingress.Location)
	if !ok {
		glog.Errorf("expected a '*ingress.Location' type but %T was returned", loc)
		return ""
	}

	upstreamName := location.Backend

	if !dynamicConfigurationEnabled {
		for _, backend := range backends {
			if backend.Name == location.Backend {
				if backend.SessionAffinity.AffinityType == "cookie" &&
					isSticky(host, location, backend.SessionAffinity.CookieSessionAffinity.Locations) {
					upstreamName = fmt.Sprintf("sticky-%v", upstreamName)
				}

				break
			}
		}
	}

	return upstreamName
}

// TODO: Needs Unit Tests
func isSticky(host string, loc *ingress.Location, stickyLocations map[string][]string) bool {
	if _, ok := stickyLocations[host]; ok {
		for _, sl := range stickyLocations[host] {
			if sl == loc.Path {
				return true
			}
		}
	}

	return false
}

func buildNextUpstream(i, r interface{}) string {
	nextUpstream, ok := i.(string)
	if !ok {
		glog.Errorf("expected a 'string' type but %T was returned", i)
		return ""
	}

	retryNonIdempotent := r.(bool)

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

func isValidClientBodyBufferSize(input interface{}) bool {
	s, ok := input.(string)
	if !ok {
		glog.Errorf("expected an 'string' type but %T was returned", input)
		return false
	}

	if s == "" {
		return false
	}

	_, err := strconv.Atoi(s)
	if err != nil {
		sLowercase := strings.ToLower(s)

		check := strings.TrimSuffix(sLowercase, "k")
		_, err := strconv.Atoi(check)
		if err == nil {
			return true
		}

		mCheck := strings.TrimSuffix(sLowercase, "m")
		_, err = strconv.Atoi(mCheck)
		if err == nil {
			return true
		}

		glog.Errorf("client-body-buffer-size '%v' was provided in an incorrect format, hence it will not be set.", s)
		return false
	}

	return true
}

type ingressInformation struct {
	Namespace   string
	Rule        string
	Service     string
	Annotations map[string]string
}

func getIngressInformation(i, p interface{}) *ingressInformation {
	ing, ok := i.(*extensions.Ingress)
	if !ok {
		glog.Errorf("expected an '*extensions.Ingress' type but %T was returned", i)
		return &ingressInformation{}
	}

	path, ok := p.(string)
	if !ok {
		glog.Errorf("expected a 'string' type but %T was returned", p)
		return &ingressInformation{}
	}

	if ing == nil {
		return &ingressInformation{}
	}

	info := &ingressInformation{
		Namespace:   ing.GetNamespace(),
		Rule:        ing.GetName(),
		Annotations: ing.Annotations,
	}

	if ing.Spec.Backend != nil {
		info.Service = ing.Spec.Backend.ServiceName
	}

	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for _, rPath := range rule.HTTP.Paths {
			if path == rPath.Path {
				info.Service = rPath.Backend.ServiceName
				return info
			}
		}
	}

	return info
}

func buildForwardedFor(input interface{}) string {
	s, ok := input.(string)
	if !ok {
		glog.Errorf("expected a 'string' type but %T was returned", input)
		return ""
	}

	ffh := strings.Replace(s, "-", "_", -1)
	ffh = strings.ToLower(ffh)
	return fmt.Sprintf("$http_%v", ffh)
}

func buildAuthSignURL(input interface{}) string {
	s, ok := input.(string)
	if !ok {
		glog.Errorf("expected an 'string' type but %T was returned", input)
		return ""
	}

	u, _ := url.Parse(s)
	q := u.Query()
	if len(q) == 0 {
		return fmt.Sprintf("%v?rd=$pass_access_scheme://$http_host$escaped_request_uri", s)
	}

	if q.Get("rd") != "" {
		return s
	}

	return fmt.Sprintf("%v&rd=$pass_access_scheme://$http_host$escaped_request_uri", s)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func buildOpentracing(input interface{}) string {
	cfg, ok := input.(config.Configuration)
	if !ok {
		glog.Errorf("expected a 'config.Configuration' type but %T was returned", input)
		return ""
	}

	if !cfg.EnableOpentracing {
		return ""
	}

	buf := bytes.NewBufferString("")
	if cfg.ZipkinCollectorHost != "" {
		buf.WriteString("opentracing_load_tracer /usr/local/lib/libzipkin_opentracing.so /etc/nginx/opentracing.json;")
	} else if cfg.JaegerCollectorHost != "" {
		buf.WriteString("opentracing_load_tracer /usr/local/lib/libjaegertracing_plugin.so 	 /etc/nginx/opentracing.json;")
	}

	buf.WriteString("\r\n")

	return buf.String()
}

// buildInfluxDB produces the single line configuration
// needed by the InfluxDB module to send request's metrics
// for the current resource
func buildInfluxDB(input interface{}) string {
	cfg, ok := input.(influxdb.Config)
	if !ok {
		glog.Errorf("expected an 'influxdb.Config' type but %T was returned", input)
		return ""
	}

	if !cfg.InfluxDBEnabled {
		return ""
	}

	return fmt.Sprintf(
		"influxdb server_name=%s host=%s port=%s measurement=%s enabled=true;",
		cfg.InfluxDBServerName,
		cfg.InfluxDBHost,
		cfg.InfluxDBPort,
		cfg.InfluxDBMeasurement,
	)
}

func proxySetHeader(loc interface{}) string {
	location, ok := loc.(*ingress.Location)
	if !ok {
		glog.Errorf("expected a '*ingress.Location' type but %T was returned", loc)
		return "proxy_set_header"
	}

	if location.GRPC || location.BackendProtocol == "GRPC" || location.BackendProtocol == "GRPCS" {
		return "grpc_set_header"
	}

	return "proxy_set_header"
}
