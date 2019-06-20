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
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/kubernetes/pkg/util/filesystem"

	"k8s.io/ingress-nginx/internal/file"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/nginx"
)

func TestNginxCheck(t *testing.T) {
	mux := http.NewServeMux()

	listener, err := net.Listen("unix", nginx.StatusSocket)
	if err != nil {
		t.Errorf("crating unix listener: %s", err)
	}
	defer listener.Close()
	defer os.Remove(nginx.StatusSocket)

	server := &httptest.Server{
		Listener: listener,
		Config: &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "ok")
			}),
		},
	}
	defer server.Close()
	server.Start()

	// mock filesystem
	fs := filesystem.DefaultFs{}

	n := &NGINXController{
		cfg: &Configuration{
			ListenPorts: &ngx_config.ListenPorts{},
		},
		fileSystem: fs,
	}

	t.Run("no pid or process", func(t *testing.T) {
		if err := callHealthz(true, mux); err == nil {
			t.Error("expected an error but none returned")
		}
	})

	// create pid file
	fs.MkdirAll("/tmp", file.ReadWriteByUser)
	pidFile, err := fs.Create(nginx.PID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("no process", func(t *testing.T) {
		if err := callHealthz(true, mux); err == nil {
			t.Error("expected an error but none returned")
		}
	})

	// start dummy process to use the PID
	cmd := exec.Command("sleep", "3600")
	cmd.Start()
	pid := cmd.Process.Pid
	defer cmd.Process.Kill()
	go func() {
		cmd.Wait()
	}()

	pidFile.Write([]byte(fmt.Sprintf("%v", pid)))
	pidFile.Close()

	healthz.InstallHandler(mux, n)

	t.Run("valid request", func(t *testing.T) {
		if err := callHealthz(false, mux); err != nil {
			t.Error(err)
		}
	})

	// pollute pid file
	pidFile.Write([]byte(fmt.Sprint("999999")))
	pidFile.Close()

	t.Run("bad pid", func(t *testing.T) {
		if err := callHealthz(true, mux); err == nil {
			t.Error("expected an error but none returned")
		}
	})
}

func callHealthz(expErr bool, mux *http.ServeMux) error {
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		return fmt.Errorf("healthz error: %v", err)
	}

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if expErr && w.Code != http.StatusInternalServerError {
		return fmt.Errorf("expected an error")
	}

	if w.Body.String() != "ok" {
		return fmt.Errorf("healthz error: %v", w.Body.String())
	}

	if w.Code != http.StatusOK {
		return fmt.Errorf("expected status code 200 but %v returned", w.Code)
	}

	return nil
}
