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
	"time"

	"github.com/ncabatoff/process-exporter/proc"
	"github.com/pkg/errors"
)

const nginxPID = "/tmp/nginx.pid"

// Name returns the healthcheck name
func (n NGINXController) Name() string {
	return "nginx-ingress-controller"
}

// Check returns if the nginx healthz endpoint is returning ok (status code 200)
func (n *NGINXController) Check(_ *http.Request) error {

	url := fmt.Sprintf("http://127.0.0.1:%v%v", n.cfg.ListenPorts.Status, ngxHealthPath)
	timeout := n.cfg.HealthCheckTimeout
	statusCode, err := simpleGet(url, timeout)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("ingress controller is not healthy")
	}

	url = fmt.Sprintf("http://127.0.0.1:%v/is-dynamic-lb-initialized", n.cfg.ListenPorts.Status)
	statusCode, err = simpleGet(url, timeout)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("dynamic load balancer not started")
	}

	// check the nginx master process is running
	fs, err := proc.NewFS("/proc", false)
	if err != nil {
		return errors.Wrap(err, "unexpected error reading /proc directory")
	}
	f, err := n.fileSystem.ReadFile(nginxPID)
	if err != nil {
		return errors.Wrapf(err, "unexpected error reading %v", nginxPID)
	}
	pid, err := strconv.Atoi(strings.TrimRight(string(f), "\r\n"))
	if err != nil {
		return errors.Wrapf(err, "unexpected error reading the nginx PID from %v", nginxPID)
	}
	_, err = fs.NewProc(pid)

	return err
}

func simpleGet(url string, timeout time.Duration) (int, error) {
	client := &http.Client{
		Timeout:   timeout * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, err
	}

	res, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	return res.StatusCode, nil
}
