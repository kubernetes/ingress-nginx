#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
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

export JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'

echo "deploying NGINX Ingress controller"
cat deploy/namespace.yaml | kubectl apply -f -
cat deploy/default-backend.yaml | kubectl apply -f -
cat deploy/configmap.yaml | kubectl apply -f -
cat deploy/tcp-services-configmap.yaml | kubectl apply -f -
cat deploy/udp-services-configmap.yaml | kubectl apply -f -
cat deploy/without-rbac.yaml | kubectl apply -f -
cat deploy/provider/baremetal/service-nodeport.yaml | kubectl apply -f -

echo "updating image..."
kubectl set image \
    deployments \
    --namespace ingress-nginx \
	--selector app=ingress-nginx \
    nginx-ingress-controller=quay.io/kubernetes-ingress-controller/nginx-ingress-controller:test

sleep 5

echo "waiting NGINX ingress pod..."

function waitForPod() {
    until kubectl get pods -n ingress-nginx -l app=ingress-nginx -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True";
    do
        sleep 1;
    done
}

export -f waitForPod

timeout 30s bash -c waitForPod

if kubectl get pods -n ingress-nginx -l app=ingress-nginx -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True";
then
    echo "Kubernetes deployments started"
else
    echo "Kubernetes deployments with issues:"
    kubectl get pods -n ingress-nginx

    echo "Reason:"
    kubectl describe pods -n ingress-nginx
    kubectl logs -n ingress-nginx -l app=ingress-nginx
    exit 1
fi
