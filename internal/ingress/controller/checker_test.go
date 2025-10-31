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
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"k8s.io/apiserver/pkg/server/healthz"

	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/util/file"
)

func TestNginxCheck(t *testing.T) {
	tests := []struct {
		healthzPath string
	}{
		{"/healthz"},
		{"/not-healthz"},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("health path: %s", tt.healthzPath)
		t.Run(testName, func(t *testing.T) {
			mux := http.NewServeMux()

			// Status server
			listener, err := tryListen("tcp", fmt.Sprintf(":%v", nginx.StatusPort))
			if err != nil {
				t.Fatalf("creating tcp listener: %s", err)
			}
			defer listener.Close()
			//nolint:gosec // Ignore not configured ReadHeaderTimeout in testing
			server := &httptest.Server{
				Listener: listener,
				Config: &http.Server{
					Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusOK)
						fmt.Fprintf(w, "ok")
					}),
				},
			}
			defer server.Close()
			server.Start()

			n := &NGINXController{
				cfg: &Configuration{
					ListenPorts: &ngx_config.ListenPorts{
						HTTP: 80,
					},
				},
			}

			// http server
			httpListener, err := tryListen("tcp", fmt.Sprintf(":%v", n.cfg.ListenPorts.HTTP))
			if err != nil {
				t.Fatalf("creating tcp listener: %s", err)
			}
			defer httpListener.Close()

			httpServer := &httptest.Server{
				Listener: httpListener,
				Config: &http.Server{
					Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						fmt.Fprintf(w, "ok")
					}),
				},
			}
			defer httpServer.Close()
			httpServer.Start()

			t.Run("no pid or process", func(t *testing.T) {
				if err := callHealthz(true, tt.healthzPath, mux); err == nil {
					t.Error("expected an error but none returned")
				}
			})

			// create pid file
			if err := os.MkdirAll("/tmp/nginx", file.ReadWriteByUser); err != nil {
				t.Errorf("unexpected error creating pid file: %v", err)
			}

			pidFile, err := os.Create(nginx.PID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			t.Run("no process", func(t *testing.T) {
				if err := callHealthz(true, tt.healthzPath, mux); err == nil {
					t.Error("expected an error but none returned")
				}
			})

			// start dummy process to use the PID
			cmd := exec.Command("sleep", "3600")
			if err := cmd.Start(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			pid := cmd.Process.Pid
			defer func() {
				if err := cmd.Process.Kill(); err != nil {
					t.Errorf("unexpected error killing the process: %v", err)
				}
			}()
			go func() {
				cmd.Wait() //nolint:errcheck // Ignore the error
			}()

			if _, err := fmt.Fprintf(pidFile, "%v", pid); err != nil {
				t.Errorf("unexpected error writing the pid file: %v", err)
			}

			pidFile.Close()

			healthz.InstallPathHandler(mux, tt.healthzPath, n)

			t.Run("valid request", func(t *testing.T) {
				if err := callHealthz(false, tt.healthzPath, mux); err != nil {
					t.Error(err)
				}
			})

			// pollute pid file
			pidFile.WriteString("999999") //nolint:errcheck // Ignore the error
			pidFile.Close()

			t.Run("bad pid", func(t *testing.T) {
				if err := callHealthz(true, tt.healthzPath, mux); err == nil {
					t.Error("expected an error but none returned")
				}
			})
		})
	}
}

func callHealthz(expErr bool, healthzPath string, mux *http.ServeMux) error {
	req, err := http.NewRequest(http.MethodGet, healthzPath, http.NoBody)
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
