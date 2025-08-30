#!/bin/bash

# Copyright 2025 The Kubernetes Authors.
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

# This script extracts the ginkgo version from go.mod
# Usage: ./hack/get-ginkgo-version.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

# Extract ginkgo version from go.mod
GINKGO_VERSION=$(grep -E '^\s*github\.com/onsi/ginkgo/v2\s+' "${SCRIPT_ROOT}/go.mod" | awk '{print $2}' | sed 's/^v//')

if [[ -z "${GINKGO_VERSION}" ]]; then
    echo "Error: Could not find ginkgo version in go.mod" >&2
    echo "Expected format: github.com/onsi/ginkgo/v2 vX.Y.Z" >&2
    exit 1
fi

echo "${GINKGO_VERSION}"
