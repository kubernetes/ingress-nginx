#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
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
    --name ${KIND_CLUSTER_NAME}
}

trap cleanup EXIT

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export K8S_VERSION=${K8S_VERSION:-v1.18.2@sha256:7b27a6d0f2517ff88ba444025beae41491b016bc6af573ba467b70c5e8e0d85f}

KIND_CLUSTER_NAME="ingress-nginx-dev"

echo "[dev-env] creating Kubernetes cluster with kind"
export KUBECONFIG="${HOME}/.kube/kind-config-${KIND_CLUSTER_NAME}"
kind create cluster \
  --name ${KIND_CLUSTER_NAME} \
  --config "${DIR}/kind.yaml" \
  --retain \
  --image "kindest/node:${K8S_VERSION}"

echo "Kubernetes cluster:"
kubectl get nodes -o wide

echo "[dev-env] running helm chart e2e tests..."
docker run --rm --interactive --network host \
    --name ct \
    --volume $KUBECONFIG:/root/.kube/config \
    --volume "${DIR}/../../":/workdir \
    --workdir /workdir \
    quay.io/helmpack/chart-testing:v3.0.0-rc.1 ct install \
        --charts charts/ingress-nginx \
        --helm-extra-args "--timeout 120s"
