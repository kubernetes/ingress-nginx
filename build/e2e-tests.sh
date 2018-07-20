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
if [ -z "${FOCUS}" ]; then
    echo "FOCUS must be set"
    exit 1
fi
if [ -z "${E2E_NODES}" ]; then
    echo "E2E_NODES must be set"
    exit 1
fi
if [ -z "${NODE_IP}" ]; then
    echo "NODE_IP must be set"
    exit 1
fi

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

mkdir -p ${SCRIPT_ROOT}/test/binaries

TEST_BINARIES=$( cd "${SCRIPT_ROOT}/test/binaries" ; pwd -P )

export PATH=${TEST_BINARIES}:$PATH

if ! [ -x "$(command -v kubectl)" ]; then
    echo "downloading kubectl..."
    curl -sSLo ${TEST_BINARIES}/kubectl \
        https://storage.googleapis.com/kubernetes-release/release/v1.11.0/bin/linux/amd64/kubectl
    chmod +x ${TEST_BINARIES}/kubectl
fi

ginkgo build ./test/e2e

exec --                 \
ginkgo                  \
    -randomizeSuites    \
    -randomizeAllSpecs  \
    -flakeAttempts=2    \
    --focus=${FOCUS}    \
    -p                  \
    -trace              \
    -nodes=${E2E_NODES} \
    test/e2e/e2e.test
