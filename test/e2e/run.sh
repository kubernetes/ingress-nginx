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

KIND_LOG_LEVEL="1"

if ! [ -z $DEBUG ]; then
  set -x
  KIND_LOG_LEVEL="6"
fi

set -o errexit
set -o nounset
set -o pipefail

cleanup() {
  if [[ "${KUBETEST_IN_DOCKER:-}" == "true" ]]; then
    kind "export" logs --name ${KIND_CLUSTER_NAME} "${ARTIFACTS}/logs" || true
  fi

  kind delete cluster \
    --verbosity=${KIND_LOG_LEVEL} \
    --name ${KIND_CLUSTER_NAME}
}

trap cleanup EXIT

if ! command -v parallel &> /dev/null; then
  if [[ "$OSTYPE" == "linux-gnu" ]]; then
    echo "Parallel is not installed. Use the package manager to install it"
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Parallel is not installed. Install it running brew install parallel"
  fi

  exit 1
fi

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Use 1.0.0-dev to make sure we use the latest configuration in the helm template
export TAG=1.0.0-dev
export ARCH=amd64
export REGISTRY=ingress-controller

export K8S_VERSION=${K8S_VERSION:-v1.17.2@sha256:59df31fc61d1da5f46e8a61ef612fa53d3f9140f82419d1ef1a6b9656c6b737c}

export DOCKER_CLI_EXPERIMENTAL=enabled

KIND_CLUSTER_NAME="ingress-nginx-dev"

echo "[dev-env] creating Kubernetes cluster with kind"

export KUBECONFIG="${HOME}/.kube/kind-config-${KIND_CLUSTER_NAME}"
kind create cluster \
  --verbosity=${KIND_LOG_LEVEL} \
  --name ${KIND_CLUSTER_NAME} \
  --config ${DIR}/kind.yaml \
  --retain \
  --image "kindest/node:${K8S_VERSION}"

echo "Kubernetes cluster:"
kubectl get nodes -o wide

echo "[dev-env] building container"
export EXIT_CODE=-1
echo "
make -C ${DIR}/../../ build container
make -C ${DIR}/../../ e2e-test-image
make -C ${DIR}/../../images/fastcgi-helloserver/ GO111MODULE=\"on\" build container
make -C ${DIR}/../../images/echo/ container
make -C ${DIR}/../../images/httpbin/ container
" | parallel --joblog /tmp/log {} || EXIT_CODE=$?
if [ ${EXIT_CODE} -eq 0 ] || [ ${EXIT_CODE} -eq -1 ]; 
then
  echo "Image builds were ok! Log:"
  cat /tmp/log
  unset EXIT_CODE
else
  echo "Image builds were not ok! Log:"
  cat /tmp/log
  exit
fi

# Remove after https://github.com/kubernetes/ingress-nginx/pull/4271 is merged
docker tag ${REGISTRY}/nginx-ingress-controller-${ARCH}:${TAG} ${REGISTRY}/nginx-ingress-controller:${TAG}

# Preload images used in e2e tests
docker pull openresty/openresty:1.15.8.2-alpine
docker pull moul/grpcbin

echo "[dev-env] copying docker images to cluster..."
export EXIT_CODE=-1
echo "
kind load docker-image --name="${KIND_CLUSTER_NAME}" nginx-ingress-controller:e2e
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/nginx-ingress-controller:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/fastcgi-helloserver:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" openresty/openresty:1.15.8.2-alpine
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/httpbin:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" ${REGISTRY}/echo:${TAG}
kind load docker-image --name="${KIND_CLUSTER_NAME}" moul/grpcbin
" | parallel --joblog /tmp/log {} || EXIT_CODE=$?
if [ ${EXIT_CODE} -eq 0 ] || [ ${EXIT_CODE} -eq -1 ]; 
then
  echo "Image loads were ok! Log:"
  cat /tmp/log
  unset EXIT_CODE
else
  echo "Image loads were not ok! Log:"
  cat /tmp/log
  exit
fi

echo "[dev-env] running e2e tests..."
make -C ${DIR}/../../ e2e-test
