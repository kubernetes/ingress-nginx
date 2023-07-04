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

#set -o errexit
set -o nounset
set -o pipefail

K8S_VERSION="1.22"

DIR=$(cd $(dirname "${BASH_SOURCE}")/.. && pwd -P)

# clean
rm -rf ${DIR}/deploy/static/provider/*

TEMPLATE_DIR="${DIR}/hack/manifest-templates"

# each helm values file `values.yaml` under `hack/manifest-templates/provider` will be generated as provider/<provider>[/variant]/deploy.yaml
# TARGET is provider/<provider>[/variant]
TARGETS=$(dirname $(cd $DIR/hack/manifest-templates/ && find . -type f -name "values.yaml" ) | cut -d'/' -f2-)
for TARGET in ${TARGETS}
do
  TARGET_DIR="${TEMPLATE_DIR}/${TARGET}"
  MANIFEST="${TEMPLATE_DIR}/common/manifest.yaml" # intermediate manifest
  OUTPUT_DIR="${DIR}/deploy/static/${TARGET}"
  echo $OUTPUT_DIR

  mkdir -p ${OUTPUT_DIR}
  cd ${TARGET_DIR}
  helm template ingress-nginx ${DIR}/charts/ingress-nginx \
    --values values.yaml \
    --namespace ingress-nginx \
    --kube-version ${K8S_VERSION} \
    > $MANIFEST
  sed -i.bak '/app.kubernetes.io\/managed-by: Helm/d' $MANIFEST
  sed -i.bak '/helm.sh/d' $MANIFEST

  kustomize --load-restrictor=LoadRestrictionsNone build . > ${OUTPUT_DIR}/deploy.yaml
  rm $MANIFEST $MANIFEST.bak
  cd ~-
  # automatically generate the (unsupported) kustomization.yaml for each target
  sed "s_{TARGET}_${TARGET}_" $TEMPLATE_DIR/static-kustomization-template.yaml > ${OUTPUT_DIR}/kustomization.yaml
done
