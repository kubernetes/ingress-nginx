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

echo "downloading kubectl..."
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl && \
    chmod +x kubectl && sudo mv kubectl /usr/local/bin/

echo "downloading minikube..."
curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.25.2/minikube-linux-amd64 && \
    chmod +x minikube && \
    sudo mv minikube /usr/local/bin/

echo "starting minikube..."
# Using a lower value for sync-frequency to speed up the tests (during the cleanup of resources inside a namespace)
sudo minikube start --vm-driver=none --kubernetes-version=$KUBERNETES_VERSION --extra-config=kubelet.sync-frequency=1s

minikube update-context

echo "waiting for kubernetes cluster"
until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True";
do
    sleep 1;
done
