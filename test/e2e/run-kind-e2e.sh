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

set -o errexit
set -o nounset
set -o pipefail

cleanup() {
  if [[ "${KUBETEST_IN_DOCKER:-}" == "true" ]]; then
    kind "export" logs --name ${KIND_CLUSTER_NAME} "${ARTIFACTS}/logs" || true
  fi

  kind delete cluster \
    --verbosity="${KIND_LOG_LEVEL}" \
    --name "${KIND_CLUSTER_NAME}"
}

DEBUG=${DEBUG:=false}

if [ "${DEBUG}" = "true" ]; then
  set -x
  KIND_LOG_LEVEL="6"
else
  trap cleanup EXIT
fi

KIND_LOG_LEVEL="1"
IS_CHROOT="${IS_CHROOT:-false}"
export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-ingress-nginx-dev}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Use 1.0.0-dev to make sure we use the latest configuration in the helm template
export TAG=1.0.0-dev
export ARCH=${ARCH:-amd64}
export REGISTRY=ingress-controller
NGINX_BASE_IMAGE=${NGINX_BASE_IMAGE:-$(cat "$DIR"/../../NGINX_BASE)}
export NGINX_BASE_IMAGE=$NGINX_BASE_IMAGE
export DOCKER_CLI_EXPERIMENTAL=enabled
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/kind-config-$KIND_CLUSTER_NAME}"
SKIP_INGRESS_IMAGE_CREATION="${SKIP_INGRESS_IMAGE_CREATION:-false}"
SKIP_E2E_IMAGE_CREATION="${SKIP_E2E_IMAGE_CREATION:=false}"
SKIP_CLUSTER_CREATION="${SKIP_CLUSTER_CREATION:-false}"

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

echo "Running e2e with nginx base image ${NGINX_BASE_IMAGE}"

if [ "${SKIP_CLUSTER_CREATION}" = "false" ]; then
  echo "[dev-env] creating Kubernetes cluster with kind"

  export K8S_VERSION=${K8S_VERSION:-v1.35.0@sha256:4613778f3cfcd10e615029370f5786704559103cf27bef934597ba562b269661}

  # delete the cluster if it exists
  if kind get clusters | grep "${KIND_CLUSTER_NAME}"; then
    kind delete cluster --name "${KIND_CLUSTER_NAME}"
  fi

  kind create cluster \
    --verbosity="${KIND_LOG_LEVEL}" \
    --name "${KIND_CLUSTER_NAME}" \
    --config "${DIR}"/kind.yaml \
    --retain \
    --image "kindest/node:${K8S_VERSION}"

  echo "Kubernetes cluster:"
  kubectl get nodes -o wide
fi

if [ "${SKIP_INGRESS_IMAGE_CREATION}" = "false" ]; then
  echo "[dev-env] building image"
  if [ "${IS_CHROOT}" = "true" ]; then
    make BASE_IMAGE="${NGINX_BASE_IMAGE}" -C "${DIR}"/../../ clean-image build image-chroot
    docker tag ${REGISTRY}/controller-chroot:${TAG} ${REGISTRY}/controller:${TAG}
  else
    make BASE_IMAGE="${NGINX_BASE_IMAGE}" -C "${DIR}"/../../ clean-image build image
  fi

  echo "[dev-env] .. done building controller images"
fi

if [ "${SKIP_E2E_IMAGE_CREATION}" = "false" ]; then
  if ! command -v ginkgo &> /dev/null; then
    go install github.com/onsi/ginkgo/v2/ginkgo@v2.27.3
  fi

  echo "[dev-env] .. done building controller images"
  echo "[dev-env] now building e2e-image.."
  make -C "${DIR}"/../e2e-image image
  echo "[dev-env] ..done building e2e-image"
fi

# Preload images used in e2e tests
KIND_WORKERS=$(kind get nodes --name="${KIND_CLUSTER_NAME}" | grep worker | awk '{printf (NR>1?",":"") $1}')

echo "[dev-env] copying docker images to cluster..."

kind load docker-image --name="${KIND_CLUSTER_NAME}" --nodes="${KIND_WORKERS}" nginx-ingress-controller:e2e
kind load docker-image --name="${KIND_CLUSTER_NAME}" --nodes="${KIND_WORKERS}" "${REGISTRY}"/controller:"${TAG}"
echo "[dev-env] running e2e tests..."
make -C "${DIR}"/../../ e2e-test
