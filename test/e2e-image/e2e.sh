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

set -eu

NC='\e[0m'
BGREEN='\e[32m'

E2E_NODES=${E2E_NODES:-5}
E2E_CHECK_LEAKS=${E2E_CHECK_LEAKS:-""}

reportFile="report-e2e-test-suite.xml"
ginkgo_args=(
  "--fail-fast"
  "--flake-attempts=2"
  "--junit-report=${reportFile}"
  "--nodes=${E2E_NODES}"
  "--poll-progress-after=180s"
  "--randomize-all"
  "--show-node-events"
  "--succinct"
  "--timeout=75m"
)

if [ -n "${FOCUS}" ]; then
  ginkgo_args+=("--focus=${FOCUS}")
fi

if [ -z "${E2E_CHECK_LEAKS}" ]; then
  ginkgo_args+=("--skip=\[Memory Leak\]")
fi

echo -e "${BGREEN}Running e2e test suite...${NC}"
(set -x; ginkgo "${ginkgo_args[@]}" /e2e.test)

# Create configMap out of a compressed report file for extraction later
gzip -k ${reportFile}
kubectl create cm ${reportFile}.gz --from-file ${reportFile}.gz
kubectl label cm ${reportFile}.gz junitreport=true
