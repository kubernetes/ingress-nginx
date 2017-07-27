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
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	api_v1 "k8s.io/client-go/pkg/api/v1"

	"k8s.io/ingress/controllers/nginx/pkg/config"
	ngx_template "k8s.io/ingress/controllers/nginx/pkg/template"
	"k8s.io/ingress/controllers/nginx/pkg/version"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/defaults"
	//"k8s.io/ingress/core/pkg/net/dns"
	"k8s.io/ingress/core/pkg/net/ssl"
)

var (
	tmplPath        = "/etc/nginx/template/nginx.tmpl"
	cfgPath         = "/etc/nginx/nginx.conf"
	binary          = "/usr/sbin/nginx"
	defIngressClass = "nginx"
)

// newExtExtNGINXController creates a new NGINX Ingress controller.
// If the environment variable NGINX_BINARY exists it will be used
// as source for nginx commands
func newExtNGINXController() ingress.Controller {
	ngx := os.Getenv("NGINX_BINARY")
	if ngx == "" {
		ngx = binary
	}

	/*	h, err := dns.GetSystemNameServers()
		if err != nil {
			glog.Warningf("unexpected error reading system nameservers: %v", err)
		}
	*/
	n := &ExtNGINXController{
		binary:        ngx,
		configmap:     &api_v1.ConfigMap{},
		isIPV6Enabled: isIPv6Enabled(),
		//		resolver:      h,
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

	return ingress.Controller(n)
}

type server struct {
	Hostname string
	IP       string
	Port     int
}

// ExtNGINXController ...
type ExtNGINXController struct {
	t *ngx_template.Template

	configmap *api_v1.ConfigMap

	storeLister ingress.StoreLister

	binary   string
	resolver []net.IP

	cmdArgs []string

	// returns true if IPV6 is enabled in the pod
	isIPV6Enabled bool
}

// Start a dummy function since we don't manage nginx start and stop in a external environment.
func (n *ExtNGINXController) Start() {
	glog.Info("starting External NGINX Controller")

}

// BackendDefaults returns the nginx defaults
func (n ExtNGINXController) BackendDefaults() defaults.Backend {
	if n.configmap == nil {
		d := config.NewDefault()
		return d.Backend
	}

	return ngx_template.ReadConfig(n.configmap.Data).Backend
}

// printDiff returns the difference between the running configuration
// and the new one
func (n ExtNGINXController) printDiff(data []byte) {
	in, err := os.Open(cfgPath)
	if err != nil {
		return
	}
	src, err := ioutil.ReadAll(in)
	in.Close()
	if err != nil {
		return
	}

	if !bytes.Equal(src, data) {
		tmpfile, err := ioutil.TempFile("", "nginx-cfg-diff")
		if err != nil {
			glog.Errorf("error creating temporal file: %s", err)
			return
		}
		defer tmpfile.Close()
		err = ioutil.WriteFile(tmpfile.Name(), data, 0644)
		if err != nil {
			return
		}

		diffOutput, err := diff(src, data)
		if err != nil {
			glog.Errorf("error computing diff: %s", err)
			return
		}

		if glog.V(2) {
			glog.Infof("NGINX configuration diff\n")
			glog.Infof("%v", string(diffOutput))
		}
		os.Remove(tmpfile.Name())
	}
}

// Info return build information
func (n ExtNGINXController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "ExtNGINX",
		Release:    version.RELEASE,
		Build:      version.COMMIT,
		Repository: version.REPO,
	}
}

// ConfigureFlags allow to configure more flags before the parsing of
// command line arguments
func (n *ExtNGINXController) ConfigureFlags(flags *pflag.FlagSet) {
}

// OverrideFlags customize NGINX controller flags
func (n *ExtNGINXController) OverrideFlags(flags *pflag.FlagSet) {
}

// DefaultIngressClass just return the default ingress class
func (n ExtNGINXController) DefaultIngressClass() string {
	return defIngressClass
}

// testTemplate checks if the NGINX configuration inside the byte array is valid
// running the command "nginx -t" using a temporal file.
func (n ExtNGINXController) testTemplate(cfg []byte) error {
	if len(cfg) == 0 {
		return fmt.Errorf("invalid nginx configuration (empty)")
	}
	tmpfile, err := ioutil.TempFile("", "nginx-cfg")
	if err != nil {
		return err
	}
	defer tmpfile.Close()
	err = ioutil.WriteFile(tmpfile.Name(), cfg, 0644)
	if err != nil {
		return err
	}
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

// SetConfig sets the configured configmap
func (n *ExtNGINXController) SetConfig(cmap *api_v1.ConfigMap) {
	n.configmap = cmap

	if cmap == nil {
		return
	}
}

// SetListers sets the configured store listers in the generic ingress controller
func (n *ExtNGINXController) SetListers(lister ingress.StoreLister) {
	n.storeLister = lister
}

// OnUpdate is called by syncQueue in https://github.com/aledbf/ingress-controller/blob/master/pkg/ingress/controller/controller.go#L82
// periodically to keep the configuration in sync.
//
// convert configmap to custom configuration object (different in each implementation)
// write the custom template (the complexity depends on the implementation)
// write the configuration file
// returning nill implies the backend will be reloaded.
// if an error is returned means requeue the update
func (n *ExtNGINXController) OnUpdate(ingressCfg ingress.Configuration) error {
	var longestName int
	var serverNameBytes int
	for _, srv := range ingressCfg.Servers {
		if longestName < len(srv.Hostname) {
			longestName = len(srv.Hostname)
		}
		serverNameBytes += len(srv.Hostname)
	}

	cfg := ngx_template.ReadConfig(n.configmap.Data)
	cfg.Resolver = n.resolver

	servers := []*server{}
	for _, pb := range ingressCfg.PassthroughBackends {
		svc := pb.Service
		if svc == nil {
			glog.Warningf("missing service for PassthroughBackends %v", pb.Backend)
			continue
		}
		port, err := strconv.Atoi(pb.Port.String())
		if err != nil {
			for _, sp := range svc.Spec.Ports {
				if sp.Name == pb.Port.String() {
					port = int(sp.Port)
					break
				}
			}
		} else {
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(port) {
					port = int(sp.Port)
					break
				}
			}
		}

		//TODO: Allow PassthroughBackends to specify they support proxy-protocol
		servers = append(servers, &server{
			Hostname: pb.Hostname,
			IP:       svc.Spec.ClusterIP,
			Port:     port,
		})
	}

	// NGINX cannot resize the has tables used to store server names.
	// For this reason we check if the defined size defined is correct
	// for the FQDN defined in the ingress rules adjusting the value
	// if is required.
	// https://trac.nginx.org/nginx/ticket/352
	// https://trac.nginx.org/nginx/ticket/631
	nameHashBucketSize := nginxHashBucketSize(longestName)
	if cfg.ServerNameHashBucketSize == 0 {
		glog.V(3).Infof("adjusting ServerNameHashBucketSize variable to %v", nameHashBucketSize)
		cfg.ServerNameHashBucketSize = nameHashBucketSize
	}
	serverNameHashMaxSize := nextPowerOf2(serverNameBytes)
	if cfg.ServerNameHashMaxSize < serverNameHashMaxSize {
		glog.V(3).Infof("adjusting ServerNameHashMaxSize variable to %v", serverNameHashMaxSize)
		cfg.ServerNameHashMaxSize = serverNameHashMaxSize
	}

	// the limit of open files is per worker process
	// and we leave some room to avoid consuming all the FDs available
	wp, err := strconv.Atoi(cfg.WorkerProcesses)
	glog.V(3).Infof("number of worker processes: %v", wp)
	if err != nil {
		wp = 1
	}
	maxOpenFiles := (sysctlFSFileMax() / wp) - 1024
	glog.V(3).Infof("maximum number of open file descriptors : %v", sysctlFSFileMax())
	if maxOpenFiles < 1024 {
		// this means the value of RLIMIT_NOFILE is too low.
		maxOpenFiles = 1024
	}

	setHeaders := map[string]string{}
	if cfg.ProxySetHeaders != "" {
		cmap, exists, err := n.storeLister.ConfigMap.GetByKey(cfg.ProxySetHeaders)
		if err != nil {
			glog.Warningf("unexpected error reading configmap %v: %v", cfg.ProxySetHeaders, err)
		}

		if exists {
			setHeaders = cmap.(*api_v1.ConfigMap).Data
		}
	}

	addHeaders := map[string]string{}
	if cfg.AddHeaders != "" {
		cmap, exists, err := n.storeLister.ConfigMap.GetByKey(cfg.AddHeaders)
		if err != nil {
			glog.Warningf("unexpected error reading configmap %v: %v", cfg.AddHeaders, err)
		}

		if exists {
			addHeaders = cmap.(*api_v1.ConfigMap).Data
		}
	}

	sslDHParam := ""
	if cfg.SSLDHParam != "" {
		secretName := cfg.SSLDHParam
		s, exists, err := n.storeLister.Secret.GetByKey(secretName)
		if err != nil {
			glog.Warningf("unexpected error reading secret %v: %v", secretName, err)
		}

		if exists {
			secret := s.(*api_v1.Secret)
			nsSecName := strings.Replace(secretName, "/", "-", -1)

			dh, ok := secret.Data["dhparam.pem"]
			if ok {
				pemFileName, err := ssl.AddOrUpdateDHParam(nsSecName, dh)
				if err != nil {
					glog.Warningf("unexpected error adding or updating dhparam %v file: %v", nsSecName, err)
				} else {
					sslDHParam = pemFileName
				}
			}
		}
	}

	cfg.SSLDHParam = sslDHParam

	content, err := n.t.Write(config.TemplateConfig{
		ProxySetHeaders:     setHeaders,
		AddHeaders:          addHeaders,
		MaxOpenFiles:        maxOpenFiles,
		BacklogSize:         sysctlSomaxconn(),
		Backends:            ingressCfg.Backends,
		PassthroughBackends: ingressCfg.PassthroughBackends,
		Servers:             ingressCfg.Servers,
		TCPBackends:         ingressCfg.TCPEndpoints,
		UDPBackends:         ingressCfg.UDPEndpoints,
		CustomErrors:        len(cfg.CustomHTTPErrors) > 0,
		Cfg:                 cfg,
		IsIPV6Enabled:       n.isIPV6Enabled && !cfg.DisableIpv6,
	})

	if err != nil {
		return err
	}

	err = n.testTemplate(content)
	if err != nil {
		return err
	}

	n.printDiff(content)

	err = ioutil.WriteFile(cfgPath, content, 0644)
	if err != nil {
		return err
	}

	o, err := exec.Command(n.binary, "-s", "reload", "-c", cfgPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%v", err, string(o))
	}

	return nil
}

// nginxHashBucketSize computes the correct nginx hash_bucket_size for a hash with the given longest key
func nginxHashBucketSize(longestString int) int {
	// See https://github.com/kubernetes/ingress/issues/623 for an explanation
	wordSize := 8 // Assume 64 bit CPU
	n := longestString + 2
	aligned := (n + wordSize - 1) & ^(wordSize - 1)
	rawSize := wordSize + wordSize + aligned
	return nextPowerOf2(rawSize)
}

// Name returns the healthcheck name
func (n ExtNGINXController) Name() string {
	return "Ingress Controller"
}

// Check returns nil because we are not running in a pod.
func (n ExtNGINXController) Check(_ *http.Request) error {
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

func isIPv6Enabled() bool {
	cmd := exec.Command("test", "-f", "/proc/net/if_inet6")
	return cmd.Run() == nil
}
