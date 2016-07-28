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
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/golang/glog"

	"github.com/fatih/structs"
	"github.com/ghodss/yaml"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/flowcontrol"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
)

// Manager ...
type Manager struct {
	ConfigFile string

	defCfg config.Configuration

	defResolver string

	sslDHParam string

	reloadRateLimiter flowcontrol.RateLimiter

	// template loaded ready to be used to generate the nginx configuration file
	template *template.Template

	reloadLock *sync.Mutex
}

// NewManager ...
func NewManager(kubeClient *client.Client) *Manager {
	ngx := &Manager{
		ConfigFile:        "/etc/nginx/nginx.conf",
		defCfg:            config.NewDefault(),
		defResolver:       strings.Join(getDNSServers(), " "),
		reloadLock:        &sync.Mutex{},
		reloadRateLimiter: flowcontrol.NewTokenBucketRateLimiter(0.1, 1),
	}

	ngx.createCertsDir(config.SSLDirectory)

	ngx.sslDHParam = ngx.SearchDHParamFile(config.SSLDirectory)

	if err := ngx.loadTemplate(); err != nil {
		glog.Fatalf("invalid NGINX template: %v", err)
	}

	return ngx
}

func (nginx *Manager) createCertsDir(base string) {
	if err := os.Mkdir(base, os.ModeDir); err != nil {
		if os.IsExist(err) {
			glog.Infof("%v already exists", err)
			return
		}
		glog.Fatalf("Couldn't create directory %v: %v", base, err)
	}
}

// ConfigMapAsString returns a ConfigMap with the default NGINX
// configuration to be used a guide to provide a custom configuration
func ConfigMapAsString() string {
	cfg := &api.ConfigMap{}
	cfg.Name = "custom-name"
	cfg.Namespace = "a-valid-namespace"
	cfg.Data = make(map[string]string)

	data := structs.Map(config.NewDefault())
	for k, v := range data {
		cfg.Data[k] = fmt.Sprintf("%v", v)
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		glog.Warningf("Unexpected error creating default configuration: %v", err)
		return ""
	}

	return string(out)
}
