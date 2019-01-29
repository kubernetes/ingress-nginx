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

SKIP_MINIKUBE_START=${SKIP_MINIKUBE_START:-}
NAMESPACE="${NAMESPACE:-ingress-nginx}"
echo "NAMESPACE is set to ${NAMESPACE}"

kubectl config set-context minikube

export TAG=dev
export ARCH=amd64
export REGISTRY=${REGISTRY:-ingress-controller}

DEV_IMAGE=${REGISTRY}/nginx-ingress-controller:${TAG}

if [ -z "${SKIP_MINIKUBE_START}" ]; then
    test $(minikube status | grep Running | wc -l) -eq 2 && $(minikube status | grep -q 'Correctly Configured') || minikube start \
        --extra-config=kubelet.sync-frequency=1s \
        --extra-config=apiserver.authorization-mode=RBAC

    eval $(minikube docker-env --shell bash)
fi

echo "[dev-env] building container"
make build container

docker save "${DEV_IMAGE}" | (eval $(minikube docker-env --shell bash) && docker load) || true

echo "[dev-env] installing kubectl"
kubectl version || brew install kubectl

echo "[dev-env] deploying NGINX Ingress controller in namespace $NAMESPACE"
cat ./deploy/mandatory.yaml                            | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/provider/baremetal/service-nodeport.yaml  | kubectl apply --namespace=$NAMESPACE -f -

echo "updating image..."
kubectl set image \
    deployments \
    --namespace ingress-nginx \
    --selector app.kubernetes.io/name=ingress-nginx \
    nginx-ingress-controller=${DEV_IMAGE}
