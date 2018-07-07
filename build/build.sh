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
if [ -z "${TAG}" ]; then
    echo "TAG must be set"
    exit 1
fi

export CGO_ENABLED=0

go build -a -installsuffix cgo 	\
    -ldflags "-s -w \
        -X ${PKG}/version.RELEASE=${TAG} \
        -X ${PKG}/version.COMMIT=${GIT_COMMIT} \
        -X ${PKG}/version.REPO=${REPO_INFO}" \
    -o bin/${ARCH}/nginx-ingress-controller ${PKG}/cmd/nginx
