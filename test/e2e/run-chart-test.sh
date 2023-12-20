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
    --name ${KIND_CLUSTER_NAME}
}

trap cleanup EXIT

export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-ingress-nginx-dev}

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Use 1.0.0-dev to make sure we use the latest configuration in the helm template
export TAG=1.0.0-dev
export ARCH=${ARCH:-amd64}
export REGISTRY=ingress-controller

BASEDIR=$(dirname "$0")
NGINX_BASE_IMAGE=$(cat $BASEDIR/../../NGINX_BASE)

echo "Running e2e with nginx base image ${NGINX_BASE_IMAGE}"

export NGINX_BASE_IMAGE=$NGINX_BASE_IMAGE

export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/kind-config-$KIND_CLUSTER_NAME}"

if [ "${SKIP_CLUSTER_CREATION:-false}" = "false" ]; then
  echo "[dev-env] creating Kubernetes cluster with kind"

  export K8S_VERSION=${K8S_VERSION:-v1.26.3@sha256:61b92f38dff6ccc29969e7aa154d34e38b89443af1a2c14e6cfbd2df6419c66f}

  kind create cluster \
    --verbosity=${KIND_LOG_LEVEL} \
    --name ${KIND_CLUSTER_NAME} \
    --config ${DIR}/kind.yaml \
    --retain \
    --image "kindest/node:${K8S_VERSION}"

  echo "Kubernetes cluster:"
  kubectl get nodes -o wide

fi

if [ "${SKIP_IMAGE_CREATION:-false}" = "false" ]; then
  if ! command -v ginkgo &> /dev/null; then
    go install github.com/onsi/ginkgo/v2/ginkgo@v2.13.1
  fi
  echo "[dev-env] building image"
  make -C ${DIR}/../../ clean-image build image
fi


KIND_WORKERS=$(kind get nodes --name="${KIND_CLUSTER_NAME}" | awk '{printf (NR>1?",":"") $1}')
echo "[dev-env] copying docker images to cluster..."

kind load docker-image --name="${KIND_CLUSTER_NAME}" --nodes=${KIND_WORKERS} ${REGISTRY}/controller:${TAG}

if [ "${SKIP_CERT_MANAGER_CREATION:-false}" = "false" ]; then
  curl -fsSL -o cmctl.tar.gz https://github.com/cert-manager/cert-manager/releases/download/v1.11.1/cmctl-linux-amd64.tar.gz
  tar xzf cmctl.tar.gz
  chmod +x cmctl
 ./cmctl help
  echo "[dev-env] apply cert-manager ..."
  kubectl apply --wait -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
  kubectl wait --timeout=30s --for=condition=available deployment/cert-manager -n cert-manager
  kubectl get validatingwebhookconfigurations cert-manager-webhook -ojson | jq '.webhooks[].clientConfig'
  kubectl get endpoints -n cert-manager cert-manager-webhook
  ./cmctl check api --wait=2m
fi

echo "[dev-env] running helm chart e2e tests..."
docker run --rm --interactive --network host \
    --name ct \
    --volume $KUBECONFIG:/root/.kube/config \
    --volume "${DIR}/../../":/workdir \
    --workdir /workdir \
    registry.k8s.io/ingress-nginx/e2e-test-runner:v20231208-4c39e6acc@sha256:0607184ca9c53c9c24a47b6f52347dd96137b05c6f276efa67051929a39e8f7a \
        ct install \
        --charts charts/ingress-nginx \
        --helm-extra-args "--timeout 60s"
