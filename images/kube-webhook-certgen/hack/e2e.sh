#!/bin/bash

# Copyright 2021 The Kubernetes Authors.
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

pushd rootfs
go build
popd

export KUBECONFIG=${KUBECONFIG:-~/.kube/config}

function error {
  trap - EXIT
  echo $@
  exit 1
}

function cleanup {
  kubectl delete --recursive -f hack/e2e.yaml --ignore-not-found
}

trap cleanup EXIT

cleanup

kubectl apply --recursive -f hack/e2e.yaml

./rootfs/kube-webhook-certgen create --kubeconfig ${KUBECONFIG} --namespace test --secret-name test --host test --cert-name test
./rootfs/kube-webhook-certgen patch --kubeconfig ${KUBECONFIG} --namespace test --secret-name test --webhook-name pod-policy.example.com
./rootfs/kube-webhook-certgen patch --kubeconfig ${KUBECONFIG} --namespace test --secret-name test --apiservice-name v1alpha1.pod-policy.example.com
./rootfs/kube-webhook-certgen patch --kubeconfig ${KUBECONFIG} --namespace test --secret-name test --webhook-name pod-policy.example.com --apiservice-name v1alpha1.pod-policy.example.com

if [[ -z "$(kubectl get validatingwebhookconfiguration pod-policy.example.com -o jsonpath='{.webhooks[0].clientConfig.caBundle}')" ]]; then
  error "ValidatingWebhookConfiguration pod-policy.example.com did not get CA injected"
fi

if [[ -z "$(kubectl get mutatingwebhookconfiguration pod-policy.example.com -o jsonpath='{.webhooks[0].clientConfig.caBundle}')" ]]; then
  error "MutatingWebhookConfiguration pod-policy.example.com did not get CA injected"
fi

if [[ -z "$(kubectl get apiservice v1alpha1.pod-policy.example.com -o jsonpath='{.spec.caBundle}')" ]];then
  error "APIService v1alpha1.pod-policy.example.com did not get CA injected"
fi
