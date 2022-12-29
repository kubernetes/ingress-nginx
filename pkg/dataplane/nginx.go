/*
Copyright 2022 The Kubernetes Authors.

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

package dataplane

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/process"
	ngx_template "k8s.io/ingress-nginx/internal/ingress/controller/template"
	"k8s.io/ingress-nginx/internal/ingress/metric"
	ing_net "k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/internal/net/dns"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/dataplane/grpcclient"
	"k8s.io/ingress-nginx/pkg/tcpproxy"
	"k8s.io/ingress-nginx/pkg/util/file"
	ingressruntime "k8s.io/ingress-nginx/pkg/util/runtime"
	"k8s.io/klog/v2"
)

// NewNGINXController creates a new NGINX Ingress controller.
func NewNGINXConfigurer(ingressconfig *Configuration, mc metric.Collector) *NGINXConfigurer {

	h, err := dns.GetSystemNameServers()
	if err != nil {
		klog.Warningf("Error reading system nameservers: %v", err)
	}

	errCh := make(chan error)
	grpcErrCh := make(chan error)
	// TODO: Get this from configuration @Volatus is checking this
	grpcconf := grpcclient.Config{
		Options: grpcclient.GRPCDialOptions{
			Address: ingressconfig.GRPCAddress,
		},
		ErrorCh:   errCh,
		GRPCErrCh: grpcErrCh,
	}
	grpccl, err := grpcclient.NewGRPCClient(grpcconf)
	if err != nil {
		klog.Fatalf("error creating GRPC Client: %s", err)
	}

	n := &NGINXConfigurer{
		isIPV6Enabled: ing_net.IsIPv6Enabled(),
		resolver:      h,
		cfg:           ingressconfig,
		stopCh:        make(chan struct{}),
		ngxErrCh:      errCh,
		grpcErrCh:     grpcErrCh,
		configureLock: &sync.Mutex{},
		stopLock:      &sync.Mutex{},
		// TOOD: Right now we will receive the full configuration, but we may want to receive and validate just checksums
		templateConfig: new(config.TemplateConfig),
		Proxy:          &tcpproxy.TCPProxy{},

		BacklogSize: ingressruntime.SysctlSomaxconn(),

		metricCollector: mc,

		GRPCClient: grpccl,
		command:    NewNginxCommand(),
	}

	// TODO: This seems wrong, it is configured in ConfigMap and should be sent as part of the
	// struct just to be calculated during the template generation
	if ingressconfig.MaxWorkerOpenFiles == 0 {
		// the limit of open files is per worker process
		// and we leave some room to avoid consuming all the FDs available
		maxOpenFiles := ingressruntime.RlimitMaxNumFiles() - 1024
		klog.V(3).InfoS("Maximum number of open file descriptors", "value", maxOpenFiles)
		if maxOpenFiles < 1024 {
			// this means the value of RLIMIT_NOFILE is too low.
			maxOpenFiles = 1024
		}
		klog.V(3).InfoS("Adjusting MaxWorkerOpenFiles variable", "value", maxOpenFiles)
		ingressconfig.MaxWorkerOpenFiles = maxOpenFiles
	}

	// TODO: Send MaxWorkerOpenfiles and BacklogSize to template generation

	onTemplateChange := func() {
		template, err := ngx_template.NewTemplate(nginx.TemplatePath)
		if err != nil {
			// this error is different from the rest because it must be clear why nginx is not working
			klog.ErrorS(err, "Error loading new template")
			return
		}

		n.t = template
		klog.InfoS("New NGINX configuration template loaded")
		newConfig := *n.templateConfig
		n.GRPCClient.ConfigCh <- &newConfig
	}

	ngxTpl, err := ngx_template.NewTemplate(nginx.TemplatePath)
	if err != nil {
		klog.Fatalf("Invalid NGINX configuration template: %v", err)
	}

	n.t = ngxTpl

	_, err = file.NewFileWatcher(nginx.TemplatePath, onTemplateChange)
	if err != nil {
		klog.Fatalf("Error creating file watcher for %v: %v", nginx.TemplatePath, err)
	}

	filesToWatch := []string{}
	err = filepath.Walk("/etc/nginx/geoip/", func(path string, info os.FileInfo, err error) error {
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
		klog.Fatalf("Error creating file watchers: %v", err)
	}

	for _, f := range filesToWatch {
		_, err = file.NewFileWatcher(f, func() {
			klog.InfoS("File changed detected. Reloading NGINX", "path", f)
			newConfig := *n.templateConfig
			n.GRPCClient.ConfigCh <- &newConfig
		})
		if err != nil {
			klog.Fatalf("Error creating file watcher for %v: %v", f, err)
		}
	}

	return n
}

// Start starts a new NGINX master process running in the foreground.
func (n *NGINXConfigurer) Start() {
	klog.InfoS("Starting NGINX Ingress configurer")

	cmd := n.command.ExecCommand()

	// put NGINX in another process group to prevent it
	// to receive signals meant for the controller
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	/* TODO: Add SSL Passthrough correctly
	if n.cfg.EnableSSLPassthrough {
		n.setupSSLProxy()
	}
	*/

	klog.InfoS("Starting NGINX process")
	n.start(cmd)
	n.GRPCClient.Start()

	// In case of error the temporal configuration file will
	// be available up to five minutes after the error
	// TODO: Do we need this cleanup?
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			err := cleanTempNginxCfg()
			if err != nil {
				klog.ErrorS(err, "Unexpected error removing temporal configuration files")
			}
		}
	}()

	for {
		select {
		case err := <-n.ngxErrCh:
			if n.isShuttingDown {
				return
			}

			// if the nginx master process dies, the workers continue to process requests
			// until the failure of the configured livenessProbe and restart of the pod.
			if process.IsRespawnIfRequired(err) {
				return
			}
			klog.Fatalf("Dataplane got an unexpected error, exiting: %s", err)

		case err := <-n.grpcErrCh:
			// TODO: Maybe we need a mutex or some way to not start multiple times (a map of channels to stop goroutines? a pubsub?)
			klog.Warningf("Detected error in gRPC connection, restaring the connection: %s", err)
			n.GRPCClient.Start()
		case cfg := <-n.GRPCClient.ConfigCh:
			if n.isShuttingDown {
				break
			}
			n.syncIngress(cfg)

		case <-n.stopCh:
			return
		}
	}
}

// Stop gracefully stops the NGINX master process.
func (n *NGINXConfigurer) Stop() error {
	n.isShuttingDown = true

	n.stopLock.Lock()
	defer n.stopLock.Unlock()

	time.Sleep(time.Duration(n.cfg.ShutdownGracePeriod) * time.Second)

	// send stop signal to NGINX
	klog.InfoS("Stopping NGINX process")
	cmd := n.command.ExecCommand("-s", "quit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	/*if err := n.GRPCClient.ShutdownFunc(); err != nil {
		return err
	}*/ // TODO: Should detect if connection is already closed, otherwise it will panic
	// wait for the NGINX process to terminate
	timer := time.NewTicker(time.Second * 1)
	for range timer.C {
		if !nginx.IsRunning() {
			klog.InfoS("NGINX process has stopped")
			timer.Stop()
			break
		}
	}

	return nil
}

func (n *NGINXConfigurer) start(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		klog.Fatalf("NGINX error: %v", err)
		n.ngxErrCh <- err
		return
	}

	go func() {
		n.ngxErrCh <- cmd.Wait()
	}()
}

// TODO: Shall we test this in both places? (CP and DP)
// testTemplate checks if the NGINX configuration inside the byte array is valid
// running the command "nginx -t" using a temporal file.
func (n NGINXConfigurer) testTemplate(cfg []byte) error {
	if len(cfg) == 0 {
		return fmt.Errorf("invalid NGINX configuration (empty)")
	}
	tmpDir := os.TempDir() + "/nginx"
	tmpfile, err := os.CreateTemp(tmpDir, tempNginxPattern)
	if err != nil {
		return err
	}
	defer tmpfile.Close()
	err = os.WriteFile(tmpfile.Name(), cfg, file.ReadWriteByUser)
	if err != nil {
		return err
	}
	out, err := n.command.Test(tmpfile.Name())
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
