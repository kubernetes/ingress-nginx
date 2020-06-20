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

if [ -n "$DEBUG" ]; then
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
  IMAGE
  ARCH
  TAG
)

missing=false
for var in "${mandatory[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "${RED}Environment variable $var must be set${NC}"
    missing=true
  fi
done

if [ "$missing" = true ]; then
  exit 1
fi

# temporal directory for the fake SSL certificate
SSL_VOLUME=$(mktemp -d)

function cleanup {
    echo -e "${BGREEN}Stoping kubectl proxy${NC}"
    rm -rf "${SSL_VOLUME}"
    kill "$proxy_pid"
}
trap cleanup EXIT

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..

# the ingress controller needs this two variables. To avoid the
# creation of any object in the cluster we use invalid names.
POD_NAMESPACE="invalid-namespace"
POD_NAME="invalid-namespace"

export TAG
export IMAGE

if [[ "${ARCH}" != "amd64" ]]; then
  echo -e "${BGREEN}Register ${RED}/usr/bin/qemu-ARCH-static${BGREEN} as the handler for binaries in multiple platforms${NC}"
  make -C "${KUBE_ROOT}" register-qemu
fi

USE_EXISTING_IMAGE=${USE_EXISTING_IMAGE:-false}
if [[ "${USE_EXISTING_IMAGE}" == "true" ]]; then
  echo -e "${BGREEN}Downloading ingress controller image${NC}"
  docker pull "${IMAGE}-${ARCH}:${TAG}"
else
  echo -e "${BGREEN}Building ingress controller image${NC}"
  make -C "${KUBE_ROOT}" build "sub-image-${ARCH}"
fi

CONTEXT=$(kubectl config current-context)

echo -e "Running against kubectl cluster ${BGREEN}${CONTEXT}${NC}"

kubectl proxy --accept-hosts=.* --address=0.0.0.0 &
proxy_pid=$!
sleep 1

echo -e "\n${BGREEN}kubectl proxy PID: ${BGREEN}$proxy_pid${NC}"

until curl --output /dev/null -fsSL http://localhost:8001/; do
  echo -e "${RED}waiting for kubectl proxy${NC}"
  sleep 5
done

# if we run as user we cannot bind to port 80 and 443
docker run \
  --rm \
  --name local-ingress-controller \
  --net=host \
  --user="root:root" \
  -e POD_NAMESPACE="${POD_NAMESPACE}" \
  -e POD_NAME="${POD_NAME}" \
  -v "${SSL_VOLUME}:/etc/ingress-controller/ssl/" \
  -v "${HOME}/.kube:${HOME}/.kube:ro" \
  "${IMAGE}-${ARCH}:${TAG}" /nginx-ingress-controller \
  --update-status=false \
  --v=2 \
  --apiserver-host=http://0.0.0.0:8001 \
  --kubeconfig="${HOME}/.kube/config"
