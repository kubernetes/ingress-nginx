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
	"text/template"

	"github.com/fatih/structs"
	"github.com/golang/glog"
)

var funcMap = template.FuncMap{
	"empty": func(input interface{}) bool {
		check, ok := input.(string)
		if ok {
			return len(check) == 0
		}

		return true
	},
}

func (ngx *NginxManager) loadTemplate() {
	tmpl, _ := template.New("nginx.tmpl").Funcs(funcMap).ParseFiles("./nginx.tmpl")
	ngx.template = tmpl
}

func (ngx *NginxManager) writeCfg(cfg *nginxConfiguration, ingressCfg IngressConfig) (bool, error) {
	fromMap := structs.Map(cfg)
	toMap := structs.Map(ngx.defCfg)
	curNginxCfg := merge(toMap, fromMap)

	conf := make(map[string]interface{})
	conf["upstreams"] = ingressCfg.Upstreams
	conf["servers"] = ingressCfg.Servers
	conf["tcpUpstreams"] = ingressCfg.TCPUpstreams
	conf["defResolver"] = ngx.defResolver
	conf["sslDHParam"] = ngx.sslDHParam
	conf["cfg"] = curNginxCfg

	buffer := new(bytes.Buffer)
	err := ngx.template.Execute(buffer, conf)
	if err != nil {
		glog.Infof("NGINX error: %v", err)
		return false, err
	}

	if glog.V(3) {
		b, err := json.Marshal(conf)
		if err != nil {
			fmt.Println("error:", err)
		}
		glog.Infof("NGINX configuration: %v", string(b))
	}

	changed, err := ngx.needsReload(buffer)
	if err != nil {
		return false, err
	}

	return changed, nil
}
