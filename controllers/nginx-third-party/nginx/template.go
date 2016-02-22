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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/fatih/structs"
	"github.com/golang/glog"

	"k8s.io/contrib/ingress/controllers/nginx-third-party/ssl"
)

var funcMap = template.FuncMap{
	"getSSLHost": ssl.GetSSLHost,
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

func (ngx *NginxManager) writeCfg(cfg *nginxConfiguration, servicesL4 []Service) error {
	file, err := os.Create(ngx.ConfigFile)
	if err != nil {
		return err
	}

	fromMap := structs.Map(cfg)
	toMap := structs.Map(ngx.defCfg)
	curNginxCfg := merge(toMap, fromMap)

	conf := make(map[string]interface{})
	conf["sslCertificates"] = ngx.sslCertificates
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

	if glog.V(2) {
		b, err := json.Marshal(conf)
		if err != nil {
			fmt.Println("error:", err)
		}
		glog.Infof("nginx configuration: %v", string(b))
	}

	err = ngx.template.Execute(file, conf)
	if err != nil {
		return err
	}

	return nil
}
