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

package nginx

import (
	"os"
	"strings"
	"sync"

	"github.com/golang/glog"

	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/flowcontrol"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/ingress"
	ngx_template "k8s.io/contrib/ingress/controllers/nginx/nginx/template"
)

var (
	tmplPath = "/etc/nginx/template/nginx.tmpl"
)

// Manager ...
type Manager struct {
	ConfigFile string

	defCfg config.Configuration

	defResolver string

	sslDHParam string

	reloadRateLimiter flowcontrol.RateLimiter

	// template loaded ready to be used to generate the nginx configuration file
	template *ngx_template.Template

	reloadLock *sync.Mutex
}

// NewDefaultServer return an UpstreamServer to be use as default server that returns 503.
func NewDefaultServer() ingress.UpstreamServer {
	return ingress.UpstreamServer{Address: "127.0.0.1", Port: "8181"}
}

// NewUpstream creates an upstream without servers.
func NewUpstream(name string) *ingress.Upstream {
	return &ingress.Upstream{
		Name:     name,
		Backends: []ingress.UpstreamServer{},
	}
}

// NewManager ...
func NewManager(kubeClient *client.Client) *Manager {
	ngx := &Manager{
		ConfigFile:        "/etc/nginx/nginx.conf",
		defCfg:            config.NewDefault(),
		reloadLock:        &sync.Mutex{},
		reloadRateLimiter: flowcontrol.NewTokenBucketRateLimiter(0.1, 1),
	}

	res, err := getDNSServers()
	if err != nil {
		glog.Warningf("error reading nameservers: %v", err)
	}
	ngx.defResolver = strings.Join(res, " ")

	ngx.createCertsDir(config.SSLDirectory)

	ngx.sslDHParam = ngx.SearchDHParamFile(config.SSLDirectory)

	var onChange func()
	onChange = func() {
		template, err := ngx_template.NewTemplate(tmplPath, onChange)
		if err != nil {
			glog.Warningf("invalid NGINX template: %v", err)
			return
		}

		ngx.template.Close()
		ngx.template = template
		glog.Info("new NGINX template loaded")
	}

	template, err := ngx_template.NewTemplate(tmplPath, onChange)
	if err != nil {
		glog.Fatalf("invalid NGINX template: %v", err)
	}

	ngx.template = template
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
