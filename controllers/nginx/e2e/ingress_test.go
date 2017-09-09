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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/ingress/controllers/nginx/e2e/util"
)

func TestIngressSuite(t *testing.T) {
	client, err := util.GetClient()
	if err != nil {
		t.Fatalf("unexpected error creating k8s client: %v", err)
	}

	pwd, _ := os.Getwd()
	filepath.Walk(path.Join(pwd, "/suite"),
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".yaml") {
				t.Log(path)
				runTestCase(path, client, t)
			}
			return nil
		})
}

func runTestCase(rawtc string, client kubernetes.Interface, t *testing.T) {
	tc, err := parseTestCase(rawtc)
	if err != nil {
		t.Fatalf("unexpected error reading Ingress test case file %v: %v", rawtc, err)
	}

	t.Run(tc.Name, func(t *testing.T) {
		if len(tc.Assert) == 0 {
			t.Fatal("test case does not contains tests")
		}

		if tc.Ingress == nil {
			t.Fatal("the test case does not contains an Ingress rule")
		}

		t.Logf("starting deploy of requirements for test case '%v'", tc.Name)
		err := tc.deploy(client)
		if err != nil {
			t.Fatalf("unexpected error in test case deploy process: %v", err)
		}

		for _, assert := range tc.Assert {
			t.Logf("running assert %v", assert.Name)
		}

		err = tc.undeploy(client)
		if err != nil {
			t.Fatalf("unexpected error in test case deploy process: %v", err)
		}
	})
}

// parseTestCase parses a test case from a yaml file
func parseTestCase(p string) (*IngressTestCase, error) {
	file, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var itc IngressTestCase
	err = yaml.Unmarshal(file, &itc)
	if err != nil {
		return nil, err
	}

	return &itc, nil
}

// deploy creates the kubernetes object specified in the test case
func (tc IngressTestCase) deploy(client kubernetes.Interface) error {
	_, err := client.Extensions().Ingresses(getNamespace(tc.Ingress.Namespace)).Create(tc.Ingress)
	if err != nil {
		return err
	}

	if tc.Service != nil {
		_, err := client.CoreV1().Services(getNamespace(tc.Service.Namespace)).Create(tc.Service)
		if err != nil {
			return err
		}
	}

	if tc.ReplicationController != nil {
		_, err := client.CoreV1().ReplicationControllers(getNamespace(tc.ReplicationController.Namespace)).Create(tc.ReplicationController)
		return err
	} else if tc.Deployment != nil {
		_, err := client.Extensions().Deployments(getNamespace(tc.Deployment.Namespace)).Create(tc.Deployment)
		return err
	}

	return fmt.Errorf("invalid deployment option. Please check the test case")
}

// undeploy removes the kubernetes object created by the test case
func (tc IngressTestCase) undeploy(client kubernetes.Interface) error {
	err := client.Extensions().Ingresses(getNamespace(tc.Ingress.Namespace)).Delete(tc.Ingress.Name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	if tc.Service != nil {
		err := client.CoreV1().Services(getNamespace(tc.Service.Namespace)).Delete(tc.Service.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	if tc.ReplicationController != nil {
		err := client.CoreV1().ReplicationControllers(getNamespace(tc.ReplicationController.Namespace)).Delete(tc.ReplicationController.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	} else if tc.Deployment != nil {
		err := client.Extensions().Deployments(getNamespace(tc.Deployment.Namespace)).Delete(tc.Deployment.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func getNamespace(ns string) string {
	if ns == "" {
		return "default"
	}

	return ns
}
