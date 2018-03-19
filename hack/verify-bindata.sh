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

TMP_DIR="_tmp"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap "cleanup" EXIT SIGINT

cleanup

go-bindata -nometadata -o ${TMP_DIR}/bindata.go -prefix="rootfs" -pkg=file -ignore=Dockerfile -ignore=".DS_Store" rootfs/...

ret=0
diff -Naupr "${TMP_DIR}/bindata.go" "internal/file/bindata.go" || ret=$?
if [[ ${ret} -eq 0 ]]
then
  echo "bindata.go up to date."
else
  echo "bindata.go is out of date. Please run make code-generator and NOT gofmt internal/file/bindata.go"
  exit 1
fi
