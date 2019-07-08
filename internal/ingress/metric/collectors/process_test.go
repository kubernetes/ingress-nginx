/*
Copyright 2018 The Kubernetes Authors.

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

package collectors

import (
	"os/exec"
	"syscall"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestProcessCollector(t *testing.T) {
	cases := []struct {
		name    string
		metrics []string
	}{
		{
			name:    "should return metrics",
			metrics: []string{"nginx_ingress_controller_nginx_process_num_procs"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name = "sleep"
			binary = "/bin/sleep"

			cmd := exec.Command(binary, "1000000")
			err := cmd.Start()
			if err != nil {
				t.Errorf("unexpected error creating dummy process: %v", err)
			}

			done := make(chan struct{})
			go func() {
				cmd.Wait()
				status := cmd.ProcessState.Sys().(syscall.WaitStatus)
				if status.Signaled() {
					t.Logf("Signal: %v", status.Signal())
				} else {
					t.Logf("Status: %v", status.ExitStatus())
				}
				done <- struct{}{}
			}()

			cm, err := NewNGINXProcess("pod", "default", "nginx")
			if err != nil {
				t.Errorf("unexpected error creating nginx status collector: %v", err)
				t.FailNow()
			}

			go cm.Start()

			defer func() {
				cm.Stop()

				cmd.Process.Kill()
				<-done
				close(done)
			}()

			reg := prometheus.NewPedanticRegistry()
			if err := reg.Register(cm); err != nil {
				t.Errorf("registering collector failed: %s", err)
			}

			metrics, err := reg.Gather()
			if err != nil {
				t.Errorf("gathering metrics failed: %s", err)
			}

			m := filterMetrics(metrics, c.metrics)

			if *m[0].GetMetric()[0].Gauge.Value < 0 {
				t.Errorf("number of process should be > 0")
			}

			reg.Unregister(cm)
		})
	}
}
