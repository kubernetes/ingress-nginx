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

set -e

SLOW_E2E_THRESHOLD=${SLOW_E2E_THRESHOLD:-50}
FOCUS=${FOCUS:-.*}
E2E_NODES=${E2E_NODES:-5}

if [ ! -f ${HOME}/.kube/config ]; then
    kubectl config set-cluster dev --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt --embed-certs=true --server="https://kubernetes.default/"
    kubectl config set-credentials user --token="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
    kubectl config set-context default --cluster=dev --user=user
    kubectl config use-context default
fi

echo "Granting permissions to ingress-nginx e2e service account..."
kubectl create serviceaccount ingress-nginx-e2e || true
kubectl create clusterrolebinding permissive-binding \
--clusterrole=cluster-admin \
--user=admin \
--user=kubelet \
--serviceaccount=default:ingress-nginx-e2e || true

kubectl apply -f manifests/rbac.yaml

ginkgo_args=(
    "-randomizeSuites"
    "-randomizeAllSpecs"
    "-flakeAttempts=2"
    "-p"
    "-trace"
    "--noColor=true"
    "-slowSpecThreshold=${SLOW_E2E_THRESHOLD}"
)

echo "Running e2e test suite..."
ginkgo "${ginkgo_args[@]}"                   \
    -focus=${FOCUS}                          \
    -skip="\[Serial\]"                       \
    -nodes=${E2E_NODES}                      \
    /e2e.test

echo "Running e2e test suite with tests that require serial execution..."
ginkgo "${ginkgo_args[@]}"                   \
    -focus="\[Serial\]"                      \
    -nodes=1                                 \
    /e2e.test
