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

if ! [ -z $DEBUG ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

E2E_IMAGE=quay.io/kubernetes-ingress-controller/e2e:v05122019-7b3f69a3ba

DOCKER_OPTS=${DOCKER_OPTS:-""}

FLAGS=$@

tee .env << EOF
PKG=${PKG:-""}
ARCH=${ARCH:-""}
GIT_COMMIT=${GIT_COMMIT:-""}
TAG=${TAG:-"0.0"}
GOARCH=${GOARCH:-""}
GOBUILD_FLAGS=${GOBUILD_FLAGS:-"-v"}
PWD=${PWD}
BUSTED_ARGS=${BUSTED_ARGS:-""}
REPO_INFO=${REPO_INFO:-local}
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
    -v /var/run/docker.sock:/var/run/docker.sock \
    ${MINIKUBE_VOLUME}                           \
    -w /go/src/${PKG}                            \
    --env-file .env                              \
    --entrypoint ${FLAGS}                        \
    ${E2E_IMAGE}

rm .env
