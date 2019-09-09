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

if [ -n "$DEBUG" ]; then
	set -x
fi

set -o errexit
set -o nounset
set -o pipefail

declare -a mandatory
mandatory=(
  AWS_ACCESS_KEY
  AWS_SECRET_KEY
)

missing=false
for var in "${mandatory[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "Environment variable $var must be set"
    missing=true
  fi
done

if [ "$missing" = true ]; then
  exit 1
fi

DIR=$(cd $(dirname "${BASH_SOURCE}") && pwd -P)

# build local terraform image to build nginx
docker build -t build-nginx-terraform $DIR/images/nginx

# build nginx and publish docker images to quay.io.
# this can take up to two hours.
docker run --rm -it \
  --volume $DIR/images/nginx:/tf \
  -w /tf \
  --env AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
  --env AWS_SECRET_KEY=${AWS_SECRET_KEY} \
  --env AWS_SECRET_KEY=${AWS_SECRET_KEY} \
  --env QUAY_USERNAME=${QUAY_USERNAME} \
  --env QUAY_PASSWORD="${QUAY_PASSWORD}" \
  build-nginx-terraform
