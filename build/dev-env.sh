#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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

if [ -n "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

DIR=$(cd $(dirname "${BASH_SOURCE}") && pwd -P)

export TAG=1.0.0-dev
export REGISTRY=${REGISTRY:-ingress-controller}

DEV_IMAGE=${REGISTRY}/controller:${TAG}

if ! command -v kind &> /dev/null; then
  echo "kind is not installed"
  echo "Use a package manager (i.e 'brew install kind') or visit the official site https://kind.sigs.k8s.io"
  exit 1
fi

if ! command -v kubectl &> /dev/null; then
  echo "Please install kubectl 1.24.0 or higher"
  exit 1
fi

if ! command -v helm &> /dev/null; then
  echo "Please install helm"
  exit 1
fi

function ver { printf "%d%03d%03d" $(echo "$1" | tr '.' ' '); }

HELM_VERSION=$(helm version 2>&1 | cut -f1 -d"," | grep -oE '[0-9]+\.[0-9]+\.[0-9]+') || true
echo $HELM_VERSION
if [[ $(ver $HELM_VERSION) -lt $(ver "3.10.0") ]]; then
  echo "Please upgrade helm to v3.10.0 or higher"
  exit 1
fi

KUBE_CLIENT_VERSION=$(kubectl version --client -oyaml 2>/dev/null | grep "minor:" | awk '{print $2}' | tr -d '"') || true
if [[ ${KUBE_CLIENT_VERSION} -lt 24 ]]; then
  echo "Please update kubectl to 1.24.2 or higher"
  exit 1
fi

echo "[dev-env] building image"
make build image
docker tag "${REGISTRY}/controller:${TAG}" "${DEV_IMAGE}"

export K8S_VERSION=${K8S_VERSION:-v1.34.2@sha256:745f8ed46d8e99517774768227fd1a0af34a6bf395aef9c7ed98fbce0e263918}

KIND_CLUSTER_NAME="ingress-nginx-dev"

if ! kind get clusters -q | grep -q ${KIND_CLUSTER_NAME}; then
  echo "[dev-env] creating Kubernetes cluster with kind"
  kind create cluster --name ${KIND_CLUSTER_NAME} --image "kindest/node:${K8S_VERSION}" --config ${DIR}/kind.yaml
else
  echo "[dev-env] using existing Kubernetes kind cluster"
fi

echo "[dev-env] copying docker images to cluster..."
kind load docker-image --name="${KIND_CLUSTER_NAME}" "${DEV_IMAGE}"

echo "[dev-env] deploying NGINX Ingress controller..."
kubectl create namespace ingress-nginx &> /dev/null || true

cat << EOF | helm template ingress-nginx ${DIR}/../charts/ingress-nginx --namespace=ingress-nginx --values - | kubectl apply -n ingress-nginx -f -
controller:
  image:
    repository: ${REGISTRY}/controller
    tag: ${TAG}
    digest:
  config:
    worker-processes: "1"
  podLabels:
    deploy-date: "$(date +%s)"
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  hostPort:
    enabled: true
  terminationGracePeriodSeconds: 0
  service:
    type: NodePort
EOF

cat <<EOF

Kubernetes cluster ready and ingress-nginx listening in localhost using ports 80 and 443

To delete the dev cluster execute: 'kind delete cluster --name ingress-nginx-dev'

EOF
