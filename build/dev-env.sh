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

NAMESPACE="${NAMESPACE:-ingress-nginx}"
echo "NAMESPACE is set to ${NAMESPACE}"

kubectl config use-context minikube

export TAG=dev
export ARCH=amd64
export REGISTRY=${REGISTRY:-ingress-controller}

DEV_IMAGE=${REGISTRY}/nginx-ingress-controller:${TAG}

{ [ "$(minikube status | grep -c Running)" -ge 2 ] && minikube status | grep -qE ': Configured$|Correctly Configured'; } \
  || minikube start \
    --extra-config=kubelet.sync-frequency=1s \
    --extra-config=apiserver.authorization-mode=RBAC \
    --kubernetes-version=v1.15.0

# shellcheck disable=SC2046
eval $(minikube docker-env --shell bash)

echo "[dev-env] building container"
make build container
docker tag "${REGISTRY}/nginx-ingress-controller-${ARCH}:${TAG}" "${DEV_IMAGE}"

# kubectl >= 1.14 includes Kustomize via "apply -k". Makes it easier to use on Linux as well, assuming kubectl installed
KUBE_CLIENT_VERSION=$(kubectl version --client --short | awk '{print $3}' | cut -d. -f2) || true
if [[ ${KUBE_CLIENT_VERSION} -lt 14 ]]; then
  for tool in kubectl kustomize; do
	echo "[dev-env] installing $tool"
	$tool version || brew install $tool
  done
fi

if ! kubectl get namespace "${NAMESPACE}"; then
  kubectl create namespace "${NAMESPACE}"
fi

kubectl get deploy nginx-ingress-controller -n "${NAMESPACE}" && kubectl delete deploy nginx-ingress-controller -n "${NAMESPACE}"

ROOT=./deploy/minikube

if [[ ${KUBE_CLIENT_VERSION} -lt 14 ]]; then
  pushd $ROOT
  kustomize edit set namespace "${NAMESPACE}"
  kustomize edit set image "quay.io/kubernetes-ingress-controller/nginx-ingress-controller=${DEV_IMAGE}"
  popd

  echo "[dev-env] deploying NGINX Ingress controller in namespace $NAMESPACE"
  kustomize build $ROOT | kubectl apply -f -
else
  sed -i -e "s|^namespace: .*|namespace: ${NAMESPACE}|g" "${ROOT}/kustomization.yaml"

  echo "[dev-env] deploying NGINX Ingress controller in namespace $NAMESPACE"
  kubectl apply -k "${ROOT}"
fi
