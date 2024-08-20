#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

ANNOTATIONFILE="${SCRIPT_ROOT}/docs/user-guide/nginx-configuration/annotations-risk.md"
TMP_DIFFROOT="${SCRIPT_ROOT}/_tmp"
TMP_FILE="${TMP_DIFFROOT}/annotations-risk.md"

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${TMP_DIFFROOT}"

go run cmd/annotations/main.go -output "${TMP_FILE}"
echo "diffing ${ANNOTATIONFILE} against freshly generated annotation doc"
ret=0
diff -Naupr --no-dereference "${ANNOTATIONFILE}" "${TMP_FILE}" || ret=1

if [[ $ret -eq 0 ]]; then
  echo "${ANNOTATIONFILE} up to date."
else
  echo "${ANNOTATIONFILE} is out of date. Please run hack/update-annotation-doc.sh"
  exit 1
fi
