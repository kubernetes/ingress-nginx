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

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"

	"k8s.io/kubernetes/pkg/api"

	"k8s.io/ingress/controllers/nginx/pkg/config"
	ngx_template "k8s.io/ingress/controllers/nginx/pkg/template"
	"k8s.io/ingress/controllers/nginx/pkg/version"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

const (
	ngxHealthPort = 18080
	ngxHealthPath = "/healthz"
	ngxStatusPath = "/internal_nginx_status"
)

var (
	tmplPath = "/etc/nginx/template/nginx.tmpl"
	cfgPath  = "/etc/nginx/nginx.conf"
	binary   = "/usr/sbin/nginx"
)

// newNGINXController creates a new NGINX Ingress controller.
// If the environment variable NGINX_BINARY exists it will be used
// as source for nginx commands
func newNGINXController() ingress.Controller {
	ngx := os.Getenv("NGINX_BINARY")
	if ngx == "" {
		ngx = binary
	}
	n := NGINXController{
		binary:    ngx,
		configmap: &api.ConfigMap{},
	}

	var onChange func()
	onChange = func() {
		template, err := ngx_template.NewTemplate(tmplPath, onChange)
		if err != nil {
			// this error is different from the rest because it must be clear why nginx is not working
			glog.Errorf(`
-------------------------------------------------------------------------------
Error loading new template : %v
-------------------------------------------------------------------------------
`, err)
			return
		}

		n.t.Close()
		n.t = template
		glog.Info("new NGINX template loaded")
	}

	ngxTpl, err := ngx_template.NewTemplate(tmplPath, onChange)
	if err != nil {
		glog.Fatalf("invalid NGINX template: %v", err)
	}

	n.t = ngxTpl

	go n.Start()

	return ingress.Controller(&n)
}

// NGINXController ...
type NGINXController struct {
	t *ngx_template.Template

	configmap *api.ConfigMap

	binary string
}

// Start start a new NGINX master process running in foreground.
func (n NGINXController) Start() {
	glog.Info("starting NGINX process...")

	done := make(chan error, 1)
	cmd := exec.Command(n.binary, "-c", cfgPath)
	n.start(cmd, done)

	// if the nginx master process dies the workers continue to process requests,
	// passing checks but in case of updates in ingress no updates will be
	// reflected in the nginx configuration which can lead to confusion and report
	// issues because of this behavior.
	// To avoid this issue we restart nginx in case of errors.
	for {
		err := <-done
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			glog.Warningf(`
-------------------------------------------------------------------------------
NGINX master process died (%v): %v
-------------------------------------------------------------------------------
`, waitStatus.ExitStatus(), err)
		}
		cmd.Process.Release()
		cmd = exec.Command(n.binary, "-c", cfgPath)
		// we wait until the workers are killed
		for {
			conn, err := net.DialTimeout("tcp", "127.0.0.1:80", 1*time.Second)
			if err == nil {
				conn.Close()
				break
			}
			time.Sleep(1 * time.Second)
		}
		// start a new nginx master process
		n.start(cmd, done)
	}
}

func (n *NGINXController) start(cmd *exec.Cmd, done chan error) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		glog.Fatalf("nginx error: %v", err)
		done <- err
		return
	}

	n.setupMonitor(cmd.Args)

	go func() {
		done <- cmd.Wait()
	}()
}

// Reload checks if the running configuration file is different
// to the specified and reload nginx if required
func (n NGINXController) Reload(data []byte) ([]byte, bool, error) {
	if !n.isReloadRequired(data) {
		return []byte("Reload not required"), false, nil
	}

	err := ioutil.WriteFile(cfgPath, data, 0644)
	if err != nil {
		return nil, false, err
	}

	o, e := exec.Command(n.binary, "-s", "reload").CombinedOutput()
	return o, true, e
}

// BackendDefaults returns the nginx defaults
func (n NGINXController) BackendDefaults() defaults.Backend {
	if n.configmap == nil {
		d := config.NewDefault()
		return d.Backend
	}

	return n.backendDefaults()
}

func (n *NGINXController) backendDefaults() defaults.Backend {
	d := config.NewDefault()
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           &d,
		TagName:          "json",
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		glog.Warningf("unexpected error merging defaults: %v", err)
	}
	decoder.Decode(n.configmap.Data)
	return d.Backend
}

// isReloadRequired check if the new configuration file is different
// from the current one.
func (n NGINXController) isReloadRequired(data []byte) bool {
	in, err := os.Open(cfgPath)
	if err != nil {
		return false
	}
	src, err := ioutil.ReadAll(in)
	in.Close()
	if err != nil {
		return false
	}

	if !bytes.Equal(src, data) {
		tmpfile, err := ioutil.TempFile("", "nginx-cfg-diff")
		if err != nil {
			glog.Errorf("error creating temporal file: %s", err)
			return false
		}
		defer tmpfile.Close()
		err = ioutil.WriteFile(tmpfile.Name(), data, 0644)
		if err != nil {
			return false
		}

		diffOutput, err := diff(src, data)
		if err != nil {
			glog.Errorf("error computing diff: %s", err)
			return true
		}

		if glog.V(2) {
			glog.Infof("NGINX configuration diff\n")
			glog.Infof("%v", string(diffOutput))
		}
		return len(diffOutput) > 0
	}
	return false
}

// Info return build information
func (n NGINXController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "NGINX",
		Release:    version.RELEASE,
		Build:      version.COMMIT,
		Repository: version.REPO,
	}
}

// testTemplate checks if the NGINX configuration inside the byte array is valid
// running the command "nginx -t" using a temporal file.
func (n NGINXController) testTemplate(cfg []byte) error {
	tmpfile, err := ioutil.TempFile("", "nginx-cfg")
	if err != nil {
		return err
	}
	defer tmpfile.Close()
	ioutil.WriteFile(tmpfile.Name(), cfg, 0644)
	out, err := exec.Command(n.binary, "-t", "-c", tmpfile.Name()).CombinedOutput()
	if err != nil {
		// this error is different from the rest because it must be clear why nginx is not working
		oe := fmt.Sprintf(`
-------------------------------------------------------------------------------
Error: %v
%v
-------------------------------------------------------------------------------
`, err, string(out))
		return errors.New(oe)
	}

	os.Remove(tmpfile.Name())
	return nil
}

// SetConfig ...
func (n *NGINXController) SetConfig(cmap *api.ConfigMap) {
	n.configmap = cmap
}

// OnUpdate is called by syncQueue in https://github.com/aledbf/ingress-controller/blob/master/pkg/ingress/controller/controller.go#L82
// periodically to keep the configuration in sync.
//
// convert configmap to custom configuration object (different in each implementation)
// write the custom template (the complexity depends on the implementation)
// write the configuration file
// returning nill implies the backend will be reloaded.
// if an error is returned means requeue the update
func (n *NGINXController) OnUpdate(ingressCfg ingress.Configuration) ([]byte, error) {
	var longestName int
	var serverNames int
	for _, srv := range ingressCfg.Servers {
		serverNames += len([]byte(srv.Hostname))
		if longestName < len(srv.Hostname) {
			longestName = len(srv.Hostname)
		}
	}

	cfg := ngx_template.ReadConfig(n.configmap.Data)

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

	// the limit of open files is per worker process
	// and we leave some room to avoid consuming all the FDs available
	maxOpenFiles := (sysctlFSFileMax() / cfg.WorkerProcesses) - 1024

	return n.t.Write(config.TemplateConfig{
		MaxOpenFiles:        maxOpenFiles,
		BacklogSize:         sysctlSomaxconn(),
		Backends:            ingressCfg.Backends,
		PassthroughBackends: ingressCfg.PassthroughBackends,
		Servers:             ingressCfg.Servers,
		TCPBackends:         ingressCfg.TCPEndpoints,
		UDPBackends:         ingressCfg.UPDEndpoints,
		HealthzURI:          ngxHealthPath,
		CustomErrors:        len(cfg.CustomHTTPErrors) > 0,
		Cfg:                 cfg,
	}, n.testTemplate)
}

// Name returns the healthcheck name
func (n NGINXController) Name() string {
	return "Ingress Controller"
}

// Check returns if the nginx healthz endpoint is returning ok (status code 200)
func (n NGINXController) Check(_ *http.Request) error {
	res, err := http.Get(fmt.Sprintf("http://localhost:%v%v", ngxHealthPort, ngxHealthPath))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Ingress controller is not healthy")
	}
	return nil
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
