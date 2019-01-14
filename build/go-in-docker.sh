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

set -o errexit
set -o nounset
set -o pipefail

declare -a mandatory
mandatory=(
  PKG
  ARCH
  GIT_COMMIT
  REPO_INFO
  TAG
  HOME
)

missing=false
for var in ${mandatory[@]}; do
  if [[ -z "${!var+x}" ]]; then
    echo "Environment variable $var must be set"
    missing=true
  fi
done

if [ "$missing" = true ];then
  exit 1
fi

E2E_IMAGE=quay.io/kubernetes-ingress-controller/e2e:v01092019-b433108ea

DOCKER_OPTS=${DOCKER_OPTS:-""}

FLAGS=$@

tee .env << EOF
PKG=${PKG:-""}
ARCH=${ARCH:-""}
GIT_COMMIT=${GIT_COMMIT:-""}
E2E_NODES=${E2E_NODES:-4}
FOCUS=${FOCUS:-.*}
TAG=${TAG:-"0.0"}
HOME=${HOME:-/root}
KUBECONFIG=${HOME}/.kube/config
GOARCH=${GOARCH}
GOBUILD_FLAGS=${GOBUILD_FLAGS:-"-v"}
PWD=${PWD}
BUSTED_ARGS=${BUSTED_ARGS:-""}
REPO_INFO=${REPO_INFO:-local}
NODE_IP=${NODE_IP:-127.0.0.1}
SLOW_E2E_THRESHOLD=${SLOW_E2E_THRESHOLD:-40}
EOF

MINIKUBE_PATH=${HOME}/.minikube
MINIKUBE_VOLUME="-v ${MINIKUBE_PATH}:${MINIKUBE_PATH}"
if [ ! -d ${MINIKUBE_PATH} ]; then
    echo "Minikube directory not found! Volume will be excluded from docker build."
    MINIKUBE_VOLUME=""
fi

docker run                                       \
    --tty                                        \
    --rm                                         \
    ${DOCKER_OPTS}                               \
    -v ${HOME}/.kube:/${HOME}/.kube              \
    -v ${PWD}:/go/src/${PKG}                     \
    -v ${PWD}/.gocache:${HOME}/.cache/go-build   \
    -v ${PWD}/bin/${ARCH}:/go/bin/linux_${ARCH}  \
    ${MINIKUBE_VOLUME}                           \
    -w /go/src/${PKG}                            \
    --env-file .env                              \
    --entrypoint ${FLAGS}                        \
    ${E2E_IMAGE}

rm .env
