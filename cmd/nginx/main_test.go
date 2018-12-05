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

package main

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress/controller"
)

func TestCreateApiserverClient(t *testing.T) {
	_, err := createApiserverClient("", "")
	if err == nil {
		t.Fatal("Expected an error creating REST client without an API server URL or kubeconfig file.")
	}
}

func TestHandleSigterm(t *testing.T) {
	clientSet := fake.NewSimpleClientset()

	ns := "test"

	cm := createConfigMap(clientSet, ns, t)
	defer deleteConfigMap(cm, ns, clientSet, t)

	name := "test"
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

	_, err := clientSet.CoreV1().Pods(ns).Create(&pod)
	if err != nil {
		t.Fatalf("error creating pod %v: %v", pod, err)
	}

	resetForTesting(func() { t.Fatal("bad parse") })

	os.Setenv("POD_NAME", name)
	os.Setenv("POD_NAMESPACE", ns)
	defer os.Setenv("POD_NAME", "")
	defer os.Setenv("POD_NAMESPACE", "")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--default-backend-service", "ingress-nginx/default-backend-http", "--http-port", "0", "--https-port", "0"}

	_, conf, err := parseFlags()
	if err != nil {
		t.Errorf("Unexpected error creating NGINX controller: %v", err)
	}
	conf.Client = clientSet

	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ngx := controller.NewNGINXController(conf, nil, fs)

	go handleSigterm(ngx, func(code int) {
		if code != 1 {
			t.Errorf("Expected exit code 1 but %d received", code)
		}

		return
	})

	time.Sleep(1 * time.Second)

	t.Logf("Sending SIGTERM to PID %d", syscall.Getpid())
	err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		t.Error("Unexpected error sending SIGTERM signal.")
	}

	err = clientSet.CoreV1().Pods(ns).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error deleting pod %v: %v", pod, err)
	}
}

func createConfigMap(clientSet kubernetes.Interface, ns string, t *testing.T) string {
	t.Helper()
	t.Log("Creating temporal config map")

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "config",
			SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
		},
	}

	cm, err := clientSet.CoreV1().ConfigMaps(ns).Create(configMap)
	if err != nil {
		t.Errorf("error creating the configuration map: %v", err)
	}
	t.Logf("Temporal configmap %v created", cm)

	return cm.Name
}

func deleteConfigMap(cm, ns string, clientSet kubernetes.Interface, t *testing.T) {
	t.Helper()
	t.Logf("Deleting temporal configmap %v", cm)

	err := clientSet.CoreV1().ConfigMaps(ns).Delete(cm, &metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("error deleting the configmap: %v", err)
	}
	t.Logf("Temporal configmap %v deleted", cm)
}
