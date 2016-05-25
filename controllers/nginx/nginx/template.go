/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package nginx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/fatih/structs"
	"github.com/golang/glog"
)

const (
	slash = "/"
)

var (
	camelRegexp = regexp.MustCompile("[0-9A-Za-z]+")
	tmplPath    = "/etc/nginx/template/nginx.tmpl"

	funcMap = template.FuncMap{
		"empty": func(input interface{}) bool {
			check, ok := input.(string)
			if ok {
				return len(check) == 0
			}

			return true
		},
		"buildLocation":  buildLocation,
		"buildProxyPass": buildProxyPass,
	}
)

func (ngx *Manager) loadTemplate() {
	tmpl, _ := template.New("nginx.tmpl").Funcs(funcMap).ParseFiles(tmplPath)
	ngx.template = tmpl
}

func (ngx *Manager) writeCfg(cfg NginxConfiguration, ingressCfg IngressConfig) (bool, error) {
	conf := make(map[string]interface{})
	conf["upstreams"] = ingressCfg.Upstreams
	conf["servers"] = ingressCfg.Servers
	conf["tcpUpstreams"] = ingressCfg.TCPUpstreams
	conf["udpUpstreams"] = ingressCfg.UDPUpstreams
	conf["defResolver"] = ngx.defResolver
	conf["sslDHParam"] = ngx.sslDHParam
	conf["customErrors"] = len(cfg.CustomHTTPErrors) > 0
	conf["cfg"] = fixKeyNames(structs.Map(cfg))

	if glog.V(3) {
		b, err := json.Marshal(conf)
		if err != nil {
			fmt.Println("error:", err)
		}
		glog.Infof("NGINX configuration: %v", string(b))
	}

	buffer := new(bytes.Buffer)
	err := ngx.template.Execute(buffer, conf)
	if err != nil {
		glog.Infof("NGINX error: %v", err)
		return false, err
	}

	changed, err := ngx.needsReload(buffer)
	if err != nil {
		return false, err
	}

	return changed, nil
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

func buildLocation(input interface{}) string {
	location, ok := input.(*Location)
	if !ok {
		return slash
	}

	path := location.Path
	if len(location.Redirect.To) > 0 && location.Redirect.To != path {
		// if path != slash && !strings.HasSuffix(path, slash) {
		// 	path = fmt.Sprintf("%s/", path)
		// }
		return fmt.Sprintf("~* %s", path)
	}

	return path
}

func buildProxyPass(input interface{}) string {
	location, ok := input.(*Location)
	if !ok {
		return ""
	}

	path := location.Path

	if path == location.Redirect.To {
		return fmt.Sprintf("proxy_pass http://%s;", location.Upstream.Name)
	}

	if path != slash && !strings.HasSuffix(path, slash) {
		path = fmt.Sprintf("%s/", path)
	}

	if len(location.Redirect.To) > 0 {
		rc := ""
		if location.Redirect.Rewrite {
			rc = fmt.Sprintf(`sub_filter '<head(.*)>' '<head$1><base href="$scheme://$server_name%v">';
sub_filter_once off;`, location.Path)
		}

		if location.Redirect.To == slash {
			// special case redirect to /
			// ie /something to /
			return fmt.Sprintf(`rewrite %s(.*) /$1 break;
proxy_pass http://%s;
%v`, path, location.Upstream.Name, rc)
		}

		return fmt.Sprintf(`rewrite %s(.*) %s/$1 break;
proxy_pass http://%s;
%v`, path, location.Redirect.To, location.Upstream.Name, rc)
	}

	// default proxy_pass
	return fmt.Sprintf("proxy_pass http://%s;", location.Upstream.Name)
}
