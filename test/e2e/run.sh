#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

KIND_LOG_LEVEL="info"

if ! [ -z $DEBUG ]; then
  set -x
  KIND_LOG_LEVEL="debug"
fi

set -o errexit
set -o nounset
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export TAG=dev
export ARCH=amd64
export REGISTRY=ingress-controller

export K8S_VERSION=${K8S_VERSION:-v1.14.1}

KIND_CLUSTER_NAME="ingress-nginx-dev"

kind --version || $(echo "Please install kind before running e2e tests";exit 1)

echo "[dev-env] creating Kubernetes cluster with kind"
# TODO: replace the custom images after https://github.com/kubernetes-sigs/kind/issues/531
kind create cluster \
  --loglevel=${KIND_LOG_LEVEL} \
  --name ${KIND_CLUSTER_NAME} \
  --config ${DIR}/kind.yaml \
  --image "aledbf/kind-node:${K8S_VERSION}"

export KUBECONFIG="$(kind get kubeconfig-path --name="${KIND_CLUSTER_NAME}")"

echo "Kubernetes cluster:"
kubectl get nodes -o wide

kubectl config set-context kubernetes-admin@${KIND_CLUSTER_NAME}

echo "[dev-env] building container"
make -C ${DIR}/../../ build container
make -C ${DIR}/../../ e2e-test-image

echo "[dev-env] copying docker images to cluster..."
kind load docker-image --name="${KIND_CLUSTER_NAME}" nginx-ingress-controller:e2e
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/nginx-ingress-controller:${TAG}

echo "[dev-env] running e2e tests..."
make -C ${DIR}/../../ e2e-test

kind delete cluster \
  --loglevel=${KIND_LOG_LEVEL} \
  --name ${KIND_CLUSTER_NAME}
