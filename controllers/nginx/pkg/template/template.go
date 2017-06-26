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
	"net"
	"os"
	"os/exec"
	"strings"
	text_template "text/template"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/golang/glog"

	"github.com/pborman/uuid"
	"k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/ingress"
	ing_net "k8s.io/ingress/core/pkg/net"
	"k8s.io/ingress/core/pkg/watch"
)

const (
	slash         = "/"
	defBufferSize = 65535
)

// Template ...
type Template struct {
	tmpl      *text_template.Template
	fw        watch.FileWatcher
	s         int
	tmplBuf   *bytes.Buffer
	outCmdBuf *bytes.Buffer
}

//NewTemplate returns a new Template instance or an
//error if the specified template file contains errors
func NewTemplate(file string, onChange func()) (*Template, error) {
	tmpl, err := text_template.New("nginx.tmpl").Funcs(funcMap).ParseFiles(file)
	if err != nil {
		return nil, err
	}
	fw, err := watch.NewFileWatcher(file, onChange)
	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl:      tmpl,
		fw:        fw,
		s:         defBufferSize,
		tmplBuf:   bytes.NewBuffer(make([]byte, 0, defBufferSize)),
		outCmdBuf: bytes.NewBuffer(make([]byte, 0, defBufferSize)),
	}, nil
}

// Close removes the file watcher
func (t *Template) Close() {
	t.fw.Close()
}

// Write populates a buffer using a template with NGINX configuration
// and the servers and upstreams created by Ingress rules
func (t *Template) Write(conf config.TemplateConfig) ([]byte, error) {
	defer t.tmplBuf.Reset()
	defer t.outCmdBuf.Reset()

	defer func() {
		if t.s < t.tmplBuf.Cap() {
			glog.V(2).Infof("adjusting template buffer size from %v to %v", t.s, t.tmplBuf.Cap())
			t.s = t.tmplBuf.Cap()
			t.tmplBuf = bytes.NewBuffer(make([]byte, 0, t.tmplBuf.Cap()))
			t.outCmdBuf = bytes.NewBuffer(make([]byte, 0, t.outCmdBuf.Cap()))
		}
	}()

	if glog.V(3) {
		b, err := json.Marshal(conf)
		if err != nil {
			glog.Errorf("unexpected error: %v", err)
		}
		glog.Infof("NGINX configuration: %v", string(b))
	}

	err := t.tmpl.Execute(t.tmplBuf, conf)
	if err != nil {
		return nil, err
	}

	// squeezes multiple adjacent empty lines to be single
	// spaced this is to avoid the use of regular expressions
	cmd := exec.Command("/ingress-controller/clean-nginx-conf.sh")
	cmd.Stdin = t.tmplBuf
	cmd.Stdout = t.outCmdBuf
	if err := cmd.Run(); err != nil {
		glog.Warningf("unexpected error cleaning template: %v", err)
		return t.tmplBuf.Bytes(), nil
	}

	return t.outCmdBuf.Bytes(), nil
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
		"buildLocation":            buildLocation,
		"buildAuthLocation":        buildAuthLocation,
		"buildAuthResponseHeaders": buildAuthResponseHeaders,
		"buildProxyPass":           buildProxyPass,
		"buildRateLimitZones":      buildRateLimitZones,
		"buildRateLimit":           buildRateLimit,
		"buildResolvers":           buildResolvers,
		"buildUpstreamName":        buildUpstreamName,
		"isLocationAllowed":        isLocationAllowed,
		"buildLogFormatUpstream":   buildLogFormatUpstream,
		"buildDenyVariable":        buildDenyVariable,
		"getenv":                   os.Getenv,
		"contains":                 strings.Contains,
		"hasPrefix":                strings.HasPrefix,
		"hasSuffix":                strings.HasSuffix,
		"toUpper":                  strings.ToUpper,
		"toLower":                  strings.ToLower,
		"formatIP":                 formatIP,
		"buildNextUpstream":        buildNextUpstream,
	}
)

// fomatIP will wrap IPv6 addresses in [] and return IPv4 addresses
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

// buildResolvers returns the resolvers reading the /etc/resolv.conf file
func buildResolvers(a interface{}) string {
	// NGINX need IPV6 addresses to be surrounded by brakets
	nss := a.([]net.IP)
	if len(nss) == 0 {
		return ""
	}

	r := []string{"resolver"}
	for _, ns := range nss {
		if ing_net.IsIPV6(ns) {
			r = append(r, fmt.Sprintf("[%v]", ns))
		} else {
			r = append(r, fmt.Sprintf("%v", ns))
		}
	}
	r = append(r, "valid=30s;")

	return strings.Join(r, " ")
}

// buildLocation produces the location string, if the ingress has redirects
// (specified through the ingress.kubernetes.io/rewrite-to annotation)
func buildLocation(input interface{}) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		return slash
	}

	path := location.Path
	if len(location.Redirect.Target) > 0 && location.Redirect.Target != path {
		if path == slash {
			return fmt.Sprintf("~* %s", path)
		}
		// baseuri regex will parse basename from the given location
		baseuri := `(?<baseuri>.*)`
		if !strings.HasSuffix(path, slash) {
			// Not treat the slash after "location path" as a part of baseuri
			baseuri = fmt.Sprintf(`\/?%s`, baseuri)
		}
		return fmt.Sprintf(`~* ^%s%s`, path, baseuri)
	}

	return path
}

func buildAuthLocation(input interface{}) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		return ""
	}

	if location.ExternalAuth.URL == "" {
		return ""
	}

	str := base64.URLEncoding.EncodeToString([]byte(location.Path))
	// avoid locations containing the = char
	str = strings.Replace(str, "=", "", -1)
	return fmt.Sprintf("/_external-auth-%v", str)
}

func buildAuthResponseHeaders(input interface{}) []string {
	location, ok := input.(*ingress.Location)
	res := []string{}
	if !ok {
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
		glog.Errorf("error  an ingress.buildLogFormatUpstream type but %T was returned", input)
	}

	return cfg.BuildLogFormatUpstream()
}

// buildProxyPass produces the proxy pass string, if the ingress has redirects
// (specified through the ingress.kubernetes.io/rewrite-to annotation)
// If the annotation ingress.kubernetes.io/add-base-url:"true" is specified it will
// add a base tag in the head of the response from the service
func buildProxyPass(host string, b interface{}, loc interface{}) string {
	backends := b.([]*ingress.Backend)
	location, ok := loc.(*ingress.Location)
	if !ok {
		return ""
	}

	path := location.Path
	proto := "http"

	upstreamName := location.Backend
	for _, backend := range backends {
		if backend.Name == location.Backend {
			if backend.Secure || backend.SSLPassthrough {
				proto = "https"
			}

			if isSticky(host, location, backend.SessionAffinity.CookieSessionAffinity.Locations) {
				upstreamName = fmt.Sprintf("sticky-%v", upstreamName)
			}

			break
		}
	}

	// defProxyPass returns the default proxy_pass, just the name of the upstream
	defProxyPass := fmt.Sprintf("proxy_pass %s://%s;", proto, upstreamName)
	// if the path in the ingress rule is equals to the target: no special rewrite
	if path == location.Redirect.Target {
		return defProxyPass
	}

	if path != slash && !strings.HasSuffix(path, slash) {
		path = fmt.Sprintf("%s/", path)
	}

	if len(location.Redirect.Target) > 0 {
		abu := ""
		if location.Redirect.AddBaseURL {
			// path has a slash suffix, so that it can be connected with baseuri directly
			bPath := fmt.Sprintf("%s%s", path, "$baseuri")
			abu = fmt.Sprintf(`subs_filter '<head(.*)>' '<head$1><base href="$scheme://$http_host%v">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$http_host%v">' r;
	`, bPath, bPath)
		}

		if location.Redirect.Target == slash {
			// special case redirect to /
			// ie /something to /
			return fmt.Sprintf(`
	rewrite %s(.*) /$1 break;
	rewrite %s / break;
	proxy_pass %s://%s;
	%v`, path, location.Path, proto, location.Backend, abu)
		}

		return fmt.Sprintf(`
	rewrite %s(.*) %s/$1 break;
	proxy_pass %s://%s;
	%v`, path, location.Redirect.Target, proto, location.Backend, abu)
	}

	// default proxy_pass
	return defProxyPass
}

// buildRateLimitZones produces an array of limit_conn_zone in order to allow
// rate limiting of request. Each Ingress rule could have up to two zones, one
// for connection limit by IP address and other for limiting request per second
func buildRateLimitZones(input interface{}) []string {
	zones := sets.String{}

	servers, ok := input.([]*ingress.Server)
	if !ok {
		return zones.List()
	}

	for _, server := range servers {
		for _, loc := range server.Locations {

			if loc.RateLimit.Connections.Limit > 0 {
				zone := fmt.Sprintf("limit_conn_zone $binary_remote_addr zone=%v:%vm;",
					loc.RateLimit.Connections.Name,
					loc.RateLimit.Connections.SharedSize)
				if !zones.Has(zone) {
					zones.Insert(zone)
				}
			}

			if loc.RateLimit.RPS.Limit > 0 {
				zone := fmt.Sprintf("limit_req_zone $binary_remote_addr zone=%v:%vm rate=%vr/s;",
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
// Ingress rules. The order: connections by IP first and RPS next.
func buildRateLimit(input interface{}) []string {
	limits := []string{}

	loc, ok := input.(*ingress.Location)
	if !ok {
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

	return limits
}

func isLocationAllowed(input interface{}) bool {
	loc, ok := input.(*ingress.Location)
	if !ok {
		glog.Errorf("expected an ingress.Location type but %T was returned", input)
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
	l := a.(string)

	if _, ok := denyPathSlugMap[l]; !ok {
		denyPathSlugMap[l] = uuid.New()
	}

	return fmt.Sprintf("$deny_%v", denyPathSlugMap[l])
}

func buildUpstreamName(host string, b interface{}, loc interface{}) string {
	backends := b.([]*ingress.Backend)
	location, ok := loc.(*ingress.Location)
	if !ok {
		return ""
	}

	upstreamName := location.Backend

	for _, backend := range backends {
		if backend.Name == location.Backend {
			if backend.SessionAffinity.AffinityType == "cookie" &&
				isSticky(host, location, backend.SessionAffinity.CookieSessionAffinity.Locations) {
				upstreamName = fmt.Sprintf("sticky-%v", upstreamName)
			}

			break
		}
	}

	return upstreamName
}

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

func buildNextUpstream(input interface{}) string {
	nextUpstream, ok := input.(string)
	if !ok {
		glog.Errorf("expected an string type but %T was returned", input)
	}

	parts := strings.Split(nextUpstream, " ")

	nextUpstreamCodes := make([]string, 0, len(parts))
	for _, v := range parts {
		if v != "" && v != "non_idempotent" {
			nextUpstreamCodes = append(nextUpstreamCodes, v)
		}
	}

	return strings.Join(nextUpstreamCodes, " ")
}