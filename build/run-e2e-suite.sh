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

if ! [ -z "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

RED='\e[35m'
NC='\e[0m'
BGREEN='\e[32m'

declare -a mandatory
mandatory=(
  E2E_NODES
)

missing=false
for var in "${mandatory[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo -e "${RED}Environment variable $var must be set${NC}"
    missing=true
  fi
done

if [ "$missing" = true ]; then
  exit 1
fi

function cleanup {
  kubectl delete pod e2e 2>/dev/null || true
}
trap cleanup EXIT

E2E_CHECK_LEAKS=${E2E_CHECK_LEAKS:-}
FOCUS=${FOCUS:-.*}

export E2E_CHECK_LEAKS
export FOCUS

echo -e "${BGREEN}Granting permissions to ingress-nginx e2e service account...${NC}"
kubectl create serviceaccount ingress-nginx-e2e || true
kubectl create clusterrolebinding permissive-binding \
  --clusterrole=cluster-admin \
  --user=admin \
  --user=kubelet \
  --serviceaccount=default:ingress-nginx-e2e || true

echo -e "${BGREEN}Waiting service account...${NC}"; \
until kubectl get secret | grep -q -e ^ingress-nginx-e2e-token; do \
  echo -e "waiting for api token"; \
  sleep 3; \
done

echo -e "Starting the e2e test pod"

kubectl run --rm \
  --attach \
  --restart=Never \
  --env="E2E_NODES=${E2E_NODES}" \
  --env="FOCUS=${FOCUS}" \
  --env="E2E_CHECK_LEAKS=${E2E_CHECK_LEAKS}" \
  --overrides='{ "apiVersion": "v1", "spec":{"serviceAccountName": "ingress-nginx-e2e"}}' \
  e2e --image=nginx-ingress-controller:e2e
