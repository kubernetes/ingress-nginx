/* Copyright 2017 Cody Maloney */

package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"

	"k8s.io/ingress/core/pkg/ingress/controller"
)

const baseEnvoyConfig = `
{
  "admin": {
    "access_log_path": "/dev/null",
    "address": "tcp://0.0.0.0:10252"
  },
  "cluster_manager": {
    "cds": {
      "cluster": {
        "connect_timeout_ms": 200,
        "hosts": [
          {
            "url": "tcp://127.0.0.1:8080"
          }
        ],
        "lb_type": "random",
        "name": "local_cds",
        "type": "static"
      },
      "refresh_delay_ms": 500
    },
    "clusters": [
      {
        "connect_timeout_ms": 200,
        "hosts": [
          {
            "url": "tcp://127.0.0.1:8080"
          }
        ],
        "lb_type": "random",
        "name": "local_lds",
        "type": "static"
      }
    ],
    "sds": {
      "cluster": {
        "connect_timeout_ms": 200,
        "hosts": [
          {
            "url": "tcp://127.0.0.1:8080"
          }
        ],
        "lb_type": "random",
        "name": "local_sds",
        "type": "static"
      },
      "refresh_delay_ms": 500
    }
  },
  "lds": {
    "cluster": "local_lds",
    "refresh_delay_ms": 500
  },
  "listeners": []
}
`

type killMsg struct {
	Exited bool
	Err    error
}

type envoyRunner struct {
	cmd      *exec.Cmd
	wg       sync.WaitGroup
	killChan chan killMsg
}

func NewEnvoyRunner() *envoyRunner {
	err := ioutil.WriteFile("core_config.json", []byte(baseEnvoyConfig), 0644)
	if err != nil {
		log.Panicf("Error writing out core_config.json for envoy: %s", err)
	}

	return &envoyRunner{
		cmd: exec.Command(
			"envoy",
			"-c", "core_config.json",
			"--service-node", "TODO_NODE",
			"--service-cluster", "TODO_CLUSTER"),
	}
}

func (er *envoyRunner) Start() {
	log.Print("Starting envoy")

	// Setup stdout, stderr
	stdout, err := er.cmd.StdoutPipe()
	if err != nil {
		log.Panicf("Unable to setup stdout from envoy subprocess: %s\n", err)
	}
	stderr, err := er.cmd.StderrPipe()
	if err != nil {
		log.Panicf("Unable to setup stderr from envoy subprocess: %s\n", err)
	}
	stdin, err := er.cmd.StdinPipe()
	if err != nil {
		log.Panicf("Unable to setup stdin from envoy subprocess: %s\n", err)
	}
	stdin.Close()

	readThenClose := func(out io.Writer, in io.ReadCloser) {
		if _, err := io.Copy(out, in); err != nil {
			log.Panicf("Error reading from pipe: %s\n", err)
		}
		in.Close()
	}
	go readThenClose(os.Stdout, stdout)
	go readThenClose(os.Stderr, stderr)

	if err := er.cmd.Start(); err != nil {
		log.Panicf("Error starting envoy: %v", err)
	}

	er.wg.Add(1)
	go func() {
		er.killChan <- killMsg{Exited: true, Err: er.cmd.Wait()}
	}()

	// Wait in the background for a kill request or process exit. On kill request, send a signal to
	// tell envoy to shutdown.
	go func() {
		for v := <-er.killChan; true; {
			// Process exited
			if v.Exited {
				log.Println("envoy exited")
				if v.Err != nil {
					log.Panicf("envoy exited with an error code: %v", v.Err)
				}
				er.wg.Done()
				return
			}
			// TODO(cmaloney): Make a "max time to wait" for exit.
			er.cmd.Process.Signal(os.Interrupt)
		}
	}()

}

func (er *envoyRunner) Stop() {
	// Send "shutdown" to "waiter" process, causing it to send a sigterm to the process, then wait
	er.killChan <- killMsg{Exited: false}
	er.wg.Wait()
}

func main() {

	// Make an EnvoyDiscoveryService
	// Make a kubernetes Ingress Controller that feeds its constructed results
	// to the discovery service.

	er := NewEnvoyRunner()
	ds := NewDiscoveryService(er.Start)
	ec := NewEnvoyController(ds)

	ic := controller.NewIngressController(ec)
	defer func() {
		log.Printf("Shutting down ingress controller...")
		er.Stop()
		ic.Stop()
		ds.Stop()
	}()
	ic.Start()
}
