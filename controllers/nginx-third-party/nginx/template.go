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
	"errors"
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
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
}

func (ngx *NginxManager) loadTemplate() {
	tmpl, _ := template.New("nginx.tmpl").Funcs(funcMap).ParseFiles("./nginx.tmpl")
	ngx.template = tmpl
}

func (ngx *NginxManager) writeCfg(cfg *nginxConfiguration, upstreams []Upstream, servers []Server, servicesL4 []Service) (bool, error) {
	fromMap := structs.Map(cfg)
	toMap := structs.Map(ngx.defCfg)
	curNginxCfg := merge(toMap, fromMap)

	conf := make(map[string]interface{})
	conf["upstreams"] = upstreams
	conf["servers"] = servers
	conf["tcpServices"] = servicesL4
	conf["defBackend"] = ngx.defBackend
	conf["defResolver"] = ngx.defResolver
	conf["sslDHParam"] = ngx.sslDHParam
	conf["cfg"] = curNginxCfg

	if ngx.defError.ServiceName != "" {
		conf["defErrorSvc"] = ngx.defError
	} else {
		conf["defErrorSvc"] = false
	}

	buffer := new(bytes.Buffer)
	err := ngx.template.Execute(buffer, conf)
	if err != nil {
		return false, err
	}

	changed, err := checkChanges(ngx.ConfigFile, buffer)
	if err != nil {
		return false, err
	}

	if glog.V(3) {
		b, err := json.Marshal(conf)
		if err != nil {
			fmt.Println("error:", err)
		}
		glog.Infof("nginx configuration: %v", string(b))
	}

	return changed, nil
}
