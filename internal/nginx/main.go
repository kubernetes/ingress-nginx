/*
Copyright 2019 The Kubernetes Authors.

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

package nginx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tv42/httpunix"
)

// PID defines the location of the pid file used by NGINX
var PID = "/tmp/nginx.pid"

// StatusSocket defines the location of the unix socket used by NGINX for the status server
var StatusSocket = "/tmp/nginx-status-server.sock"

// HealthPath defines the path used to define the health check location in NGINX
var HealthPath = "/healthz"

// StatusPath defines the path used to expose the NGINX status page
// http://nginx.org/en/docs/http/ngx_http_stub_status_module.html
var StatusPath = "/nginx_status"

// StreamSocket defines the location of the unix socket used by NGINX for the NGINX stream configuration socket
var StreamSocket = "/tmp/ingress-stream.sock"

var statusLocation = "nginx-status"

var socketClient = buildUnixSocketClient()

// NewGetStatusRequest creates a new GET request to the internal NGINX status server
func NewGetStatusRequest(path string) (int, []byte, error) {
	url := fmt.Sprintf("http+unix://%v%v", statusLocation, path)
	res, err := socketClient.Get(url)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	return res.StatusCode, data, nil
}

// NewPostStatusRequest creates a new POST request to the internal NGINX status server
func NewPostStatusRequest(path, contentType string, data interface{}) (int, []byte, error) {
	url := fmt.Sprintf("http+unix://%v%v", statusLocation, path)

	buf, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}

	res, err := socketClient.Post(url, contentType, bytes.NewReader(buf))
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	return res.StatusCode, body, nil
}

// GetServerBlock takes an nginx.conf file and a host and tries to find the server block for that host
func GetServerBlock(conf string, host string) (string, error) {
	startMsg := fmt.Sprintf("## start server %v\n", host)
	endMsg := fmt.Sprintf("## end server %v", host)

	blockStart := strings.Index(conf, startMsg)
	if blockStart < 0 {
		return "", fmt.Errorf("Host %v was not found in the controller's nginx.conf", host)
	}
	blockStart = blockStart + len(startMsg)

	blockEnd := strings.Index(conf, endMsg)
	if blockEnd < 0 {
		return "", fmt.Errorf("The end of the host server block could not be found, but the beginning was")
	}

	return conf[blockStart:blockEnd], nil
}

// ReadNginxConf reads the nginx configuration file into a string
func ReadNginxConf() (string, error) {
	return ReadFileToString("/etc/nginx/nginx.conf")
}

// ReadFileToString reads any file into a string
func ReadFileToString(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func buildUnixSocketClient() *http.Client {
	u := &httpunix.Transport{
		DialTimeout:           1 * time.Second,
		RequestTimeout:        10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	u.RegisterLocation(statusLocation, StatusSocket)

	return &http.Client{
		Transport: u,
	}
}
