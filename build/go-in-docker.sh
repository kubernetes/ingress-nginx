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

if [ -z "${PKG}" ]; then
    echo "PKG must be set"
    exit 1
fi
if [ -z "${ARCH}" ]; then
    echo "ARCH must be set"
    exit 1
fi
if [ -z "${GIT_COMMIT}" ]; then
    echo "GIT_COMMIT must be set"
    exit 1
fi
if [ -z "${REPO_INFO}" ]; then
    echo "REPO_INFO must be set"
    exit 1
fi
if [ -z "${TAG}" ]; then
    echo "TAG must be set"
    exit 1
fi
if [ -z "${HOME}" ]; then
    echo "HOME must be set"
    exit 1
fi

DOCKER_OPTS=${DOCKER_OPTS:-""}

docker build -t ingress-nginx:build build

FLAGS=$@

tee .env << EOF
PKG=${PKG:-""}
ARCH=${ARCH:-""}
GIT_COMMIT=${GIT_COMMIT:-""}
E2E_NODES=${E2E_NODES:-3}
FOCUS=${FOCUS:-.*}
TAG=${TAG:-"0.0"}
HOME=${HOME:-/root}
KUBECONFIG=${HOME}/.kube/config
GOARCH=${GOARCH}
PWD=${PWD}
BUSTED_ARGS=${BUSTED_ARGS:-""}
REPO_INFO=${REPO_INFO:-local}
NODE_IP=${NODE_IP:-127.0.0.1}
EOF

docker run                                       \
    --tty                                        \
    --rm                                         \
    ${DOCKER_OPTS}                               \
    -v ${HOME}/.kube:/${HOME}/.kube              \
    -v ${HOME}/.minikube:${HOME}/.minikube       \
    -v ${PWD}:/go/src/${PKG}                     \
    -v ${PWD}/.gocache:${HOME}/.cache/go-build   \
    -v ${PWD}/bin/${ARCH}:/go/bin/linux_${ARCH}  \
    -w /go/src/${PKG}                            \
    --env-file .env                              \
    ingress-nginx:build ${FLAGS}

rm .env
