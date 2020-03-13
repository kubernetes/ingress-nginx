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
	"fmt"
	"os/exec"
	"syscall"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/sysctl"
)

// newUpstream creates an upstream without servers.
func newUpstream(name string) *ingress.Backend {
	return &ingress.Backend{
		Name:      name,
		Endpoints: []ingress.Endpoint{},
		Service:   &api.Service{},
		SessionAffinity: ingress.SessionAffinityConfig{
			CookieSessionAffinity: ingress.CookieSessionAffinity{
				Locations: make(map[string][]string),
			},
		},
	}
}

// upstreamName returns a formatted upstream name based on namespace, service, and port
func upstreamName(namespace string, service string, port intstr.IntOrString) string {
	return fmt.Sprintf("%v-%v-%v", namespace, service, port.String())
}

// sysctlSomaxconn returns the maximum number of connections that can be queued
// for acceptance (value of net.core.somaxconn)
// http://nginx.org/en/docs/http/ngx_http_core_module.html#listen
func sysctlSomaxconn() int {
	maxConns, err := sysctl.New().GetSysctl("net/core/somaxconn")
	if err != nil || maxConns < 512 {
		klog.V(3).Infof("net.core.somaxconn=%v (using system default)", maxConns)
		return 511
	}

	return maxConns
}

// rlimitMaxNumFiles returns hard limit for RLIMIT_NOFILE
func rlimitMaxNumFiles() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		klog.Errorf("Error reading system maximum number of open file descriptors (RLIMIT_NOFILE): %v", err)
		return 0
	}
	klog.V(2).Infof("rlimit.max=%v", rLimit.Max)
	return int(rLimit.Max)
}

const (
	defBinary = "/usr/local/nginx/sbin/nginx"
	cfgPath   = "/etc/nginx/nginx.conf"
)

// NginxExecTester defines the interface to execute
// command like reload or test configuration
type NginxExecTester interface {
	ExecCommand(args ...string) *exec.Cmd
	Test(cfg string) ([]byte, error)
}

// NginxCommand stores context around a given nginx executable path
type NginxCommand struct {
	Binary string
}

// NewNginxCommand returns a new NginxCommand from which path
// has been detected from environment variable NGINX_BINARY or default
func NewNginxCommand() NginxCommand {
	return NginxCommand{
		Binary: defBinary,
	}
}

// ExecCommand instanciates an exec.Cmd object to call nginx program
func (nc NginxCommand) ExecCommand(args ...string) *exec.Cmd {
	cmdArgs := []string{}

	cmdArgs = append(cmdArgs, "-c", cfgPath)
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(nc.Binary, cmdArgs...)
}

// Test checks if config file is a syntax valid nginx configuration
func (nc NginxCommand) Test(cfg string) ([]byte, error) {
	return exec.Command(nc.Binary, "-c", cfg, "-t").CombinedOutput()
}
