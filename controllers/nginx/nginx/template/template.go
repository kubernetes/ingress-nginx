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
	"regexp"
	"strings"
	text_template "text/template"

	"github.com/fatih/structs"
	"github.com/golang/glog"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/ingress"
	"k8s.io/kubernetes/pkg/util/sysctl"
)

const (
	slash = "/"
)

var (
	camelRegexp = regexp.MustCompile("[0-9A-Za-z]+")

	funcMap = text_template.FuncMap{
		"empty": func(input interface{}) bool {
			check, ok := input.(string)
			if ok {
				return len(check) == 0
			}

			return true
		},
		"buildLocation":       buildLocation,
		"buildAuthLocation":   buildAuthLocation,
		"buildProxyPass":      buildProxyPass,
		"buildRateLimitZones": buildRateLimitZones,
		"buildRateLimit":      buildRateLimit,
		"contains":            strings.Contains,
		"hasPrefix":           strings.HasPrefix,
		"hasSuffix":           strings.HasSuffix,
		"toUpper":             strings.ToUpper,
		"toLower":             strings.ToLower,
	}
)

// Template ...
type Template struct {
	tmpl *text_template.Template
	fw   fileWatcher
}

//NewTemplate returns a new Template instance or an
//error if the specified template file contains errors
func NewTemplate(file string, onChange func()) (*Template, error) {
	tmpl, err := text_template.New("nginx.tmpl").Funcs(funcMap).ParseFiles(file)
	if err != nil {
		return nil, err
	}
	fw, err := newFileWatcher(file, onChange)
	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl: tmpl,
		fw:   fw,
	}, nil
}

// Close removes the file watcher
func (t *Template) Close() {
	t.fw.close()
}

// Write populates a buffer using a template with NGINX configuration
// and the servers and upstreams created by Ingress rules
func (t *Template) Write(
	cfg config.Configuration,
	ingressCfg ingress.Configuration,
	isValidTemplate func([]byte) error) ([]byte, error) {
	var longestName int
	var serverNames int
	for _, srv := range ingressCfg.Servers {
		serverNames += len([]byte(srv.Name))
		if longestName < len(srv.Name) {
			longestName = len(srv.Name)
		}
	}

	// NGINX cannot resize the has tables used to store server names.
	// For this reason we check if the defined size defined is correct
	// for the FQDN defined in the ingress rules adjusting the value
	// if is required.
	// https://trac.nginx.org/nginx/ticket/352
	// https://trac.nginx.org/nginx/ticket/631
	nameHashBucketSize := nextPowerOf2(longestName)
	if nameHashBucketSize > cfg.ServerNameHashBucketSize {
		glog.V(3).Infof("adjusting ServerNameHashBucketSize variable from %v to %v",
			cfg.ServerNameHashBucketSize, nameHashBucketSize)
		cfg.ServerNameHashBucketSize = nameHashBucketSize
	}
	serverNameHashMaxSize := nextPowerOf2(serverNames)
	if serverNameHashMaxSize > cfg.ServerNameHashMaxSize {
		glog.V(3).Infof("adjusting ServerNameHashMaxSize variable from %v to %v",
			cfg.ServerNameHashMaxSize, serverNameHashMaxSize)
		cfg.ServerNameHashMaxSize = serverNameHashMaxSize
	}

	conf := make(map[string]interface{})
	conf["backlogSize"] = sysctlSomaxconn()
	conf["upstreams"] = ingressCfg.Upstreams
	conf["servers"] = ingressCfg.Servers
	conf["tcpUpstreams"] = ingressCfg.TCPUpstreams
	conf["udpUpstreams"] = ingressCfg.UDPUpstreams
	conf["defResolver"] = cfg.Resolver
	conf["sslDHParam"] = cfg.SSLDHParam
	conf["customErrors"] = len(cfg.CustomHTTPErrors) > 0
	conf["cfg"] = fixKeyNames(structs.Map(cfg))

	if glog.V(3) {
		b, err := json.Marshal(conf)
		if err != nil {
			glog.Errorf("unexpected error: %v", err)
		}
		glog.Infof("NGINX configuration: %v", string(b))
	}

	buffer := new(bytes.Buffer)
	err := t.tmpl.Execute(buffer, conf)
	if err != nil {
		glog.V(3).Infof("%v", string(buffer.Bytes()))
		return nil, err
	}

	err = isValidTemplate(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func fixKeyNames(data map[string]interface{}) map[string]interface{} {
	fixed := make(map[string]interface{})
	for k, v := range data {
		fixed[toCamelCase(k)] = v
	}

	return fixed
}

func toCamelCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelRegexp.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = bytes.Title(val)
		}
	}
	return string(bytes.Join(chunks, nil))
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
		return fmt.Sprintf("~* %s", path)
	}

	return path
}

func buildAuthLocation(input interface{}) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		return ""
	}

	if location.ExternalAuthURL.URL == "" {
		return ""
	}

	str := base64.URLEncoding.EncodeToString([]byte(location.Path))
	// avoid locations containing the = char
	str = strings.Replace(str, "=", "", -1)
	return fmt.Sprintf("/_external-auth-%v", str)
}

// buildProxyPass produces the proxy pass string, if the ingress has redirects
// (specified through the ingress.kubernetes.io/rewrite-to annotation)
// If the annotation ingress.kubernetes.io/add-base-url:"true" is specified it will
// add a base tag in the head of the response from the service
func buildProxyPass(input interface{}) string {
	location, ok := input.(*ingress.Location)
	if !ok {
		return ""
	}

	path := location.Path

	proto := "http"
	if location.SecureUpstream {
		proto = "https"
	}
	// defProxyPass returns the default proxy_pass, just the name of the upstream
	defProxyPass := fmt.Sprintf("proxy_pass %s://%s;", proto, location.Upstream.Name)
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
			bPath := location.Redirect.Target
			if !strings.HasSuffix(bPath, slash) {
				bPath = fmt.Sprintf("%s/", bPath)
			}

			abu = fmt.Sprintf(`subs_filter '<head(.*)>' '<head$1><base href="$scheme://$server_name%v">' r;
	subs_filter '<HEAD(.*)>' '<HEAD$1><base href="$scheme://$server_name%v">' r;
	`, bPath, bPath)
		}

		if location.Redirect.Target == slash {
			// special case redirect to /
			// ie /something to /
			return fmt.Sprintf(`
	rewrite %s(.*) /$1 break;
	rewrite %s / break;
	proxy_pass %s://%s;
	%v`, path, location.Path, proto, location.Upstream.Name, abu)
		}

		return fmt.Sprintf(`
	rewrite %s(.*) %s/$1 break;
	proxy_pass %s://%s;
	%v`, path, location.Redirect.Target, proto, location.Upstream.Name, abu)
	}

	// default proxy_pass
	return defProxyPass
}

// buildRateLimitZones produces an array of limit_conn_zone in order to allow
// rate limiting of request. Each Ingress rule could have up to two zones, one
// for connection limit by IP address and other for limiting request per second
func buildRateLimitZones(input interface{}) []string {
	zones := []string{}

	servers, ok := input.([]*ingress.Server)
	if !ok {
		return zones
	}

	for _, server := range servers {
		for _, loc := range server.Locations {

			if loc.RateLimit.Connections.Limit > 0 {
				zone := fmt.Sprintf("limit_conn_zone $binary_remote_addr zone=%v:%vm;",
					loc.RateLimit.Connections.Name,
					loc.RateLimit.Connections.SharedSize)
				zones = append(zones, zone)
			}

			if loc.RateLimit.RPS.Limit > 0 {
				zone := fmt.Sprintf("limit_conn_zone $binary_remote_addr zone=%v:%vm rate=%vr/s;",
					loc.RateLimit.Connections.Name,
					loc.RateLimit.Connections.SharedSize,
					loc.RateLimit.Connections.Limit)
				zones = append(zones, zone)
			}
		}
	}

	return zones
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
			loc.RateLimit.Connections.Name, loc.RateLimit.Connections.Burst)
		limits = append(limits, limit)
	}

	return limits
}

// sysctlSomaxconn returns the value of net.core.somaxconn, i.e.
// maximum number of connections that can be queued for acceptance
// http://nginx.org/en/docs/http/ngx_http_core_module.html#listen
func sysctlSomaxconn() int {
	maxConns, err := sysctl.New().GetSysctl("net/core/somaxconn")
	if err != nil || maxConns < 512 {
		glog.Warningf("system net.core.somaxconn=%v. Using NGINX default (511)", maxConns)
		return 511
	}

	return maxConns
}

// http://graphics.stanford.edu/~seander/bithacks.html#RoundUpPowerOf2
// https://play.golang.org/p/TVSyCcdxUh
func nextPowerOf2(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++

	return v
}
