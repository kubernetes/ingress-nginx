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

if ! [ -z "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

RED='\e[35m'
NC='\e[0m'
BGREEN='\e[32m'

declare -a mandatory
mandatory=(
  E2E_NODES
)

missing=false
for var in "${mandatory[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo -e "${RED}Environment variable $var must be set${NC}"
    missing=true
  fi
done

if [ "$missing" = true ]; then
  exit 1
fi

function cleanup {
  kubectl delete pod e2e 2>/dev/null || true
}
trap cleanup EXIT

E2E_CHECK_LEAKS=${E2E_CHECK_LEAKS:-}
FOCUS=${FOCUS:-.*}

BASEDIR=$(dirname "$0")
NGINX_BASE_IMAGE=$(cat $BASEDIR/../NGINX_BASE)

export E2E_CHECK_LEAKS
export FOCUS

echo -e "${BGREEN}Granting permissions to ingress-nginx e2e service account...${NC}"
kubectl create serviceaccount ingress-nginx-e2e || true
kubectl create clusterrolebinding permissive-binding \
  --clusterrole=cluster-admin \
  --user=admin \
  --user=kubelet \
  --serviceaccount=default:ingress-nginx-e2e || true


VER=$(kubectl version  --client=false -o json |jq '.serverVersion.minor |tonumber')
if [ $VER -lt 24 ]; then
  echo -e "${BGREEN}Waiting service account...${NC}"; \
  until kubectl get secret | grep -q -e ^ingress-nginx-e2e-token; do \
    echo -e "waiting for api token"; \
    sleep 3; \
  done
fi


echo -e "Starting the e2e test pod"

kubectl run --rm \
  --attach \
  --restart=Never \
  --env="E2E_NODES=${E2E_NODES}" \
  --env="FOCUS=${FOCUS}" \
  --env="E2E_CHECK_LEAKS=${E2E_CHECK_LEAKS}" \
  --env="NGINX_BASE_IMAGE=${NGINX_BASE_IMAGE}" \
  --overrides='{ "apiVersion": "v1", "spec":{"serviceAccountName": "ingress-nginx-e2e"}}' \
  e2e --image=nginx-ingress-controller:e2e

# Get the junit-reports stored in the configMaps created during e2etests
echo "Getting the report files out now.."
reportsDir="test/junitreports"
reportFileName="report-e2e-test-suite"
[ ! -e ${reportsDir} ] && mkdir $reportsDir
cd $reportsDir

# TODO: Seeking Rikatz help here to extract in a loop. Tried things like below without success
#for cmName in `k get cm -l junitreport=true -o json | jq  '.items[].binaryData | keys[]' | tr '\"' ' '`
#do
#
#
# kubectl get cm -l junitreport=true -o json | jq -r  '[.items[].binaryData | to_entries[] | {"key": .key, "value": .value  }] | from_entries'
#

# Below lines successfully extract the report but they are one line per report.
# We only have 3 ginkgo reports so its ok for now
# But still, ideally this should be a loop as talked about in comments a few lines above
kubectl get cm $reportFileName.xml.gz -o "jsonpath={.binaryData['report-e2e-test-suite\.xml\.gz']}" > $reportFileName.xml.gz.base64
kubectl get cm $reportFileName-serial.xml.gz -o "jsonpath={.binaryData['report-e2e-test-suite-serial\.xml\.gz']}" > $reportFileName-serial.xml.gz.base64

cat $reportFileName.xml.gz.base64 | base64 -d > $reportFileName.xml.gz
cat $reportFileName-serial.xml.gz.base64 | base64 -d > $reportFileName-serial.xml.gz

gzip -d $reportFileName.xml.gz
gzip -d $reportFileName-serial.xml.gz

rm *.base64
cd ../..

# TODO Temporary: if condition to check if the memleak cm exists and only then try the extract for the memleak report
#
#kubectl get cm $reportFileName-serial  -o "jsonpath={.data['report-e2e-test-suite-memleak\.xml\.gz']}" > $reportFileName-memleak.base64
#cat $reportFileName-memleak.base64 | base64 -d > $reportFileName-memleak.xml.gz
#gzip -d $reportFileName-memleak.xml.gz
echo "done getting the reports files out.."
