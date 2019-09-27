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

DIR=$(cd $(dirname "${BASH_SOURCE}") && pwd -P)

AWS_FILE="${DIR}/images/nginx/aws.tfvars"
ENV_FILE="${DIR}/images/nginx/env.tfvars"

if [ ! -f "${AWS_FILE}" ]; then
  echo "File $AWS_FILE does not exist. Please create this file with keys access_key an secret_key"
  exit 1
fi

if [ ! -f "${ENV_FILE}" ]; then
  echo "File $ENV_FILE does not exist. Please create this file with keys docker_username and docker_password"
  exit 1
fi

# build local terraform image to build nginx
docker build -t build-nginx-terraform $DIR/images/nginx

# build nginx and publish docker images to quay.io.
# this can take up to two hours.
docker run --rm -it \
  --volume $DIR/images/nginx:/tf \
  -w /tf \
  -v ${AWS_FILE}:/root/aws.tfvars:ro \
  -v ${ENV_FILE}:/root/env.tfvars:ro \
  build-nginx-terraform

docker rmi -f build-nginx-terraform
