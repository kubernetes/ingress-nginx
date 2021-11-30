#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
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

DIR=$(cd $(dirname "${BASH_SOURCE}")/.. && pwd -P)

RELEASE_NAME=ingress-nginx
NAMESPACE=ingress-nginx

NAMESPACE_VAR="
apiVersion: v1
kind: Namespace
metadata:
  name: $NAMESPACE
  labels:
    app.kubernetes.io/name: $RELEASE_NAME
    app.kubernetes.io/instance: ingress-nginx
"

# clean
# BUG # TODO
# rm -rf ${DIR}/deploy/static/provider/*



  # VALUES_FILE="./helm-values/baremetal/deploy.yaml"
  # OUTPUT_FILE="${DIR}/deploy/static/provider/baremetal/deploy.yaml"
TARGET_FILES=$(cd $DIR/hack/helm-values/ && find . -type f -name "*.yaml" | cut -d'/' -f2-)
# echo $TARGET_FILES
# exit 2
for TARGET_FILE in ${TARGET_FILES}
do
  VALUES_FILE="$DIR/hack/helm-values/${TARGET_FILE}"
  OUTPUT_FILE="${DIR}/deploy/static/provider/${TARGET_FILE}"

  echo "${NAMESPACE_VAR}" > ${OUTPUT_FILE}
  helm template $RELEASE_NAME ${DIR}/charts/ingress-nginx --values ${VALUES_FILE} --namespace $NAMESPACE | $DIR/hack/add-namespace.py $NAMESPACE >> ${OUTPUT_FILE}
  # helm template $RELEASE_NAME ${DIR}/charts/ingress-nginx --values ${VALUES_FILE} --namespace $NAMESPACE >> ${OUTPUT_FILE}
done
