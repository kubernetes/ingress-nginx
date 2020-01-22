#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
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

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export TAG=dev
export ARCH=amd64
export DOCKER_CLI_EXPERIMENTAL=enabled
export CUSTOM_REGISTRY=true

echo "Build and push e2e test images"
echo "
make -C ${DIR}/../../ build container push
make -C ${DIR}/../../ e2e-test-binary
make -C ${DIR}/../../test/e2e-image container push
make -C ${DIR}/../../images/echo/ container push
make -C ${DIR}/../../images/fastcgi-helloserver/ GO111MODULE="on" build container push
make -C ${DIR}/../../images/httpbin/ container push
" | parallel --joblog /tmp/log {} || cat /tmp/log

echo "Start e2e test suite"
make -C ${DIR}/../../ e2e-test
