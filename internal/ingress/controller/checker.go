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

package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ncabatoff/process-exporter/proc"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"k8s.io/ingress-nginx/internal/nginx"
)

// Name returns the healthcheck name
func (n NGINXController) Name() string {
	return "nginx-ingress-controller"
}

// Check returns if the nginx healthz endpoint is returning ok (status code 200)
func (n *NGINXController) Check(_ *http.Request) error {
	statusCode, _, err := nginx.NewGetStatusRequest(nginx.HealthPath)
	if err != nil {
		klog.Errorf("healthcheck error: %v", err)
		return err
	}

	if statusCode != 200 {
		klog.Errorf("healthcheck error: %v", statusCode)
		return fmt.Errorf("ingress controller is not healthy")
	}

	statusCode, _, err = nginx.NewGetStatusRequest("/is-dynamic-lb-initialized")
	if err != nil {
		klog.Errorf("healthcheck error: %v", err)
		return err
	}

	if statusCode != 200 {
		klog.Errorf("healthcheck error: %v", statusCode)
		return fmt.Errorf("dynamic load balancer not started")
	}

	// check the nginx master process is running
	fs, err := proc.NewFS("/proc", false)
	if err != nil {
		return errors.Wrap(err, "unexpected error reading /proc directory")
	}
	f, err := n.fileSystem.ReadFile(nginx.PID)
	if err != nil {
		return errors.Wrapf(err, "unexpected error reading %v", nginx.PID)
	}
	pid, err := strconv.Atoi(strings.TrimRight(string(f), "\r\n"))
	if err != nil {
		return errors.Wrapf(err, "unexpected error reading the nginx PID from %v", nginx.PID)
	}
	_, err = fs.NewProc(pid)
	return err
}
