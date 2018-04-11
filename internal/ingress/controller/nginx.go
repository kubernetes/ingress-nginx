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

package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"

	proxyproto "github.com/armon/go-proxyproto"
	"github.com/eapache/channels"
	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/kubernetes/pkg/util/filesystem"

	"path/filepath"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/process"
	"k8s.io/ingress-nginx/internal/ingress/controller/store"
	ngx_template "k8s.io/ingress-nginx/internal/ingress/controller/template"
	"k8s.io/ingress-nginx/internal/ingress/status"
	ing_net "k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/internal/net/dns"
	"k8s.io/ingress-nginx/internal/net/ssl"
	"k8s.io/ingress-nginx/internal/task"
	"k8s.io/ingress-nginx/internal/watch"
)

type statusModule string

const (
	ngxHealthPath = "/healthz"

	defaultStatusModule statusModule = "default"
	vtsStatusModule     statusModule = "vts"
)

var (
	tmplPath    = "/etc/nginx/template/nginx.tmpl"
	geoipPath   = "/etc/nginx/geoip"
	cfgPath     = "/etc/nginx/nginx.conf"
	nginxBinary = "/usr/sbin/nginx"
)

// NewNGINXController creates a new NGINX Ingress controller.
// If the environment variable NGINX_BINARY exists it will be used
// as source for nginx commands
func NewNGINXController(config *Configuration, fs file.Filesystem) *NGINXController {
	ngx := os.Getenv("NGINX_BINARY")
	if ngx == "" {
		ngx = nginxBinary
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{
		Interface: config.Client.CoreV1().Events(config.Namespace),
	})

	h, err := dns.GetSystemNameServers()
	if err != nil {
		glog.Warningf("unexpected error reading system nameservers: %v", err)
	}

	n := &NGINXController{
		binary: ngx,

		isIPV6Enabled: ing_net.IsIPv6Enabled(),

		resolver:        h,
		cfg:             config,
		syncRateLimiter: flowcontrol.NewTokenBucketRateLimiter(config.SyncRateLimit, 1),

		recorder: eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{
			Component: "nginx-ingress-controller",
		}),

		stopCh:   make(chan struct{}),
		updateCh: channels.NewRingChannel(1024),

		stopLock: &sync.Mutex{},

		fileSystem: fs,

		// create an empty configuration.
		runningConfig: &ingress.Configuration{},

		Proxy: &TCPProxy{},
	}

	n.store = store.New(
		config.EnableSSLChainCompletion,
		config.Namespace,
		config.ConfigMapName,
		config.TCPConfigMapName,
		config.UDPConfigMapName,
		config.DefaultSSLCertificate,
		config.ResyncPeriod,
		config.Client,
		fs,
		n.updateCh)

	n.stats = newStatsCollector(config.Namespace, class.IngressClass, n.binary, n.cfg.ListenPorts.Status)

	n.syncQueue = task.NewTaskQueue(n.syncIngress)

	n.annotations = annotations.NewAnnotationExtractor(n.store)

	if config.UpdateStatus {
		n.syncStatus = status.NewStatusSyncer(status.Config{
			Client:                 config.Client,
			PublishService:         config.PublishService,
			PublishStatusAddress:   config.PublishStatusAddress,
			IngressLister:          n.store,
			ElectionID:             config.ElectionID,
			IngressClass:           class.IngressClass,
			DefaultIngressClass:    class.DefaultClass,
			UpdateStatusOnShutdown: config.UpdateStatusOnShutdown,
			UseNodeInternalIP:      config.UseNodeInternalIP,
		})
	} else {
		glog.Warning("Update of ingress status is disabled (flag --update-status=false was specified)")
	}

	var onTemplateChange func()
	onTemplateChange = func() {
		template, err := ngx_template.NewTemplate(tmplPath, fs)
		if err != nil {
			// this error is different from the rest because it must be clear why nginx is not working
			glog.Errorf(`
-------------------------------------------------------------------------------
Error loading new template : %v
-------------------------------------------------------------------------------
`, err)
			return
		}

		n.t = template
		glog.Info("new NGINX template loaded")
		n.SetForceReload(true)
	}

	ngxTpl, err := ngx_template.NewTemplate(tmplPath, fs)
	if err != nil {
		glog.Fatalf("invalid NGINX template: %v", err)
	}

	n.t = ngxTpl

	// TODO: refactor
	if _, ok := fs.(filesystem.DefaultFs); !ok {
		watch.NewDummyFileWatcher(tmplPath, onTemplateChange)
	} else {

		_, err = watch.NewFileWatcher(tmplPath, onTemplateChange)
		if err != nil {
			glog.Fatalf("unexpected error creating file watcher: %v", err)
		}

		filesToWatch := []string{}
		err := filepath.Walk("/etc/nginx/geoip/", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			filesToWatch = append(filesToWatch, path)
			return nil
		})

		if err != nil {
			glog.Fatalf("unexpected error creating file watcher: %v", err)
		}

		for _, f := range filesToWatch {
			_, err = watch.NewFileWatcher(f, func() {
				glog.Info("file %v changed. Reloading NGINX", f)
				n.SetForceReload(true)
			})
			if err != nil {
				glog.Fatalf("unexpected error creating file watcher: %v", err)
			}
		}

	}

	return n
}

// NGINXController ...
type NGINXController struct {
	cfg *Configuration

	annotations annotations.Extractor

	recorder record.EventRecorder

	syncQueue *task.Queue

	syncStatus status.Sync

	syncRateLimiter flowcontrol.RateLimiter

	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock *sync.Mutex

	stopCh   chan struct{}
	updateCh *channels.RingChannel

	// ngxErrCh channel used to detect errors with the nginx processes
	ngxErrCh chan error

	// runningConfig contains the running configuration in the Backend
	runningConfig *ingress.Configuration

	forceReload int32

	t *ngx_template.Template

	binary   string
	resolver []net.IP

	stats        *statsCollector
	statusModule statusModule

	// returns true if IPV6 is enabled in the pod
	isIPV6Enabled bool

	isShuttingDown bool

	Proxy *TCPProxy

	store store.Storer

	fileSystem filesystem.Filesystem
}

// Start start a new NGINX master process running in foreground.
func (n *NGINXController) Start() {
	glog.Infof("starting Ingress controller")

	n.store.Run(n.stopCh)

	if n.syncStatus != nil {
		go n.syncStatus.Run()
	}

	cmd := exec.Command(n.binary, "-c", cfgPath)

	// put nginx in another process group to prevent it
	// to receive signals meant for the controller
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	if n.cfg.EnableSSLPassthrough {
		n.setupSSLProxy()
	}

	glog.Info("starting NGINX process...")
	n.start(cmd)

	go n.syncQueue.Run(time.Second, n.stopCh)
	// force initial sync
	n.syncQueue.Enqueue(&extensions.Ingress{})

	for {
		select {
		case err := <-n.ngxErrCh:
			if n.isShuttingDown {
				break
			}

			// if the nginx master process dies the workers continue to process requests,
			// passing checks but in case of updates in ingress no updates will be
			// reflected in the nginx configuration which can lead to confusion and report
			// issues because of this behavior.
			// To avoid this issue we restart nginx in case of errors.
			if process.IsRespawnIfRequired(err) {
				process.WaitUntilPortIsAvailable(n.cfg.ListenPorts.HTTP)
				// release command resources
				cmd.Process.Release()
				// start a new nginx master process if the controller is not being stopped
				cmd = exec.Command(n.binary, "-c", cfgPath)
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setpgid: true,
					Pgid:    0,
				}
				n.start(cmd)
			}
		case event := <-n.updateCh.Out():
			if n.isShuttingDown {
				break
			}
			if evt, ok := event.(store.Event); ok {
				glog.V(3).Infof("Event %v received - object %v", evt.Type, evt.Obj)
				if evt.Type == store.ConfigurationEvent {
					n.SetForceReload(true)
				}

				n.syncQueue.Enqueue(evt.Obj)
			} else {
				glog.Warningf("unexpected event type received %T", event)
			}
		case <-n.stopCh:
			break
		}
	}
}

// Stop gracefully stops the NGINX master process.
func (n *NGINXController) Stop() error {
	n.isShuttingDown = true

	n.stopLock.Lock()
	defer n.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if n.syncQueue.IsShuttingDown() {
		return fmt.Errorf("shutdown already in progress")
	}

	glog.Infof("shutting down controller queues")
	close(n.stopCh)
	go n.syncQueue.Shutdown()
	if n.syncStatus != nil {
		n.syncStatus.Shutdown()
	}

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
	timer := time.NewTicker(time.Second * 1)
	for range timer.C {
		if !process.IsNginxRunning() {
			glog.Info("NGINX process has stopped")
			timer.Stop()
			break
		}
	}

	return nil
}

func (n *NGINXController) start(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		glog.Fatalf("nginx error: %v", err)
		n.ngxErrCh <- err
		return
	}

	go func() {
		n.ngxErrCh <- cmd.Wait()
	}()
}

// DefaultEndpoint returns the default endpoint to be use as default server that returns 404.
func (n NGINXController) DefaultEndpoint() ingress.Endpoint {
	return ingress.Endpoint{
		Address: "127.0.0.1",
		Port:    fmt.Sprintf("%v", n.cfg.ListenPorts.Default),
		Target:  &apiv1.ObjectReference{},
	}
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

// OnUpdate is called periodically by syncQueue to keep the configuration in sync.
//
// 1. converts configmap configuration to custom configuration object
// 2. write the custom template (the complexity depends on the implementation)
// 3. write the configuration file
//
// returning nil implies the backend will be reloaded.
// if an error is returned means requeue the update
func (n *NGINXController) OnUpdate(ingressCfg ingress.Configuration) error {
	cfg := n.store.GetBackendConfiguration()
	cfg.Resolver = n.resolver

	if n.cfg.EnableSSLPassthrough {
		servers := []*TCPServer{}
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
			servers = append(servers, &TCPServer{
				Hostname:      pb.Hostname,
				IP:            svc.Spec.ClusterIP,
				Port:          port,
				ProxyProtocol: false,
			})
		}

		n.Proxy.ServerList = servers
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
	glog.V(2).Infof("maximum number of open file descriptors : %v", maxOpenFiles)
	if maxOpenFiles < 1024 {
		// this means the value of RLIMIT_NOFILE is too low.
		maxOpenFiles = 1024
	}

	setHeaders := map[string]string{}
	if cfg.ProxySetHeaders != "" {
		cmap, err := n.store.GetConfigMap(cfg.ProxySetHeaders)
		if err != nil {
			glog.Warningf("unexpected error reading configmap %v: %v", cfg.ProxySetHeaders, err)
		}

		setHeaders = cmap.Data
	}

	addHeaders := map[string]string{}
	if cfg.AddHeaders != "" {
		cmap, err := n.store.GetConfigMap(cfg.AddHeaders)
		if err != nil {
			glog.Warningf("unexpected error reading configmap %v: %v", cfg.AddHeaders, err)
		}

		addHeaders = cmap.Data
	}

	sslDHParam := ""
	if cfg.SSLDHParam != "" {
		secretName := cfg.SSLDHParam

		secret, err := n.store.GetSecret(secretName)
		if err != nil {
			glog.Warningf("unexpected error reading secret %v: %v", secretName, err)
		}

		nsSecName := strings.Replace(secretName, "/", "-", -1)

		dh, ok := secret.Data["dhparam.pem"]
		if ok {
			pemFileName, err := ssl.AddOrUpdateDHParam(nsSecName, dh, n.fileSystem)
			if err != nil {
				glog.Warningf("unexpected error adding or updating dhparam %v file: %v", nsSecName, err)
			} else {
				sslDHParam = pemFileName
			}
		}
	}

	cfg.SSLDHParam = sslDHParam

	tc := ngx_config.TemplateConfig{
		ProxySetHeaders:             setHeaders,
		AddHeaders:                  addHeaders,
		MaxOpenFiles:                maxOpenFiles,
		BacklogSize:                 sysctlSomaxconn(),
		Backends:                    ingressCfg.Backends,
		PassthroughBackends:         ingressCfg.PassthroughBackends,
		Servers:                     ingressCfg.Servers,
		TCPBackends:                 ingressCfg.TCPEndpoints,
		UDPBackends:                 ingressCfg.UDPEndpoints,
		HealthzURI:                  ngxHealthPath,
		CustomErrors:                len(cfg.CustomHTTPErrors) > 0,
		Cfg:                         cfg,
		IsIPV6Enabled:               n.isIPV6Enabled && !cfg.DisableIpv6,
		NginxStatusIpv4Whitelist:    cfg.NginxStatusIpv4Whitelist,
		NginxStatusIpv6Whitelist:    cfg.NginxStatusIpv6Whitelist,
		RedirectServers:             redirectServers,
		IsSSLPassthroughEnabled:     n.cfg.EnableSSLPassthrough,
		ListenPorts:                 n.cfg.ListenPorts,
		PublishService:              n.GetPublishService(),
		DynamicConfigurationEnabled: n.cfg.DynamicConfigurationEnabled,
		DisableLua:                  n.cfg.DisableLua,
	}

	content, err := n.t.Write(tc)

	if err != nil {
		return err
	}

	err = n.testTemplate(content)
	if err != nil {
		return err
	}

	if glog.V(2) {
		src, _ := ioutil.ReadFile(cfgPath)
		if !bytes.Equal(src, content) {
			tmpfile, err := ioutil.TempFile("", "new-nginx-cfg")
			if err != nil {
				return err
			}
			defer tmpfile.Close()
			err = ioutil.WriteFile(tmpfile.Name(), content, 0644)
			if err != nil {
				return err
			}

			// executing diff can return exit code != 0
			diffOutput, _ := exec.Command("diff", "-u", cfgPath, tmpfile.Name()).CombinedOutput()

			glog.Infof("NGINX configuration diff\n")
			glog.Infof("%v\n", string(diffOutput))

			// Do not use defer to remove the temporal file.
			// This is helpful when there is an error in the
			// temporal configuration (we can manually inspect the file).
			// Only remove the file when no error occurred.
			os.Remove(tmpfile.Name())
		}
	}

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
	// See https://github.com/kubernetes/ingress-nginxs/issues/623 for an explanation
	wordSize := 8 // Assume 64 bit CPU
	n := longestString + 2
	aligned := (n + wordSize - 1) & ^(wordSize - 1)
	rawSize := wordSize + wordSize + aligned
	return nextPowerOf2(rawSize)
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

func (n *NGINXController) setupSSLProxy() {
	sslPort := n.cfg.ListenPorts.HTTPS
	proxyPort := n.cfg.ListenPorts.SSLProxy

	glog.Info("starting TLS proxy for SSL passthrough")
	n.Proxy = &TCPProxy{
		Default: &TCPServer{
			Hostname:      "localhost",
			IP:            "127.0.0.1",
			Port:          proxyPort,
			ProxyProtocol: true,
		},
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", sslPort))
	if err != nil {
		glog.Fatalf("%v", err)
	}

	proxyList := &proxyproto.Listener{Listener: listener}

	// start goroutine that accepts tcp connections in port 443
	go func() {
		for {
			var conn net.Conn
			var err error

			if n.store.GetBackendConfiguration().UseProxyProtocol {
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
			go n.Proxy.Handle(conn)
		}
	}()
}

// IsDynamicConfigurationEnough decides if the new configuration changes can be dynamically applied without reloading
func (n *NGINXController) IsDynamicConfigurationEnough(pcfg *ingress.Configuration) bool {
	var copyOfRunningConfig ingress.Configuration = *n.runningConfig
	var copyOfPcfg ingress.Configuration = *pcfg

	copyOfRunningConfig.Backends = []*ingress.Backend{}
	copyOfPcfg.Backends = []*ingress.Backend{}

	return copyOfRunningConfig.Equal(&copyOfPcfg)
}

// ConfigureDynamically JSON encodes new Backends and POSTs it to an internal HTTP endpoint
// that is handled by Lua
func (n *NGINXController) ConfigureDynamically(pcfg *ingress.Configuration) error {
	backends := make([]*ingress.Backend, len(pcfg.Backends))

	for i, backend := range pcfg.Backends {
		cleanedupBackend := *backend
		cleanedupBackend.Service = nil
		backends[i] = &cleanedupBackend
	}

	buf, err := json.Marshal(backends)
	if err != nil {
		return err
	}

	glog.V(2).Infof("posting backends configuration: %s", buf)

	url := fmt.Sprintf("http://localhost:%d/configuration/backends", n.cfg.ListenPorts.Status)
	resp, err := http.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			glog.Warningf("error while closing response body: \n%v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected error code: %d", resp.StatusCode)
	}

	return nil
}
