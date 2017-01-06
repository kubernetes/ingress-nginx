/*
Copyright 2014 The Kubernetes Authors.

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

// e2e.go runs the e2e test suite. No non-standard package dependencies; call with "go run".
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	build      = flag.Bool("build", true, "Build the backends images indicated by the env var BACKENDS required to run e2e tests.")
	up         = flag.Bool("up", true, "Creates a kubernetes cluster using hyperkube (containerized kubelet).")
	down       = flag.Bool("down", true, "destroys the created cluster.")
	test       = flag.Bool("test", true, "Run Ginkgo tests.")
	dump       = flag.String("dump", "", "If set, dump cluster logs to this location on test or cluster-up failure")
	testArgs   = flag.String("test-args", "", "Space-separated list of arguments to pass to Ginkgo test runner.")
	deployment = flag.String("deployment", "bash", "up/down mechanism")
	verbose    = flag.Bool("v", false, "If true, print all command output.")
)

func appendError(errs []error, err error) []error {
	if err != nil {
		return append(errs, err)
	}
	return errs
}

func validWorkingDirectory() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get pwd: %v", err)
	}
	acwd, err := filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("failed to convert %s to an absolute path: %v", cwd, err)
	}
	if !strings.Contains(filepath.Base(acwd), "ingress") {
		return fmt.Errorf("must run from git root directory: %v", acwd)
	}
	return nil
}

type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	ClassName string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Time      float64  `xml:"time,attr"`
	Failure   string   `xml:"failure,omitempty"`
}

type TestSuite struct {
	XMLName  xml.Name `xml:"testsuite"`
	Failures int      `xml:"failures,attr"`
	Tests    int      `xml:"tests,attr"`
	Time     float64  `xml:"time,attr"`
	Cases    []TestCase
}

var suite TestSuite

func xmlWrap(name string, f func() error) error {
	start := time.Now()
	err := f()
	duration := time.Since(start)
	c := TestCase{
		Name:      name,
		ClassName: "e2e.go",
		Time:      duration.Seconds(),
	}
	if err != nil {
		c.Failure = err.Error()
		suite.Failures++
	}
	suite.Cases = append(suite.Cases, c)
	suite.Tests++
	return err
}

func writeXML(start time.Time) {
	suite.Time = time.Since(start).Seconds()
	out, err := xml.MarshalIndent(&suite, "", "    ")
	if err != nil {
		log.Fatalf("Could not marshal XML: %s", err)
	}
	path := filepath.Join(*dump, "junit_runner.xml")
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Could not create file: %s", err)
	}
	defer f.Close()
	if _, err := f.WriteString(xml.Header); err != nil {
		log.Fatalf("Error writing XML header: %s", err)
	}
	if _, err := f.Write(out); err != nil {
		log.Fatalf("Error writing XML data: %s", err)
	}
	log.Printf("Saved XML output to %s.", path)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if err := validWorkingDirectory(); err != nil {
		log.Fatalf("Called from invalid working directory: %v", err)
	}

	deploy, err := getDeployer()
	if err != nil {
		log.Fatalf("Error creating deployer: %v", err)
	}

	if err := run(deploy); err != nil {
		log.Fatalf("Something went wrong: %s", err)
	}
}

func run(deploy deployer) error {
	if *dump != "" {
		defer writeXML(time.Now())
	}

	if *build {
		if err := xmlWrap("Build", Build); err != nil {
			return fmt.Errorf("error building: %s", err)
		}
	}

	if *up {
		if err := xmlWrap("TearDown", deploy.Down); err != nil {
			return fmt.Errorf("error tearing down previous cluster: %s", err)
		}
	}

	var errs []error

	if *up {
		// If we tried to bring the cluster up, make a courtesy
		// attempt to bring it down so we're not leaving resources around.
		//
		// TODO: We should try calling deploy.Down exactly once. Though to
		// stop the leaking resources for now, we want to be on the safe side
		// and call it explicitly in defer if the other one is not called.
		if *down {
			defer xmlWrap("Deferred TearDown", deploy.Down)
		}
		// Start the cluster using this version.
		if err := xmlWrap("Up", deploy.Up); err != nil {
			return fmt.Errorf("starting e2e cluster: %s", err)
		}
		if *dump != "" {
			cmd := exec.Command("./cluster/kubectl.sh", "--match-server-version=false", "get", "nodes", "-oyaml")
			b, err := cmd.CombinedOutput()
			if *verbose {
				log.Printf("kubectl get nodes:\n%s", string(b))
			}
			if err == nil {
				if err := ioutil.WriteFile(filepath.Join(*dump, "nodes.yaml"), b, 0644); err != nil {
					errs = appendError(errs, fmt.Errorf("error writing nodes.yaml: %v", err))
				}
			} else {
				errs = appendError(errs, fmt.Errorf("error running get nodes: %v", err))
			}
		}
	}

	if *test {
		if err := xmlWrap("IsUp", deploy.IsUp); err != nil {
			errs = appendError(errs, err)
		} else {
			errs = appendError(errs, Test())
		}
	}

	if len(errs) > 0 && *dump != "" {
		errs = appendError(errs, xmlWrap("DumpClusterLogs", func() error {
			return DumpClusterLogs(*dump)
		}))
	}

	if *down {
		errs = appendError(errs, xmlWrap("TearDown", deploy.Down))
	}

	if len(errs) != 0 {
		return fmt.Errorf("encountered %d errors: %v", len(errs), errs)
	}
	return nil
}

func Build() error {
	// The build-release script needs stdin to ask the user whether
	// it's OK to download the docker image.
	cmd := exec.Command("make", "docker-build")
	cmd.Stdin = os.Stdin
	if err := finishRunning("build-release", cmd); err != nil {
		return fmt.Errorf("error building: %v", err)
	}
	return nil
}

type deployer interface {
	Up() error
	IsUp() error
	SetupKubecfg() error
	Down() error
}

func getDeployer() (deployer, error) {
	switch *deployment {
	case "bash":
		return bash{}, nil
	default:
		return nil, fmt.Errorf("unknown deployment strategy %q", *deployment)
	}
}

type bash struct{}

func (b bash) Up() error {
	return finishRunning("up", exec.Command("./hack/e2e-internal/e2e-up.sh"))
}

func (b bash) IsUp() error {
	return finishRunning("get status", exec.Command("./hack/e2e-internal/e2e-status.sh"))
}

func (b bash) SetupKubecfg() error {
	return nil
}

func (b bash) Down() error {
	return finishRunning("teardown", exec.Command("./hack/e2e-internal/e2e-down.sh"))
}

func DumpClusterLogs(location string) error {
	log.Printf("Dumping cluster logs to: %v", location)
	return finishRunning("dump cluster logs", exec.Command("./hack/e2e-internal/log-dump.sh", location))
}

func Test() error {
	if *testArgs == "" {
		*testArgs = "--ginkgo.focus=\\[Feature:Ingress\\]"
	}
	return finishRunning("Ginkgo tests", exec.Command("./hack/e2e-internal/ginkgo-e2e.sh", strings.Fields(*testArgs)...))
}

func finishRunning(stepName string, cmd *exec.Cmd) error {
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	log.Printf("Running: %v", stepName)
	defer func(start time.Time) {
		log.Printf("Step '%s' finished in %s", stepName, time.Since(start))
	}(time.Now())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %v: %v", stepName, err)
	}
	return nil
}
