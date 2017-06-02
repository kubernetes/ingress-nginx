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
limitations
*/

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestIngressSuite(t *testing.T) {
	pwd, _ := os.Getwd()
	filepath.Walk(path.Join(pwd, "/suite"),
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".yaml") {
				t.Log(path)
				runTestCase(path, t)
			}
			return nil
		})
}

func runTestCase(rawtc string, t *testing.T) {
	tc, err := parseTestCase(rawtc)
	if err != nil {
		t.Fatalf("unexpected error reading Ingress test case file %v: %v", rawtc, err)
	}

	t.Run(tc.Name, func(t *testing.T) {
		t.Logf("starting deploy of requirements for test case '%v'", tc.Name)
		err := tc.deploy()
		if err != nil {
			t.Fatalf("unexpected error in test case deploy process: %v", err)
		}

		if tc.Assert == nil {
			t.Fatalf("test case %v does not contains tests", tc.Name)
		}

		for _, assert := range tc.Assert {
			t.Logf("running assert %v", assert.Name)
		}

		err = tc.undeploy()
		if err != nil {
			t.Fatalf("unexpected error in test case deploy process: %v", err)
		}
	})
}

// parseTestCase parses a test case from a yaml file
func parseTestCase(p string) (IngressTestCase, error) {
	return IngressTestCase{
		Name: "basic test",
	}, nil
}

// deploy creates the kubernetes object specified in the test case
func (tc IngressTestCase) deploy() error {
	if tc.Pod != nil {
		return nil
	}

	if tc.ReplicationController != nil {
		return nil
	}

	if tc.Deployment != nil {
		return nil
	}

	return fmt.Errorf("invalid deployment option. Please check the test case")
}

// undeploy removes the kubernetes object created by the test case
func (tc IngressTestCase) undeploy() error {
	if tc.Pod != nil {
		return nil
	}

	if tc.ReplicationController != nil {
		return nil
	}

	if tc.Deployment != nil {
		return nil
	}

	return nil
}

func deployIngressController() error {
	return nil
}
