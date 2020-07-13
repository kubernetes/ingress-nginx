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
  echo "Please install kubectl 1.15 or higher"
  exit 1
fi

if ! command -v helm &> /dev/null; then
  echo "Please install helm"
  exit 1
fi

HELM_VERSION=$(helm version 2>&1 | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+') || true
if [[ ${HELM_VERSION} < "v3.0.0" ]]; then
  echo "Please upgrade helm to v3.0.0 or higher"
  exit 1
fi

KUBE_CLIENT_VERSION=$(kubectl version --client --short | awk '{print $3}' | cut -d. -f2) || true
if [[ ${KUBE_CLIENT_VERSION} -lt 14 ]]; then
  echo "Please update kubectl to 1.15 or higher"
  exit 1
fi

echo "[dev-env] building image"
make build image
docker tag "${REGISTRY}/controller:${TAG}" "${DEV_IMAGE}"

export K8S_VERSION=${K8S_VERSION:-v1.18.4@sha256:d8ff5fc405fc679dd3dd0cccc01543ba4942ed90823817d2e9e2c474a5343c4f}

KIND_CLUSTER_NAME="ingress-nginx-dev"

if ! kind get clusters -q | grep -q ${KIND_CLUSTER_NAME}; then
echo "[dev-env] creating Kubernetes cluster with kind"
cat <<EOF | kind create cluster --name ${KIND_CLUSTER_NAME} --image "kindest/node:${K8S_VERSION}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
        authorization-mode: "AlwaysAllow"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
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
