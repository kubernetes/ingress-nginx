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

# clean
rm -rf ${DIR}/deploy/static/provider/*

TEMPLATE_DIR="${DIR}/hack/manifest-templates"

# each helm values file `values.yaml` will be generated as provider/<provider>[/variant]/deploy.yaml
# TARGET is provider/<provider>[/variant]
TARGETS=$(dirname $(cd $DIR/hack/manifest-templates/ && find . -type f -name "values.yaml" ) | cut -d'/' -f2-)
echo $TARGETS
for TARGET in ${TARGETS}
do
  TARGET_DIR="${TEMPLATE_DIR}/${TARGET}"
  OUTPUT_DIR="${DIR}/deploy/static/${TARGET}"
  MANIFEST=manifest.yaml # intermediate manifest

  mkdir -p ${OUTPUT_DIR}
  pushd ${TARGET_DIR}
  helm template ingress-nginx ${DIR}/charts/ingress-nginx --values values.yaml --namespace ingress-nginx > $MANIFEST
  kustomize --load-restrictor=LoadRestrictionsNone build . > $OUTPUT_DIR/deploy.yaml
  rm $MANIFEST
  popd
  # automatically generate the (unsupported) kustomization.yaml for each target
  sed "s_{TARGET}_${TARGET}_" $TEMPLATE_DIR/kustomization-template.yaml > $OUTPUT_DIR/kustomization.yaml
done
