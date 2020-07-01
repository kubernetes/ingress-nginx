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

package annotations

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("influxdb-*", func() {
	f := framework.NewDefaultFramework("influxdb")

	ginkgo.BeforeEach(func() {
		f.NewInfluxDBDeployment()
		f.NewEchoDeployment()
	})

	ginkgo.Context("when influxdb is enabled", func() {
		ginkgo.It("should send the request metric to the influxdb server", func() {
			ifs := createInfluxDBService(f)

			// Ingress configured with InfluxDB annotations
			host := "influxdb.e2e.local"
			createInfluxDBIngress(
				f,
				host,
				framework.EchoService,
				80,
				map[string]string{
					"nginx.ingress.kubernetes.io/enable-influxdb":      "true",
					"nginx.ingress.kubernetes.io/influxdb-host":        ifs.Spec.ClusterIP,
					"nginx.ingress.kubernetes.io/influxdb-port":        "8089",
					"nginx.ingress.kubernetes.io/influxdb-measurement": "requests",
					"nginx.ingress.kubernetes.io/influxdb-servername":  "e2e-nginx-srv",
				},
			)

			// Do a request to the echo server ingress that sends metrics
			// to the InfluxDB backend.
			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)

			framework.Sleep(10 * time.Second)

			var measurements string
			var err error

			err = wait.Poll(time.Second, time.Minute, func() (bool, error) {
				measurements, err = extractInfluxDBMeasurements(f)
				if err != nil {
					return false, nil
				}
				return true, nil
			})
			assert.Nil(ginkgo.GinkgoT(), err)

			var results map[string][]map[string]interface{}
			_ = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(measurements), &results)

			assert.NotEqual(ginkgo.GinkgoT(), len(measurements), 0)
			for _, elem := range results["results"] {
				assert.NotEqual(ginkgo.GinkgoT(), len(elem), 0)
			}
		})
	})
})

func createInfluxDBService(f *framework.Framework) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inflxudb",
			Namespace: f.Namespace,
		},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{
			{
				Name:       "udp",
				Port:       8089,
				TargetPort: intstr.FromInt(8089),
				Protocol:   "UDP",
			},
		},
			Selector: map[string]string{
				"app": "influxdb",
			},
		},
	}

	return f.EnsureService(service)
}

func createInfluxDBIngress(f *framework.Framework, host, service string, port int, annotations map[string]string) {
	ing := framework.NewSingleIngress(host, "/", host, f.Namespace, service, port, annotations)
	f.EnsureIngress(ing)

	f.WaitForNginxServer(host,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("server_name %v", host))
		})
}

func extractInfluxDBMeasurements(f *framework.Framework) (string, error) {
	l, err := f.KubeClientSet.CoreV1().Pods(f.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=influxdb",
	})
	if err != nil {
		return "", err
	}

	if len(l.Items) == 0 {
		return "", err
	}

	cmd := "influx -database 'nginx' -execute 'select * from requests' -format 'json' -pretty"

	var pod *corev1.Pod
	for _, p := range l.Items {
		pod = &p
		break
	}

	if pod == nil {
		return "", fmt.Errorf("no influxdb pods found")
	}

	o, err := execInfluxDBCommand(pod, cmd)
	if err != nil {
		return "", err
	}

	return o, nil
}

func execInfluxDBCommand(pod *corev1.Pod, command string) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v exec --namespace %s %s -- %s", framework.KubectlPath, pod.Namespace, pod.Name, command))
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()

	if execErr.Len() > 0 {
		return "", fmt.Errorf("stderr: %v", execErr.String())
	}

	if err != nil {
		return "", fmt.Errorf("could not execute '%s %s': %v", cmd.Path, cmd.Args, err)
	}

	return execOut.String(), nil
}
