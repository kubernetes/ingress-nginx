#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
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

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export TAG=dev
export ARCH=amd64
export REGISTRY=${REGISTRY:-ingress-controller}

KIND_CLUSTER_NAME="ingress-nginx-dev"

kind --version || $(echo "Please install kind before running e2e tests";exit 1)

SKIP_CLUSTER_CREATION=${SKIP_CLUSTER_CREATION:-}
if [ -z "${SKIP_CLUSTER_CREATION}" ]; then
    echo "[dev-env] creating Kubernetes cluster with kind"
    kind create cluster --name ${KIND_CLUSTER_NAME} --config ${DIR}/kind.yaml
fi

export KUBECONFIG="$(kind get kubeconfig-path --name="${KIND_CLUSTER_NAME}")"

sleep 60

echo "Kubernetes cluster:"
kubectl get nodes -o wide

echo "[dev-env] installing kubectl"
kubectl version || $(echo "Please install kubectl before running e2e tests";exit 1)

kubectl config set-context kubernetes-admin@${KIND_CLUSTER_NAME}

echo "[dev-env] building container"
make -C ${DIR}/../../ build container
make -C ${DIR}/../../ e2e-test-image

echo "copying docker images to cluster..."
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/nginx-ingress-controller:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" nginx-ingress-controller:e2e

make -C ${DIR}/../../ e2e-test
