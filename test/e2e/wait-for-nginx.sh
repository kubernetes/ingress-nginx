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

set -e
set -x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export NAMESPACE=$1

echo "deploying NGINX Ingress controller in namespace $NAMESPACE"

function on_exit {
    local error_code="$?"

    test $error_code == 0 && return;

    echo "Obtaining ingress controller pod logs..."
    kubectl logs -l app.kubernetes.io/name=ingress-nginx -n $NAMESPACE
}
trap on_exit EXIT

kubectl apply --namespace=$NAMESPACE -f $DIR/manifests/service.yaml

sed "s@\${NAMESPACE}@${NAMESPACE}@" $DIR/manifests/mandatory.yaml | kubectl apply --namespace=$NAMESPACE -f -
cat $DIR/manifests/service.yaml | kubectl apply --namespace=$NAMESPACE -f -

# wait for the deployment and fail if there is an error before starting the execution of any test
kubectl rollout status \
    --request-timeout=3m \
    --namespace $NAMESPACE \
    deployment nginx-ingress-controller
