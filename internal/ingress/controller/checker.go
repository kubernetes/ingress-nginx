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
	"os"
	"strconv"
	"strings"

	"github.com/ncabatoff/process-exporter/proc"

	"k8s.io/ingress-nginx/internal/nginx"
)

// Name returns the healthcheck name
func (n NGINXController) Name() string {
	return "nginx-ingress-controller"
}

// Check returns if the nginx healthz endpoint is returning ok (status code 200)
func (n *NGINXController) Check(_ *http.Request) error {
	if n.isShuttingDown {
		return fmt.Errorf("the ingress controller is shutting down")
	}

	// check the nginx master process is running
	fs, err := proc.NewFS("/proc", false)
	if err != nil {
		return fmt.Errorf("reading /proc directory: %w", err)
	}

	f, err := os.ReadFile(nginx.PID)
	if err != nil {
		return fmt.Errorf("reading %v: %w", nginx.PID, err)
	}

	pid, err := strconv.Atoi(strings.TrimRight(string(f), "\r\n"))
	if err != nil {
		return fmt.Errorf("reading NGINX PID from file %v: %w", nginx.PID, err)
	}

	_, err = fs.Proc(pid)
	if err != nil {
		return fmt.Errorf("checking for NGINX process with PID %v: %w", pid, err)
	}

	statusCode, _, err := nginx.NewGetStatusRequest("/is-dynamic-lb-initialized")
	if err != nil {
		return fmt.Errorf("checking if the dynamic load balancer started: %w", err)
	}

	if statusCode != 200 {
		return fmt.Errorf("dynamic load balancer not started")
	}

	return nil
}
