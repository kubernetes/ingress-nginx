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
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"

	api "k8s.io/client-go/pkg/api/v1"

	nginxconfig "k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/controller"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

func main() {
	dc := newDummyController()
	ic := controller.NewIngressController(dc)
	defer func() {
		log.Printf("Shutting down ingress controller...")
		ic.Stop()
	}()
	ic.Start()
}

func newDummyController() ingress.Controller {
	return &DummyController{}
}

type DummyController struct{}

func (dc DummyController) SetConfig(cfgMap *api.ConfigMap) {
	log.Printf("Config map %+v", cfgMap)
}

func (dc DummyController) Reload(data []byte) ([]byte, bool, error) {
	out, err := exec.Command("echo", string(data)).CombinedOutput()
	if err != nil {
		log.Printf("Reloaded new config %s", out)
	} else {
		return out, false, err
	}
	return out, true, err
}

func (dc DummyController) Test(file string) *exec.Cmd {
	return exec.Command("echo", file)
}

func (dc DummyController) OnUpdate(updatePayload ingress.Configuration) ([]byte, error) {
	log.Printf("Received OnUpdate notification")
	for _, b := range updatePayload.Backends {
		eps := []string{}
		for _, e := range b.Endpoints {
			eps = append(eps, e.Address)
		}
		log.Printf("%v: %v", b.Name, strings.Join(eps, ", "))
	}
	return []byte(`<string containing a configuration file>`), nil
}

func (dc DummyController) BackendDefaults() defaults.Backend {
	// Just adopt nginx's default backend config
	return nginxconfig.NewDefault().Backend
}

func (n DummyController) Name() string {
	return "dummy Controller"
}

func (n DummyController) Check(_ *http.Request) error {
	return nil
}

func (dc DummyController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "dummy",
		Release:    "0.0.0",
		Build:      "git-00000000",
		Repository: "git://foo.bar.com",
	}
}

func (n DummyController) OverrideFlags(*pflag.FlagSet) {
}

func (n DummyController) SetListers(lister ingress.StoreLister) {

}

func (n DummyController) DefaultIngressClass() string {
	return "dummy"
}
