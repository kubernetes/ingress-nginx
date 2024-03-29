/*
Copyright 2024 The Kubernetes Authors.

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

package steps

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Kind mg.Namespace

var KIND_CONFIG_FILE string = "test/e2e/kind.yaml"
var KIND_CLUSTER_NAME string = "ingress-nginx-dev"
var K8_VERSION string = "kindest/node:v1.29.2@sha256:51a1434a5397193442f0be2a297b488b6c919ce8a3931be0ce822606ea5ca245"
var KIND_LOG_LEVEL string = "6"
var GATEWAY_API_VERSION string = "v1.0.0"
var CRD_CHANNEL string = "standard"
var GWAPI_CRD_BASE_URL string = "https://raw.githubusercontent.com/kubernetes-sigs/gateway-api"

// Creates a new cluster mage kind:createcluster name k8sversion
func (Kind) CreateCluster(clusterName, k8sVersion string) error {

	if len(clusterName) > 0 {
		KIND_CLUSTER_NAME = clusterName
	}

	if len(k8sVersion) > 0 {
		K8_VERSION = fmt.Sprintf("kindest/node:%s", k8sVersion)
	}

	cluster, err := sh.Output("kind", "get", "clusters")
	if err != nil {
		return err
	}
	if strings.Contains(cluster, KIND_CLUSTER_NAME) {
		err := sh.Run("kind", "delete", "cluster", "--name", KIND_CLUSTER_NAME)
		if err != nil {
			return err
		}
	}

	err = sh.RunV("kind", "create", "cluster",
		"--verbosity", KIND_LOG_LEVEL,
		"--name", KIND_CLUSTER_NAME,
		"--config", KIND_CONFIG_FILE,
		"--retain",
		"--image", K8_VERSION,
	)
	if err != nil {
		return err
	}
	return nil
}

// InstallGatewayCRD Install Gateway API CRDS for testing
func (Kind) InstallGatewayCRD() error {
	var k = sh.RunCmd("kubectl", "apply", "-f")

	gatewayClass := fmt.Sprintf("%s/%s/config/crd/%s/gateway.networking.k8s.io_gatewayclasses.yaml", GWAPI_CRD_BASE_URL, GATEWAY_API_VERSION, CRD_CHANNEL)
	gateway := fmt.Sprintf("%s/%s/config/crd/%s/gateway.networking.k8s.io_gateways.yaml", GWAPI_CRD_BASE_URL, GATEWAY_API_VERSION, CRD_CHANNEL)
	httpRoutes := fmt.Sprintf("%s/%s/config/crd/%s/gateway.networking.k8s.io_httproutes.yaml", GWAPI_CRD_BASE_URL, GATEWAY_API_VERSION, CRD_CHANNEL)

	if err := k(gatewayClass); err != nil {
		return err
	}
	if err := k(gateway); err != nil {
		return err
	}
	if err := k(httpRoutes); err != nil {
		return err
	}

	err := sh.RunV("kubectl", "get", "crds")
	if err != nil {
		return err
	}
	return nil
}

func getWorkers() (string, error) {
	return sh.Output("kind", "get", "nodes", fmt.Sprintf("--name=%s", KIND_CLUSTER_NAME))
}

func (Kind) InstallMetalLB() error {

	err := sh.RunV("kubectl", "apply", "-f", "https://raw.githubusercontent.com/metallb/metallb/v0.14.4/config/manifests/metallb-native.yaml")
	if err != nil {
		return err
	}
	err = sh.RunV("kubectl", "wait", "--namespace", "metallb-system",
		"--for=condition=ready", "pod",
		"--selector=app=metallb",
		"--timeout=90s")
	if err != nil {
		return err
	}

	err = sh.RunV("kubectl", "apply", "-f", "test/metallb-config.yaml")
	if err != nil {
		return err
	}
	return nil
}
