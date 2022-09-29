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
export IS_DATAPLANE=$2
export NAMESPACE_OVERLAY=$3

echo "deploying NGINX Ingress controller in namespace $NAMESPACE"

EXTRA_DATAPLANE_ARG=""

if [ "${IS_DATAPLANE:-false}" = "true" ]; then
   EXTRA_DATAPLANE_ARG=" --set useDataplaneMode=true --set controller.livenessProbe=null --set controller.readinessProbe=null "
fi 

function on_exit {
    local error_code="$?"

    test $error_code == 0 && return;

    echo "Obtaining ingress controller pod logs..."
    kubectl logs -l app.kubernetes.io/name=ingress-nginx -n $NAMESPACE
}
trap on_exit EXIT

cat << EOF | kubectl apply --namespace=$NAMESPACE -f -
# Required for e2e tcp tests
kind: ConfigMap
apiVersion: v1
metadata:
  name: tcp-services
  namespace: $NAMESPACE

EOF

# Use the namespace overlay if it was requested
if [[ ! -z "$NAMESPACE_OVERLAY" && -d "$DIR/namespace-overlays/$NAMESPACE_OVERLAY" ]]; then
    echo "Namespace overlay $NAMESPACE_OVERLAY is being used for namespace $NAMESPACE"
    helm install nginx-ingress ${DIR}/charts/ingress-nginx \
        --namespace=$NAMESPACE ${EXTRA_DATAPLANE_ARG} \
        --values "$DIR/namespace-overlays/$NAMESPACE_OVERLAY/values.yaml"
else
    helm install nginx-ingress ${DIR}/charts/ingress-nginx \
       --namespace=$NAMESPACE -f ${DIR}/ci-values.yaml ${EXTRA_DATAPLANE_ARG} \
       --set controller.extraArgs.tcp-services-configmap=$NAMESPACE/tcp-services
fi
