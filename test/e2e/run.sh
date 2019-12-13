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

KIND_LOG_LEVEL="0"

if ! [ -z $DEBUG ]; then
  set -x
  KIND_LOG_LEVEL="6"
fi

set -o errexit
set -o nounset
set -o pipefail

if ! command -v parallel &> /dev/null; then
  if [[ "$OSTYPE" == "linux-gnu" ]]; then
    echo "Parallel is not installed. Use the package manager to install it"
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Parallel is not installed. Install it running brew install parallel"
  fi

  exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export TAG=dev
export ARCH=amd64
export REGISTRY=ingress-controller

export K8S_VERSION=${K8S_VERSION:-v1.17.0}

KIND_CLUSTER_NAME="ingress-nginx-dev"

kind --version || $(echo "Please install kind before running e2e tests";exit 1)

echo "[dev-env] creating Kubernetes cluster with kind"

export KUBECONFIG="${HOME}/.kube/kind-config-${KIND_CLUSTER_NAME}"
kind create cluster \
  --verbosity=${KIND_LOG_LEVEL} \
  --name ${KIND_CLUSTER_NAME} \
  --config ${DIR}/kind.yaml \
  --image "kindest/node:${K8S_VERSION}"

echo "Kubernetes cluster:"
kubectl get nodes -o wide

echo "[dev-env] building container"
echo "
make -C ${DIR}/../../ build container
make -C ${DIR}/../../ e2e-test-image
make -C ${DIR}/../../images/fastcgi-helloserver/ build container
make -C ${DIR}/../../images/httpbin/ container
" | parallel --progress --joblog /tmp/log {} || cat /tmp/log

# Remove after https://github.com/kubernetes/ingress-nginx/pull/4271 is merged
docker tag ${REGISTRY}/nginx-ingress-controller-${ARCH}:${TAG} ${REGISTRY}/nginx-ingress-controller:${TAG}

# Preload images used in e2e tests
docker pull openresty/openresty:1.15.8.2-alpine

echo "[dev-env] copying docker images to cluster..."
echo "
kind load docker-image --name="${KIND_CLUSTER_NAME}" nginx-ingress-controller:e2e
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/nginx-ingress-controller:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/fastcgi-helloserver:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" openresty/openresty:1.15.8.2-alpine
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/httpbin:${TAG}
" | parallel --progress --joblog /tmp/log {} || cat /tmp/log

echo "[dev-env] running e2e tests..."
make -C ${DIR}/../../ e2e-test

kind delete cluster \
  --verbosity=${KIND_LOG_LEVEL} \
  --name ${KIND_CLUSTER_NAME}
