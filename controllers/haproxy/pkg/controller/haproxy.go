/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/ingress/controllers/haproxy/pkg/version"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/controller"
	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/kubernetes/pkg/api"
	"net/http"
	"os"
	"os/exec"
)

type haproxyController struct {
	controller *controller.GenericController
	configMap  *api.ConfigMap
	command    string
	configFile string
	template   *template
}

func newHAProxyController() *haproxyController {
	return &haproxyController{
		command:    "/haproxy-wrapper",
		configFile: "/usr/local/etc/haproxy/haproxy.cfg",
		template:   newTemplate("haproxy.tmpl", "/usr/local/etc/haproxy/haproxy.tmpl"),
	}
}

func (haproxy *haproxyController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "HAProxy",
		Release:    version.RELEASE,
		Build:      version.COMMIT,
		Repository: version.REPO,
	}
}

func (haproxy *haproxyController) Start() {
	controller := controller.NewIngressController(haproxy)
	haproxy.controller = controller
	haproxy.controller.Start()
}

func (haproxy *haproxyController) Stop() error {
	err := haproxy.controller.Stop()
	return err
}

func (haproxy *haproxyController) Name() string {
	return "HAProxy Ingress Controller"
}

func (haproxy *haproxyController) Check(_ *http.Request) error {
	return nil
}

func (haproxy *haproxyController) SetConfig(configMap *api.ConfigMap) {
	haproxy.configMap = configMap
}

func (haproxy *haproxyController) BackendDefaults() defaults.Backend {
	return defaults.Backend{}
}

func (haproxy *haproxyController) OnUpdate(cfg ingress.Configuration) ([]byte, error) {
	conf := newConfig(&cfg, haproxy.configMap.Data)
	data, err := haproxy.template.execute(conf)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (haproxy *haproxyController) Reload(data []byte) ([]byte, bool, error) {
	if !haproxy.configChanged(data) {
		return nil, false, nil
	}
	// TODO missing HAProxy validation before overwrite and try to reload
	err := ioutil.WriteFile(haproxy.configFile, data, 0644)
	if err != nil {
		return nil, false, err
	}
	out, err := haproxy.reloadHaproxy()
	if len(out) > 0 {
		glog.Infof("HAProxy output:\n%v", string(out))
	}
	return out, true, err
}

func (haproxy *haproxyController) configChanged(data []byte) bool {
	if _, err := os.Stat(haproxy.configFile); os.IsNotExist(err) {
		return true
	}
	cfg, err := ioutil.ReadFile(haproxy.configFile)
	if err != nil {
		return false
	}
	return !bytes.Equal(cfg, data)
}

func (haproxy *haproxyController) reloadHaproxy() ([]byte, error) {
	out, err := exec.Command(haproxy.command, haproxy.configFile).CombinedOutput()
	return out, err
}
