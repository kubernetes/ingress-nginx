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

set -e

NC='\e[0m'
BGREEN='\e[32m'

#SLOW_E2E_THRESHOLD=${SLOW_E2E_THRESHOLD:-"5s"}
FOCUS=${FOCUS:-.*}
E2E_NODES=${E2E_NODES:-5}
E2E_CHECK_LEAKS=${E2E_CHECK_LEAKS:-""}

ginkgo_args=(
  "-randomize-all"
  "-flake-attempts=2"
  "-fail-fast"
  "--show-node-events"
  "--poll-progress-after=180s"
#  "-slow-spec-threshold=${SLOW_E2E_THRESHOLD}"
  "-succinct"
  "-timeout=75m"
)

# Variable for the prefix of report filenames
reportFileNamePrefix="report-e2e-test-suite"

echo -e "${BGREEN}Running e2e test suite (FOCUS=${FOCUS})...${NC}"
ginkgo "${ginkgo_args[@]}"               \
  -focus="${FOCUS}"                  \
  -skip="\[Serial\]|\[MemoryLeak\]|\[TopologyHints\]"  \
  -nodes="${E2E_NODES}" \
  --junit-report=$reportFileNamePrefix.xml \
  /e2e.test
# Create configMap out of a compressed report file for extraction later

# Must be isolated, there is a collision if multiple helms tries to install same clusterRole at same time
echo -e "${BGREEN}Running e2e test for topology aware hints...${NC}"
ginkgo "${ginkgo_args[@]}" \
  -focus="\[TopologyHints\]" \
  -skip="\[Serial\]|\[MemoryLeak\]]" \
  -nodes="${E2E_NODES}" \
  --junit-report=$reportFileNamePrefix-topology.xml \
  /e2e.test
# Create configMap out of a compressed report file for extraction later

echo -e "${BGREEN}Running e2e test suite with tests that require serial execution...${NC}"
ginkgo "${ginkgo_args[@]}"   \
  -focus="\[Serial\]"    \
  -skip="\[MemoryLeak\]" \
  --junit-report=$reportFileNamePrefix-serial.xml \
  /e2e.test
# Create configMap out of a compressed report file for extraction later

if [[ ${E2E_CHECK_LEAKS} != "" ]]; then
  echo -e "${BGREEN}Running e2e test suite with tests that check for memory leaks...${NC}"
  ginkgo "${ginkgo_args[@]}"    \
    -focus="\[MemoryLeak\]" \
    -skip="\[Serial\]" \
    --junit-report=$reportFileNamePrefix-memleak.xml \
    /e2e.test
# Create configMap out of a compressed report file for extraction later
fi

for rFile in `ls $reportFileNamePrefix*` 
do
  gzip -k $rFile
  kubectl create cm $rFile.gz --from-file $rFile.gz
  kubectl label cm $rFile.gz junitreport=true
done
