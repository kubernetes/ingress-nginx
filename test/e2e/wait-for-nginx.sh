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
if ! [ -z $DEBUG ]; then
	set -x
fi

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

CLUSTER_WIDE="$DIR/cluster-wide-$NAMESPACE"

mkdir "$CLUSTER_WIDE"

cat << EOF > "$CLUSTER_WIDE/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../cluster-wide
nameSuffix: "-$NAMESPACE"
EOF

OVERLAY="$DIR/overlay-$NAMESPACE"

mkdir "$OVERLAY"

cat << EOF > "$OVERLAY/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: $NAMESPACE
bases:
- ../overlay
- ../cluster-wide-$NAMESPACE
EOF

kubectl apply --kustomize "$OVERLAY"

# wait for the deployment and fail if there is an error before starting the execution of any test
kubectl rollout status \
    --request-timeout=3m \
    --namespace $NAMESPACE \
    deployment nginx-ingress-controller
