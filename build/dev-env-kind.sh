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

set -o errexit
set -o nounset
set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

NAMESPACE="${NAMESPACE:-ingress-nginx}"
echo "NAMESPACE is set to ${NAMESPACE}"

export TAG=dev
export ARCH=amd64
export REGISTRY=${REGISTRY:-ingress-controller}

DEV_IMAGE=${REGISTRY}/nginx-ingress-controller:${TAG}

KIND_CLUSTER_NAME="ingress-nginx-dev"

echo "[dev-env] checking for kind binary"
kind --version || go get -u -v sigs.k8s.io/kind

SKIP_CLUSTER_CREATION=${SKIP_CLUSTER_CREATION:-}
if [ -z "${SKIP_CLUSTER_CREATION}" ]; then
    echo "[dev-env] creating Kubernetes cluster with kind"
    kind create cluster --name ${KIND_CLUSTER_NAME} --config ${DIR}/kind.yaml
fi

export KUBECONFIG="$(kind get kubeconfig-path --name="${KIND_CLUSTER_NAME}")"

echo "[dev-env] building container"
make -C ${DIR}/.. build container

TMP_DIR=$(mktemp -d)
BUNDLE_FILE="${TMP_DIR}"/cmbundle.tar.gz

docker save "${DEV_IMAGE}" -o ${BUNDLE_FILE}

CONTAINERS="${KIND_CLUSTER_NAME}-control-plane ${KIND_CLUSTER_NAME}-worker1 ${KIND_CLUSTER_NAME}-worker2"

for TARGET in $CONTAINERS;do
    docker cp "${BUNDLE_FILE}" "${TARGET}":/bundle.tar.gz
    docker exec "${TARGET}" docker load -i /bundle.tar.gz
done

echo "[dev-env] installing kubectl"
kubectl version || brew install kubectl

kubectl config set-context kubernetes-admin@${KIND_CLUSTER_NAME}

echo "[dev-env] deploying NGINX Ingress controller in namespace $NAMESPACE"
cat ${DIR}/../deploy/mandatory.yaml                            | kubectl apply --namespace=$NAMESPACE -f -
cat ${DIR}/../deploy/provider/baremetal/service-nodeport.yaml  | kubectl apply --namespace=$NAMESPACE -f -

echo "updating image..."
kubectl set image \
    deployments \
    --namespace ingress-nginx \
    --selector app.kubernetes.io/name=ingress-nginx \
    nginx-ingress-controller=${DEV_IMAGE}
