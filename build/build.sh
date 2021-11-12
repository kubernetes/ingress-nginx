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

declare -a mandatory
mandatory=(
  PKG
  ARCH
  COMMIT_SHA
  REPO_INFO
  TAG
)

missing=false
for var in "${mandatory[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "Environment variable $var must be set"
    missing=true
  fi
done

if [ "$missing" = true ]; then
  exit 1
fi

export CGO_ENABLED=0
export GOARCH=${ARCH}

TARGETS_DIR="rootfs/bin/${ARCH}"
echo "Building targets for ${ARCH}, generated targets in ${TARGETS_DIR} directory."

go build \
  -trimpath -ldflags="-buildid= -w -s \
    -X ${PKG}/version.RELEASE=${TAG} \
    -X ${PKG}/version.COMMIT=${COMMIT_SHA} \
    -X ${PKG}/version.REPO=${REPO_INFO}" \
  -o "${TARGETS_DIR}/nginx-ingress-controller" "${PKG}/cmd/nginx"

go build \
  -trimpath -ldflags="-buildid= -w -s \
    -X ${PKG}/version.RELEASE=${TAG} \
    -X ${PKG}/version.COMMIT=${COMMIT_SHA} \
    -X ${PKG}/version.REPO=${REPO_INFO}" \
  -o "${TARGETS_DIR}/dbg" "${PKG}/cmd/dbg"

go build \
  -trimpath -ldflags="-buildid= -w -s \
    -X ${PKG}/version.RELEASE=${TAG} \
    -X ${PKG}/version.COMMIT=${COMMIT_SHA} \
    -X ${PKG}/version.REPO=${REPO_INFO}" \
  -o "${TARGETS_DIR}/wait-shutdown" "${PKG}/cmd/waitshutdown"
