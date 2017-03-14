package main

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"

	nginxconfig "k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/controller"
	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/kubernetes/pkg/api"
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
