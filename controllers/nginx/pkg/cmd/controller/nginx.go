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
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/mitchellh/go-ps"
	"github.com/spf13/pflag"

	proxyproto "github.com/armon/go-proxyproto"
	"github.com/ncabatoff/process-exporter/proc"
	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/controllers/nginx/pkg/config"
	ngx_template "k8s.io/ingress/controllers/nginx/pkg/template"
	"k8s.io/ingress/controllers/nginx/pkg/version"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/controller"
	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/ingress/core/pkg/net/dns"
	"k8s.io/ingress/core/pkg/net/ssl"
)

type statusModule string

const (
	ngxHealthPath = "/healthz"

	defaultStatusModule statusModule = "default"
	vtsStatusModule     statusModule = "vts"

	defUpstreamName = "upstream-default-backend"
)

var (
	tmplPath        = "/etc/nginx/template/nginx.tmpl"
	cfgPath         = "/etc/nginx/nginx.conf"
	nginxBinary     = "/usr/sbin/nginx"
	defIngressClass = "nginx"
)

// newNGINXController creates a new NGINX Ingress controller.
// If the environment variable NGINX_BINARY exists it will be used
// as source for nginx commands
func newNGINXController() *NGINXController {
	ngx := os.Getenv("NGINX_BINARY")
	if ngx == "" {
		ngx = nginxBinary
	}

	h, err := dns.GetSystemNameServers()
	if err != nil {
		glog.Warningf("unexpected error reading system nameservers: %v", err)
	}

	n := &NGINXController{
		binary:          ngx,
		configmap:       &apiv1.ConfigMap{},
		isIPV6Enabled:   isIPv6Enabled(),
		resolver:        h,
		ports:           &config.ListenPorts{},
		backendDefaults: config.NewDefault().Backend,
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

	return n
}

// NGINXController ...
type NGINXController struct {
	controller *controller.GenericController
	t          *ngx_template.Template

	configmap *apiv1.ConfigMap

	storeLister *ingress.StoreLister

	binary   string
	resolver []net.IP

	cmdArgs []string

	stats        *statsCollector
	statusModule statusModule

	// returns true if IPV6 is enabled in the pod
	isIPV6Enabled bool

	// returns true if proxy protocol es enabled
	isProxyProtocolEnabled bool

	isSSLPassthroughEnabled bool

	isShuttingDown bool

	proxy *proxy

	ports *config.ListenPorts

	backendDefaults defaults.Backend
}

// Start start a new NGINX master process running in foreground.
func (n *NGINXController) Start() {
	n.isShuttingDown = false

	n.controller = controller.NewIngressController(n)
	go n.controller.Start()

	done := make(chan error, 1)
	cmd := exec.Command(n.binary, "-c", cfgPath)

	// put nginx in another process group to prevent it
	// to receive signals meant for the controller
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	glog.Info("starting NGINX process...")
	n.start(cmd, done)

	// if the nginx master process dies the workers continue to process requests,
	// passing checks but in case of updates in ingress no updates will be
	// reflected in the nginx configuration which can lead to confusion and report
	// issues because of this behavior.
	// To avoid this issue we restart nginx in case of errors.
	for {
		err := <-done

		if n.isShuttingDown {
			break
		}

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
			if err != nil {
				break
			}
			conn.Close()
			// kill nginx worker processes
			fs, err := proc.NewFS("/proc")
			procs, _ := fs.FS.AllProcs()
			for _, p := range procs {
				pn, err := p.Comm()
				if err != nil {
					glog.Errorf("unexpected error obtaining process information: %v", err)
					continue
				}

				if pn == "nginx" {
					osp, err := os.FindProcess(p.PID)
					if err != nil {
						glog.Errorf("unexpected error obtaining process information: %v", err)
						continue
					}
					osp.Signal(syscall.SIGQUIT)
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		// restart a new nginx master process if the controller
		// is not being stopped
		n.start(cmd, done)
	}
}

// Stop gracefully stops the NGINX master process.
func (n *NGINXController) Stop() error {
	n.isShuttingDown = true
	n.controller.Stop()

	// Send stop signal to Nginx
	glog.Info("stopping NGINX process...")
	cmd := exec.Command(n.binary, "-c", cfgPath, "-s", "quit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	// Wait for the Nginx process disappear
	waitForNginxShutdown()
	glog.Info("NGINX process has stopped")

	return nil
}

func (n *NGINXController) start(cmd *exec.Cmd, done chan error) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		glog.Fatalf("nginx error: %v", err)
		done <- err
		return
	}

	n.cmdArgs = cmd.Args

	go func() {
		done <- cmd.Wait()
	}()
}

// BackendDefaults returns the nginx defaults
func (n NGINXController) BackendDefaults() defaults.Backend {
	return n.backendDefaults
}

// printDiff returns the difference between the running configuration
// and the new one
func (n NGINXController) printDiff(data []byte) {
	if !glog.V(2) {
		return
	}

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

		glog.Infof("NGINX configuration diff\n")
		glog.Infof("%v", string(diffOutput))

		os.Remove(tmpfile.Name())
	}
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

// DefaultEndpoint returns the default endpoint to be use as default server that returns 404.
func (n NGINXController) DefaultEndpoint() ingress.Endpoint {
	return ingress.Endpoint{
		Address: "127.0.0.1",
		Port:    fmt.Sprintf("%v", n.ports.Default),
		Target:  &apiv1.ObjectReference{},
	}
}

// ConfigureFlags allow to configure more flags before the parsing of
// command line arguments
func (n *NGINXController) ConfigureFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&n.isSSLPassthroughEnabled, "enable-ssl-passthrough", false, `Enable SSL passthrough feature. Default is disabled`)
	flags.IntVar(&n.ports.HTTP, "http-port", 80, `Indicates the port to use for HTTP traffic`)
	flags.IntVar(&n.ports.HTTPS, "https-port", 443, `Indicates the port to use for HTTPS traffic`)
	flags.IntVar(&n.ports.Status, "status-port", 18080, `Indicates the TCP port to use for exposing the nginx status page`)
	flags.IntVar(&n.ports.SSLProxy, "ssl-passtrough-proxy-port", 442, `Default port to use internally for SSL when SSL Passthgough is enabled`)
	flags.IntVar(&n.ports.Default, "default-server-port", 8181, `Default port to use for exposing the default server (catch all)`)
}

// OverrideFlags customize NGINX controller flags
func (n *NGINXController) OverrideFlags(flags *pflag.FlagSet) {
	// we check port collisions
	if !isPortAvailable(n.ports.HTTP) {
		glog.Fatalf("Port %v is already in use. Please check the flag --http-port", n.ports.HTTP)
	}
	if !isPortAvailable(n.ports.HTTPS) {
		glog.Fatalf("Port %v is already in use. Please check the flag --https-port", n.ports.HTTPS)
	}
	if !isPortAvailable(n.ports.Status) {
		glog.Fatalf("Port %v is already in use. Please check the flag --status-port", n.ports.Status)
	}
	if !isPortAvailable(n.ports.Default) {
		glog.Fatalf("Port %v is already in use. Please check the flag --default-server-port", n.ports.Default)
	}

	ic, _ := flags.GetString("ingress-class")
	wc, _ := flags.GetString("watch-namespace")

	if ic == "" {
		ic = defIngressClass
	}

	if ic != defIngressClass {
		glog.Warningf("only Ingress with class %v will be processed by this ingress controller", ic)
	}

	flags.Set("ingress-class", ic)

	h, _ := flags.GetInt("healthz-port")
	n.ports.Health = h

	n.stats = newStatsCollector(wc, ic, n.binary, n.ports.Status)

	if n.isSSLPassthroughEnabled {
		if !isPortAvailable(n.ports.SSLProxy) {
			glog.Fatalf("Port %v is already in use. Please check the flag --ssl-passtrough-proxy-port", n.ports.SSLProxy)
		}

		glog.Info("starting TLS proxy for SSL passthrough")
		n.proxy = &proxy{
			Default: &server{
				Hostname:      "localhost",
				IP:            "127.0.0.1",
				Port:          n.ports.SSLProxy,
				ProxyProtocol: true,
			},
		}

		listener, err := net.Listen("tcp", fmt.Sprintf(":%v", n.ports.HTTPS))
		if err != nil {
			glog.Fatalf("%v", err)
		}

		proxyList := &proxyproto.Listener{Listener: listener}

		// start goroutine that accepts tcp connections in port 443
		go func() {
			for {
				var conn net.Conn
				var err error

				if n.isProxyProtocolEnabled {
					// we need to wrap the listener in order to decode
					// proxy protocol before handling the connection
					conn, err = proxyList.Accept()
				} else {
					conn, err = listener.Accept()
				}

				if err != nil {
					glog.Warningf("unexpected error accepting tcp connection: %v", err)
					continue
				}

				glog.V(3).Infof("remote address %s to local %s", conn.RemoteAddr(), conn.LocalAddr())
				go n.proxy.Handle(conn)
			}
		}()
	}
}

// DefaultIngressClass just return the default ingress class
func (n NGINXController) DefaultIngressClass() string {
	return defIngressClass
}

// testTemplate checks if the NGINX configuration inside the byte array is valid
// running the command "nginx -t" using a temporal file.
func (n NGINXController) testTemplate(cfg []byte) error {
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
func (n *NGINXController) SetConfig(cmap *apiv1.ConfigMap) {
	n.configmap = cmap
	n.isProxyProtocolEnabled = false

	m := map[string]string{}
	if cmap != nil {
		m = cmap.Data
	}

	val, ok := m["use-proxy-protocol"]
	if ok {
		b, err := strconv.ParseBool(val)
		if err == nil {
			n.isProxyProtocolEnabled = b
		}
	}

	n.backendDefaults = ngx_template.ReadConfig(m).Backend
}

// SetListers sets the configured store listers in the generic ingress controller
func (n *NGINXController) SetListers(lister *ingress.StoreLister) {
	n.storeLister = lister
}

// UpdateIngressStatus custom Ingress status update
func (n *NGINXController) UpdateIngressStatus(*extensions.Ingress) []apiv1.LoadBalancerIngress {
	return nil
}

// OnUpdate is called by syncQueue in https://github.com/kubernetes/ingress/blob/master/core/pkg/ingress/controller/controller.go#L426
// periodically to keep the configuration in sync.
//
// convert configmap to custom configuration object (different in each implementation)
// write the custom template (the complexity depends on the implementation)
// write the configuration file
// returning nill implies the backend will be reloaded.
// if an error is returned means requeue the update
func (n *NGINXController) OnUpdate(ingressCfg ingress.Configuration) error {
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
			Hostname:      pb.Hostname,
			IP:            svc.Spec.ClusterIP,
			Port:          port,
			ProxyProtocol: false,
		})
	}

	if n.isSSLPassthroughEnabled {
		n.proxy.ServerList = servers
	}

	// we need to check if the status module configuration changed
	if cfg.EnableVtsStatus {
		n.setupMonitor(vtsStatusModule)
	} else {
		n.setupMonitor(defaultStatusModule)
	}

	// NGINX cannot resize the hash tables used to store server names.
	// For this reason we check if the defined size defined is correct
	// for the FQDN defined in the ingress rules adjusting the value
	// if is required.
	// https://trac.nginx.org/nginx/ticket/352
	// https://trac.nginx.org/nginx/ticket/631
	var longestName int
	var serverNameBytes int
	redirectServers := make(map[string]string)
	for _, srv := range ingressCfg.Servers {
		if longestName < len(srv.Hostname) {
			longestName = len(srv.Hostname)
		}
		serverNameBytes += len(srv.Hostname)
		if srv.RedirectFromToWWW {
			var n string
			if strings.HasPrefix(srv.Hostname, "www.") {
				n = strings.TrimLeft(srv.Hostname, "www.")
			} else {
				n = fmt.Sprintf("www.%v", srv.Hostname)
			}
			glog.V(3).Infof("creating redirect from %v to %v", srv.Hostname, n)
			if _, ok := redirectServers[n]; !ok {
				found := false
				for _, esrv := range ingressCfg.Servers {
					if esrv.Hostname == n {
						found = true
						break
					}
				}
				if !found {
					redirectServers[n] = srv.Hostname
				}
			}
		}
	}
	if cfg.ServerNameHashBucketSize == 0 {
		nameHashBucketSize := nginxHashBucketSize(longestName)
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
			setHeaders = cmap.(*apiv1.ConfigMap).Data
		}
	}

	addHeaders := map[string]string{}
	if cfg.AddHeaders != "" {
		cmap, exists, err := n.storeLister.ConfigMap.GetByKey(cfg.AddHeaders)
		if err != nil {
			glog.Warningf("unexpected error reading configmap %v: %v", cfg.AddHeaders, err)
		}

		if exists {
			addHeaders = cmap.(*apiv1.ConfigMap).Data
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
			secret := s.(*apiv1.Secret)
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

	tc := config.TemplateConfig{
		ProxySetHeaders:         setHeaders,
		AddHeaders:              addHeaders,
		MaxOpenFiles:            maxOpenFiles,
		BacklogSize:             sysctlSomaxconn(),
		Backends:                ingressCfg.Backends,
		PassthroughBackends:     ingressCfg.PassthroughBackends,
		Servers:                 ingressCfg.Servers,
		TCPBackends:             ingressCfg.TCPEndpoints,
		UDPBackends:             ingressCfg.UDPEndpoints,
		HealthzURI:              ngxHealthPath,
		CustomErrors:            len(cfg.CustomHTTPErrors) > 0,
		Cfg:                     cfg,
		IsIPV6Enabled:           n.isIPV6Enabled && !cfg.DisableIpv6,
		RedirectServers:         redirectServers,
		IsSSLPassthroughEnabled: n.isSSLPassthroughEnabled,
		ListenPorts:             n.ports,
		PublishService:          n.controller.GetPublishService(),
	}

	content, err := n.t.Write(tc)

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
func (n NGINXController) Name() string {
	return "Ingress Controller"
}

// Check returns if the nginx healthz endpoint is returning ok (status code 200)
func (n NGINXController) Check(_ *http.Request) error {
	res, err := http.Get(fmt.Sprintf("http://localhost:%v%v", n.ports.Status, ngxHealthPath))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("ingress controller is not healthy")
	}

	// check the nginx master process is running
	fs, err := proc.NewFS("/proc")
	if err != nil {
		glog.Errorf("%v", err)
		return err
	}
	f, err := ioutil.ReadFile("/run/nginx.pid")
	if err != nil {
		glog.Errorf("%v", err)
		return err
	}
	pid, err := strconv.Atoi(strings.TrimRight(string(f), "\r\n"))
	if err != nil {
		return err
	}
	_, err = fs.NewProc(int(pid))
	if err != nil {
		glog.Errorf("%v", err)
		return err
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

func isIPv6Enabled() bool {
	cmd := exec.Command("test", "-f", "/proc/net/if_inet6")
	return cmd.Run() == nil
}

// isNginxRunning returns true if a process with the name 'nginx' is found
func isNginxProcessPresent() bool {
	processes, _ := ps.Processes()
	for _, p := range processes {
		if p.Executable() == "nginx" {
			return true
		}
	}
	return false
}

func waitForNginxShutdown() {
	timer := time.NewTicker(time.Second * 1)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			if !isNginxProcessPresent() {
				return
			}
		}
	}
}
