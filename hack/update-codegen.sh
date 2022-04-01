#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
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

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_VERSION=$(grep 'k8s.io/code-generator' go.sum | awk '{print $2}' | sed 's/\/go.mod//g' | tail -1)
CODEGEN_PKG=$(echo `go env GOPATH`"/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION}")

if [[ ! -d ${CODEGEN_PKG} ]]; then
  echo "${CODEGEN_PKG} is missing. Running 'go mod download'."
  go mod download
fi

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
#${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
#  k8s.io/ingress-nginx/pkg/client k8s.io/ingress-nginx/pkg/apis \
#  nginxingress:v1alpha1 \
#  --output-base "$(dirname ${BASH_SOURCE})/../../.."
${CODEGEN_PKG}/generate-groups.sh "deepcopy" \
  k8s.io/ingress-nginx/internal k8s.io/ingress-nginx/internal \
  .:ingress \
  --output-base "$(dirname ${BASH_SOURCE})/../../.." \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generated.go.txt
