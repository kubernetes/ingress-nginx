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

: "${NAMESPACE:=ingress-nginx}"
echo "NAMESPACE is set to ${NAMESPACE}"

test $(minikube status | grep Running | wc -l) -eq 2 && $(minikube status | grep -q 'Correctly Configured') || minikube start
eval $(minikube docker-env)

export TAG=dev
export REGISTRY=ingress-controller

echo "[dev-env] building container"
ARCH=amd64 make build container

echo "[dev-env] installing kubectl"
kubectl version || brew install kubectl

echo "[dev-env] deploying NGINX Ingress controller in namespace $NAMESPACE"
cat ./deploy/namespace.yaml                  | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/default-backend.yaml            | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/configmap.yaml                  | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/tcp-services-configmap.yaml     | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/udp-services-configmap.yaml     | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/rbac.yaml                       | kubectl apply --namespace=$NAMESPACE -f -
cat ./deploy/with-rbac.yaml                  | kubectl apply --namespace=$NAMESPACE -f -

echo "updating image..."
kubectl set image \
    deployments \
    --namespace ingress-nginx \
    --selector app=ingress-nginx \
    nginx-ingress-controller=${REGISTRY}/nginx-ingress-controller:${TAG}
